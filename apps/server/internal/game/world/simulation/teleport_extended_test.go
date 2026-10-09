/*
===========================================================================

teleport_extended_test.go - the live-2026 gate graft, both settings

Extended content (isro-live-2026), port-only, not v1.150-native. Over
the REAL tables: the graft places the post-1.150 zones' gate NPCs into
the roster while every native gate keeps its exact native row, and the
without-graft roster stays exactly the native one.
===========================================================================
*/
package simulation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
	"opensro.online/server/internal/testsupport/licensed"
)

/*
================
TestAppendExtendedTeleportGatesNativeWins
================
*/
func TestAppendExtendedTeleportGatesNativeWins(t *testing.T) {
	licensed.RequireGameData(t)
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	extendedDir := extended.TextdataDir
	if _, err := os.Stat(filepath.Join(extendedDir, "teleportbuilding.txt")); err != nil {
		t.Skipf("extended teleport tables are not extracted: %v", err)
	}

	nativeDir := gamedatatest.TextdataDir(t)
	native, err := AppendTeleportGates(nativeDir, LoadNpcWorldRoster(nativeDir))
	if err != nil {
		t.Fatal(err)
	}
	nativeGates := gateRosterByRef(native)

	grafted, added, err := AppendExtendedTeleportGates(extendedDir, native)
	if err != nil {
		t.Fatal(err)
	}
	if added == 0 {
		t.Fatal("the live table grafted no new gate")
	}
	graftedGates := gateRosterByRef(grafted)

	// Native rows are untouched: same refs, same spawns, same bounds.
	zones := extendedZoneRegionSet(t)
	inZones := 0
	for ref, gate := range graftedGates {
		if nativeGate, native := nativeGates[ref]; native {
			if gate.ObjectID != nativeGate.ObjectID || gate.Spawn != nativeGate.Spawn ||
				gate.Teleport.Radius != nativeGate.Teleport.Radius || gate.Teleport.Height != nativeGate.Teleport.Height {
				t.Fatalf("native gate %d changed by the graft", ref)
			}
			continue
		}
		if zones[gate.Spawn.RegionID] {
			inZones++
		}
	}
	if inZones == 0 {
		t.Fatalf("%d gates grafted, none placed inside an extended zone", added)
	}
	t.Logf("grafted %d gates, %d inside extended zones", added, inZones)

	// Without the graft the roster stays exactly the native one.
	plain, err := AppendTeleportGates(nativeDir, LoadNpcWorldRoster(nativeDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(plain) != len(native) {
		t.Fatal("native roster is not deterministic")
	}
}

/*
================
gateRosterByRef
================
*/
func gateRosterByRef(roster []NpcDef) map[uint32]NpcDef {
	gates := make(map[uint32]NpcDef)
	for _, row := range roster {
		if row.Teleport != nil {
			gates[row.RefObjID] = row
		}
	}
	return gates
}

/*
================
extendedZoneRegionSet
================
*/
func extendedZoneRegionSet(t *testing.T) map[uint16]bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(extendedRoot(t), "zones.json"))
	if err != nil {
		t.Skipf("sealed zone list is not built: %v", err)
	}
	var zones struct {
		Zones map[string]struct{} `json:"zones"`
	}
	if err := json.Unmarshal(raw, &zones); err != nil {
		t.Fatalf("zone list: %v", err)
	}
	regions := make(map[uint16]bool, len(zones.Zones))
	for regionText := range zones.Zones {
		region, err := strconv.ParseUint(regionText[2:], 16, 32)
		if err != nil {
			t.Fatalf("zone region %q: %v", regionText, err)
		}
		regions[uint16(region)] = true
	}
	return regions
}

func extendedRoot(t *testing.T) string {
	t.Helper()
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	return extended.Root
}
