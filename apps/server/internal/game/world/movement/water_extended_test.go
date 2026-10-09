/*
===========================================================================

water_extended_test.go - the extended movement chain's two settings

Extended content (isro-live-2026), port-only, not v1.150-native. The chain
must resolve a region only the extended mirror carries, keep the native
answer wherever both trees carry the region, and change nothing when it is
not installed (the flag-off boot).
===========================================================================
*/
package movement

import (
	"fmt"
	"strconv"
	"testing"
)

// syntheticExtendedMovementRoot builds an authority-layout mirror (the
// shape SetExtendedAuthorityRoot installs) for one live-2026 zone region,
// 0x4939 (sector 57/73): block (0,0) is type-0 water at surface 50, and a
// second region 0x6B4F - also present in the NATIVE synthetic root with
// water - is deliberately DRY here, so a precedence test can tell whose
// answer won.
func syntheticExtendedMovementRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTestAsset(t, root, "catalog.json", `{
		"format": "sro-server-movement-catalog",
		"version": 1,
		"regionsById": {
			"0x4939": [{
				"id": "0x4939",
				"area": "outdoor",
				"seedRegionId": "0x4939",
				"worldRegionsPath": "extended/world-regions.json",
				"bundlePath": "extended/regions/region-4939.json",
				"source": "extended-outdoor-live-2026"
			}],
			"0x6b4f": [{
				"id": "0x6b4f",
				"area": "outdoor",
				"seedRegionId": "0x6b4f",
				"worldRegionsPath": "extended/world-regions.json",
				"bundlePath": "extended/regions/region-6b4f.json",
				"source": "extended-outdoor-live-2026"
			}]
		}
	}`)
	writeTestAsset(t, root, "extended/world-regions.json", `{
		"format": "sro-server-world-region-index",
		"version": 1,
		"regionSize": 1920,
		"seedRegionId": "0x0000",
		"regions": [
			{"id": "0x4939", "bundlePath": "extended/regions/region-4939.json"},
			{"id": "0x6b4f", "bundlePath": "extended/regions/region-6b4f.json"}
		]
	}`)
	regionBundle := func(sectorX, sectorY int, water string) string {
		return `{
			"source": {"sectorId": "0x` + fmt.Sprintf("%04x", sectorY<<8|sectorX) + `", "sectorX": ` + strconv.Itoa(sectorX) + `, "sectorY": ` + strconv.Itoa(sectorY) + `},
			"terrain": {"sectors": [{
				"sectorX": ` + strconv.Itoa(sectorX) + `, "sectorY": ` + strconv.Itoa(sectorY) + `,
				"blocks": [{"blockX": 0, "blockZ": 0, "water": ` + water + `}]
			}]},
			"navmesh": {
				"regionSize": 1920, "tileSize": 20, "tilesPerAxis": 2,
				"regions": [{
					"dx": "0", "dz": "0",
					"blockedTiles": "AAAAAA==",
					"tileCellIds": "AAAAAAAAAAAAAAAAAAAAAA==",
					"cells": {"count": "1"}
				}]
			}
		}`
	}
	writeTestAsset(t, root, "extended/regions/region-4939.json",
		regionBundle(57, 73, `{"type": 0, "waveType": 1, "height": 50}`))
	writeTestAsset(t, root, "extended/regions/region-6b4f.json",
		regionBundle(79, 107, `{"type": 1, "waveType": 0, "height": -20}`))
	return root
}

/*
================
TestExtendedChainResolvesLive2026Region

Flag on, chain installed: a destination the native tree has no surface for
resolves through the mirror - deep water in 0x4939 is refused, exactly as
the same shape is natively in 0x6B4F.
================
*/
func TestExtendedChainResolvesLive2026Region(t *testing.T) {
	validator := NewWaterValidator(syntheticWaterRoot(t))
	if err := validator.SetExtendedAuthorityRoot(syntheticExtendedMovementRoot(t)); err != nil {
		t.Fatalf("SetExtendedAuthorityRoot: %v", err)
	}

	// A seabed click 40u under the mirrored water surface: refused.
	if refusal := validator.ValidateMovement(moveTo(0x4939, 100, 10, 100)); refusal == nil {
		t.Fatal("extended region 0x4939: expected deep-water refusal through the chain")
	}
	// The native region keeps answering: same click in 0x6B4F is refused by
	// the NATIVE surface even though the mirror carries a DRY 0x6B4F.
	if refusal := validator.ValidateMovement(moveTo(0x6B4F, 100, 10, 100)); refusal == nil {
		t.Fatal("native region 0x6B4F: expected the native water refusal to win over the dry mirror")
	}
	// Spawn availability follows the same chain.
	if !validator.SpawnRegionAvailable(0x4939) {
		t.Fatal("extended region 0x4939: expected spawn availability through the chain")
	}
	if regions := validator.ExtendedMovementRegions(); regions != 2 {
		t.Fatalf("ExtendedMovementRegions = %d, want 2", regions)
	}
	if _, err := validator.PreloadOutdoorNavigation(); err != nil {
		t.Fatalf("PreloadOutdoorNavigation with the chain: %v", err)
	}
}

/*
================
TestExtendedChainAbsentKeepsNativeAnswers

Flag off (no chain installed): the mirrored region has no surface, so the
same deep-water click is ACCEPTED (degrade-to-accept is the reference
policy for uncovered regions) and the extended count stays zero.
================
*/
func TestExtendedChainAbsentKeepsNativeAnswers(t *testing.T) {
	validator := NewWaterValidator(syntheticWaterRoot(t))

	if refusal := validator.ValidateMovement(moveTo(0x4939, 100, 10, 100)); refusal != nil {
		t.Fatalf("unchained 0x4939: unexpected refusal %v", refusal)
	}
	if validator.SpawnRegionAvailable(0x4939) {
		t.Fatal("unchained 0x4939: spawn region must not be available natively")
	}
	if regions := validator.ExtendedMovementRegions(); regions != 0 {
		t.Fatalf("ExtendedMovementRegions = %d, want 0 without a chain", regions)
	}
}

/*
================
TestExtendedChainRejectsUnreadableMirror

A named mirror without a catalog is a configuration error, never a silent
degrade.
================
*/
func TestExtendedChainRejectsUnreadableMirror(t *testing.T) {
	validator := NewWaterValidator(syntheticWaterRoot(t))
	if err := validator.SetExtendedAuthorityRoot(t.TempDir()); err == nil {
		t.Fatal("SetExtendedAuthorityRoot on an empty root: expected an error")
	}
}
