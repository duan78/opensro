/*
===========================================================================

extended_mirror_test.go - the real extended mirror, end to end

Extended content (isro-live-2026), port-only, not v1.150-native. Against
the REAL sealed projection (when it is built): the world lane's movement
mirror chains behind a native authority, every mirrored bundle loads at
boot-time preload, and one real 2026 zone per trajectory band answers
spawn-available. Skips when the projection or the mirror is not built -
the flag-off and mid-migration postures keep today's answers.
===========================================================================
*/
package movement

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"opensro.online/server/internal/gamedata"
)

type extendedMirrorZones struct {
	Zones map[string]struct {
		Bands []int `json:"bands"`
	} `json:"zones"`
}

/*
================
TestExtendedMirrorWalkableThroughTheChain
================
*/
func TestExtendedMirrorWalkableThroughTheChain(t *testing.T) {
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	movementRoot := filepath.Join(extended.Root, "movement")
	mirrorCatalog, err := os.ReadFile(filepath.Join(movementRoot, "catalog.json"))
	if err != nil {
		t.Skipf("extended movement mirror is not built: %v", err)
	}
	var mirror struct {
		RegionsByID map[string][]struct {
			BundlePath string `json:"bundlePath"`
		} `json:"regionsById"`
	}
	if err := json.Unmarshal(mirrorCatalog, &mirror); err != nil {
		t.Fatalf("movement mirror catalog: %v", err)
	}
	if len(mirror.RegionsByID) == 0 {
		t.Fatal("movement mirror catalog names no regions")
	}

	// The native half is a synthetic root whose catalog names no regions
	// (production's native tree always has one): the chain under test is
	// WHICH tree answers, and an empty native catalog proves the mirror
	// answers alone.
	nativeRoot := t.TempDir()
	writeTestAsset(t, nativeRoot, "assets/world/world-region-catalog.json", `{"format": "sro-world-region-catalog", "version": 1, "regionsById": {}}`)
	water := NewWaterValidator(nativeRoot)
	if err := water.SetExtendedAuthorityRoot(movementRoot); err != nil {
		t.Fatalf("chain the extended movement mirror: %v", err)
	}

	// The boot-time preload walks every mirrored bundle; a broken one is a
	// failed build, not a runtime degrade.
	if _, err := water.PreloadOutdoorNavigation(); err != nil {
		t.Fatalf("preload the chained mirror: %v", err)
	}
	if regions := water.ExtendedMovementRegions(); regions != len(mirror.RegionsByID) {
		t.Fatalf("ExtendedMovementRegions = %d, want %d", regions, len(mirror.RegionsByID))
	}

	// One real 2026 zone per trajectory band answers spawn-available
	// through the chain.
	zonesBytes, err := os.ReadFile(filepath.Join(extended.Root, "zones.json"))
	if err != nil {
		t.Fatalf("read the sealed zone list: %v", err)
	}
	var zones extendedMirrorZones
	if err := json.Unmarshal(zonesBytes, &zones); err != nil {
		t.Fatalf("parse the sealed zone list: %v", err)
	}
	availableByBand := make(map[int]bool)
	for regionText, zone := range zones.Zones {
		parsed, err := strconv.ParseUint(regionText[2:], 16, 32)
		if err != nil {
			t.Fatalf("zone region %q: %v", regionText, err)
		}
		regionID := uint16(parsed)
		if !water.SpawnRegionAvailable(regionID) {
			continue
		}
		for _, band := range zone.Bands {
			if band >= 91 && band <= 140 {
				availableByBand[band] = true
			}
		}
	}
	for band := 91; band <= 140; band += 10 {
		if !availableByBand[band] {
			t.Fatalf("no spawn-available extended zone in band %d", band)
		}
		t.Logf("band %d has spawn-available extended zones through the chained mirror", band)
	}
}
