/*
===========================================================================

portal_extended_test.go - the live-2026 teleport graft, both settings

Extended content (isro-live-2026), port-only, not v1.150-native. Over
the REAL tables: the merge adds the post-1.150 destinations while every
native identity keeps its row, and a character actually travels through
a grafted gate INTO an extended zone and BACK - E7's round trip at the
action layer. Skips when the licensed data or the extended projection is
not built (the flag-off world keeps exactly the native catalogue).
===========================================================================
*/
package action

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/world/simulation"
	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
	"opensro.online/server/internal/testsupport/licensed"
)

/*
================
TestExtendedPortalMergeNativeWinsAndAddsZones
================
*/
func TestExtendedPortalMergeNativeWinsAndAddsZones(t *testing.T) {
	licensed.RequireGameData(t)
	extended := extendedPortalDir(t)

	rt, _, _, _ := returnFixture(t, 30000)
	if err := rt.ConfigurePortals(gamedatatest.TextdataDir(t)); err != nil {
		t.Fatal(err)
	}
	nativeDestinations := len(rt.portals.destinations)
	nativeLinks := len(rt.portals.links)
	for id, destination := range rt.portals.destinations {
		if destination.id != id {
			t.Fatalf("native destination %d keyed as %d", destination.id, id)
		}
	}

	stats, err := rt.MergeExtendedPortalDir(extended)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("merge: %+v", stats)
	if stats.DestinationsAdded == 0 || stats.LinksAdded == 0 {
		t.Fatalf("the live plane added nothing: %+v", stats)
	}
	if len(rt.portals.destinations) != nativeDestinations+stats.DestinationsAdded {
		t.Fatal("destination accounting drifted")
	}
	if len(rt.portals.links) != nativeLinks+stats.LinksAdded {
		t.Fatal("link accounting drifted")
	}

	// The graft must place destinations inside the extended zones, and for
	// every such reachable destination the return link must exist too.
	zones := extendedZoneRegions(t)
	intoZones, roundTrips := 0, 0
	for _, link := range rt.portals.links {
		target := rt.portals.destinations[link.target]
		if !zones[target.spawn.RegionID] {
			continue
		}
		intoZones++
		if _, back := rt.portals.links[[2]uint32{link.target, link.source}]; back {
			roundTrips++
		}
	}
	if intoZones == 0 {
		t.Fatal("no link lands inside an extended zone")
	}
	if roundTrips == 0 {
		t.Fatalf("%d links into extended zones, none with a return link", intoZones)
	}
	t.Logf("extended zones reachable: %d links, %d with return", intoZones, roundTrips)
}

/*
================
TestExtendedPortalRoundTripThroughTheGates

A character at a native gate travels into an extended zone and back
through HandlePortal - the full admission, fee and world-transfer path
E7 rides in production.
================
*/
func TestExtendedPortalRoundTripThroughTheGates(t *testing.T) {
	licensed.RequireGameData(t)
	extended := extendedPortalDir(t)

	rt, c, _, _ := returnFixture(t, 30000)
	rt.NpcSpawn.Enabled = true
	// The post-1.150 zones' links carry level conditions the fixture's
	// level-20 character cannot pass (refusal 0x15): a cap-140 traveller is
	// the character the extended world exists for.
	level := int64(140)
	c.Level = &level
	if err := rt.ConfigurePortals(gamedatatest.TextdataDir(t)); err != nil {
		t.Fatal(err)
	}
	if _, err := rt.MergeExtendedPortalDir(extended); err != nil {
		t.Fatal(err)
	}

	// One native gate whose catalogue reaches an extended zone whose OWN
	// gate carries the return link (both ends must be gates: the travel
	// starts and ends at an NPC).
	zones := extendedZoneRegions(t)
	var sourceID, targetID uint32
	for _, link := range rt.portals.links {
		source, target := rt.portals.destinations[link.source], rt.portals.destinations[link.target]
		if zones[target.spawn.RegionID] && !zones[source.spawn.RegionID] &&
			target.spawn.RegionID&0x8000 == 0 && source.ref != 0 && target.ref != 0 {
			if _, back := rt.portals.links[[2]uint32{link.target, link.source}]; back {
				sourceID, targetID = link.source, link.target
				break
			}
		}
	}
	if sourceID == 0 {
		t.Fatal("no native-gate link into an extended-zone gate with a return")
	}
	source := rt.portals.destinations[sourceID]
	target := rt.portals.destinations[targetID]

	rt.NpcRoster = []simulation.NpcDef{
		{ObjectID: 2001, RefObjID: source.ref, Codename: "NPC_NATIVE_GATE", TalkFlags: 2},
		{ObjectID: 2002, RefObjID: target.ref, Codename: "NPC_EXTENDED_GATE", TalkFlags: 2},
	}

	// There and back again: both travels must move the authoritative
	// world spawn into the expected region.
	body := func(gid uint32, target uint32) []byte {
		return wire.NewWriter(9).U32(gid).U8(2).U32(target).Payload()
	}
	rt.Selected.Set(testDivision, c.Name, 2001)
	out := rt.HandlePortal(testDivision, c, body(2001, targetID))
	if len(out.Frames) == 0 || out.Frames[0].Opcode != enterworld.OpcodeResetClient {
		t.Fatalf("travel into the extended zone refused: %+v", out.Frames)
	}
	if got := c.World.Spawn.RegionID; got == nil || !zones[uint16(*got)] {
		t.Fatalf("travel landed outside the extended zones: %v", got)
	}
	rt.Selected.Set(testDivision, c.Name, 2002)
	out = rt.HandlePortal(testDivision, c, body(2002, sourceID))
	if len(out.Frames) == 0 || out.Frames[0].Opcode != enterworld.OpcodeResetClient {
		t.Fatalf("return travel refused: %+v", out.Frames)
	}
	if got := c.World.Spawn.RegionID; got == nil || uint16(*got) != source.spawn.RegionID {
		t.Fatalf("return landed in %v, want the native gate region 0x%04X", got, source.spawn.RegionID)
	}
}

/*
================
extendedPortalDir

The built extended projection's textdata tree, or a skip.
================
*/
func extendedPortalDir(t *testing.T) string {
	t.Helper()
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	if _, err := os.Stat(filepath.Join(extended.TextdataDir, "teleportbuilding.txt")); err != nil {
		t.Skipf("extended teleport tables are not extracted: %v", err)
	}
	return extended.TextdataDir
}

/*
================
extendedZoneRegions

The sealed zone list's region ids.
================
*/
func extendedZoneRegions(t *testing.T) map[uint16]bool {
	t.Helper()
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(extended.Root, "zones.json"))
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
