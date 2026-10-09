/*
===========================================================================

service_npc_extended_test.go - the live towns' service NPCs, both settings

Extended content (isro-live-2026), port-only, not v1.150-native. Over
the REAL tables: the graft places the live towns' merchants into the
roster at their own npcpos anchors (native codenames keep their native
rows), the shopkeepers among them carry their store groups, and the
ungrafted roster stays exactly native. Skips when the licensed data or
the extended projection is not built.
===========================================================================
*/
package simulation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"opensro.online/server/internal/domain"
	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
	"opensro.online/server/internal/testsupport/licensed"
)

/*
================
TestAppendExtendedNpcWorldRosterPlacesTheLiveTowns
================
*/
func TestAppendExtendedNpcWorldRosterPlacesTheLiveTowns(t *testing.T) {
	licensed.RequireGameData(t)
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	extendedDir := extended.TextdataDir
	if _, err := os.Stat(filepath.Join(extendedDir, "refshoptab.txt")); err != nil {
		t.Skipf("extended commerce tables are not extracted: %v", err)
	}

	nativeDir := gamedatatest.TextdataDir(t)
	native := LoadNpcWorldRoster(nativeDir)
	grafter, added, err := AppendExtendedNpcWorldRoster(extendedDir, native)
	if err != nil {
		t.Fatal(err)
	}
	if added < 200 {
		t.Fatalf("the live towns grafted only %d service NPCs; the extraction measured 277", added)
	}
	t.Logf("service NPCs grafted: %d (native roster %d -> %d)", added, len(native), len(grafter))

	// Native rows are untouched (the graft only appends; a native codename
	// may legally appear on several rows) and the additions live in their
	// own band.
	if len(grafter) < len(native) {
		t.Fatal("the graft shrank the roster")
	}
	for index, row := range native {
		if grafter[index].ObjectID != row.ObjectID || grafter[index].Codename != row.Codename {
			t.Fatalf("native row %d (%s) changed by the graft", index, row.Codename)
		}
	}
	nativeByCode := map[string]bool{}
	for _, row := range native {
		nativeByCode[row.Codename] = true
	}
	shopkeepers, inNewRegions := 0, 0
	nativeRegions := map[uint16]bool{}
	for _, row := range native {
		nativeRegions[row.Spawn.RegionID] = true
	}
	zones := extendedTownZoneSet(t, extended.Root)
	seenIDs := map[uint32]bool{}
	for _, row := range grafter {
		if seenIDs[row.ObjectID] {
			t.Fatalf("duplicate ObjectID %d after the graft", row.ObjectID)
		}
		seenIDs[row.ObjectID] = true
		if nativeByCode[row.Codename] {
			continue
		}
		if row.ObjectID < domain.NPCGIDBase+10000+1 || row.ObjectID >= domain.NPCGIDBase+20000 {
			t.Fatalf("extended NPC %s outside the dedicated band: %d", row.Codename, row.ObjectID)
		}
		if !nativeRegions[row.Spawn.RegionID] {
			inNewRegions++
		}
		if len(row.NpcTalkStoreGroups) > 0 {
			shopkeepers++
		}
	}
	if inNewRegions == 0 {
		t.Fatal("no grafted service NPC stands in a region the native world never served")
	}
	if shopkeepers == 0 {
		t.Fatal("no grafted service NPC carries store groups - the shops cannot open")
	}
	t.Logf("grafted in never-native regions: %d; with store groups: %d; in trajectory zones: %d",
		inNewRegions, shopkeepers, func() int {
			n := 0
			for _, row := range grafter {
				if !nativeByCode[row.Codename] && zones[row.Spawn.RegionID] {
					n++
				}
			}
			return n
		}())

	// Without the graft the roster stays exactly the native one.
	plain := LoadNpcWorldRoster(nativeDir)
	if len(plain) != len(native) {
		t.Fatal("native roster is not deterministic")
	}
}

/*
================
extendedTownZoneSet

The trajectory zones (where a 91-140 player actually hunts - the service
NPCs standing there are the ones the journey can walk to).
================
*/
func extendedTownZoneSet(t *testing.T, extendedRoot string) map[uint16]bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(extendedRoot, "zones.json"))
	if err != nil {
		t.Skipf("sealed zone list is not built: %v", err)
	}
	var zones struct {
		Zones map[string]struct{} `json:"zones"`
	}
	if err := json.Unmarshal(raw, &zones); err != nil {
		t.Fatalf("zone list: %v", err)
	}
	out := make(map[uint16]bool, len(zones.Zones))
	for text := range zones.Zones {
		region, err := strconv.ParseUint(text[2:], 16, 32)
		if err != nil {
			t.Fatalf("zone region %q: %v", text, err)
		}
		out[uint16(region)] = true
	}
	return out
}
