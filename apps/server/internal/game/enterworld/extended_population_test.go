/*
===========================================================================

extended_population_test.go - the 91-140 mobs spawn in their real zones

Extended content (isro-live-2026, port-only, not native). Over the REAL
sealed projection: the zones the build derived from the live client's
npcpos (regions the v1.150 world never served) compose into ordinary
authored areas, every trajectory band 91..140 is placed, the production
monster registry spawns each zone's population in its real region, and
none of the placed mobs exists natively - the flag-off world is
unchanged.
===========================================================================
*/
package enterworld

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"opensro.online/server/internal/game/world/monster"
	"opensro.online/server/internal/game/world/simulation"
	"opensro.online/server/internal/game/world/worldarea"
	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
)

type extendedZoneArea struct {
	Slug       string `json:"slug"`
	RegionID   uint16 `json:"regionId"`
	Population []struct {
		Codename string `json:"codename"`
	} `json:"population"`
}

type extendedZoneList struct {
	RegionCount int `json:"regionCount"`
	AnchorCount int `json:"anchorCount"`
	Zones       map[string]struct {
		Anchors int   `json:"anchors"`
		Bands   []int `json:"bands"`
	} `json:"zones"`
}

func extendedProjection(t *testing.T) gamedata.Extended {
	t.Helper()
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		if errors.Is(err, gamedata.ErrExtendedNotBuilt) {
			t.Skipf("extended projection is not built: %v", err)
		}
		t.Fatal(err)
	}
	if _, err := os.Stat(extended.TextdataDir); err != nil {
		t.Skipf("extended projection textdata is unavailable: %v", err)
	}
	return extended
}

// extendedZones reads the built projection's authored areas and zone list.
func extendedZones(t *testing.T) ([]extendedZoneArea, *extendedZoneList) {
	t.Helper()
	extended := extendedProjection(t)
	raw, err := os.ReadFile(filepath.Join(extended.AreasAuthorityDir, "areas", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
		Areas []extendedZoneArea `json:"areas"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatal(err)
	}
	zonesRaw, err := os.ReadFile(filepath.Join(extended.Root, "zones.json"))
	if err != nil {
		t.Fatal(err)
	}
	var zones extendedZoneList
	if err := json.Unmarshal(zonesRaw, &zones); err != nil {
		t.Fatal(err)
	}
	if len(catalog.Areas) != zones.RegionCount {
		t.Fatalf("areas catalog has %d areas, the zone list records %d", len(catalog.Areas), zones.RegionCount)
	}
	return catalog.Areas, &zones
}

func TestExtendedZonesCoverEveryTrajectoryBand(t *testing.T) {
	_, zones := extendedZones(t)
	bands := map[int]bool{}
	for _, zone := range zones.Zones {
		for _, band := range zone.Bands {
			if band >= 91 && band <= 140 {
				bands[band] = true
			}
		}
	}
	for band := 91; band <= 140; band += 10 {
		if !bands[band] {
			t.Fatalf("trajectory band %d has no real zone", band)
		}
	}
}

func TestExtendedZoneMobsAreAbsentFromTheNativeTemplate(t *testing.T) {
	areas, _ := extendedZones(t)
	native := monster.LoadTemplate(gamedatatest.Paths(t).TextdataDir)
	nativeCodenames := make(map[string]struct{}, len(native.Refs))
	for _, ref := range native.Refs {
		nativeCodenames[ref.Codename] = struct{}{}
	}
	checked := 0
	for _, area := range areas {
		for _, row := range area.Population {
			if _, known := nativeCodenames[row.Codename]; known {
				t.Fatalf("zone %s places %q, which the native template already ships", area.Slug, row.Codename)
			}
			checked++
		}
	}
	if checked < 100 {
		t.Fatalf("only %d placed rows read; the real-zone derivation is missing", checked)
	}
}

func TestExtendedZonesSpawnInTheirRealRegions(t *testing.T) {
	extended := extendedProjection(t)
	areas, zones := extendedZones(t)
	native := monster.LoadTemplate(gamedatatest.Paths(t).TextdataDir)
	grafted := monster.GraftRefs(native, monster.LoadMonsterRefs(extended.TextdataDir))

	authored, err := worldarea.LoadAuthority(extended.AreasAuthorityDir)
	if err != nil {
		t.Fatal(err)
	}
	composed, err := appendAuthoredAreaPopulation(grafted, authored)
	if err != nil {
		t.Fatal(err)
	}
	registry := simulation.NewMonsterState(composed)
	registry.StartDivision("test")
	registry.AdvancePopulation(registry.CurrentTimeMillis())

	// Every zone's whole population spawns in its own region; the band
	// spread comes from the zone list, the row sum from the registry.
	spawnedInBand := map[int]bool{}
	spawnedRows := 0
	for _, area := range areas {
		instances := registry.InstancesInRegions("test", []uint16{area.RegionID})
		present := make(map[string]struct{}, len(instances))
		for _, instance := range instances {
			present[instance.Ref.Codename] = struct{}{}
		}
		for _, row := range area.Population {
			if _, ok := present[row.Codename]; !ok {
				t.Fatalf("zone %s (region 0x%04X): %q did not spawn", area.Slug, area.RegionID, row.Codename)
			}
			spawnedRows++
		}
		zone, ok := zones.Zones[fmt.Sprintf("0x%04x", area.RegionID)]
		if !ok {
			t.Fatalf("zone list lacks region 0x%04X", area.RegionID)
		}
		for _, band := range zone.Bands {
			if band >= 91 && band <= 140 {
				spawnedInBand[band] = true
			}
		}
	}
	for band := 91; band <= 140; band += 10 {
		if !spawnedInBand[band] {
			t.Fatalf("no spawned zone covers band %d", band)
		}
	}
	if spawnedRows < 1000 {
		t.Fatalf("only %d zone rows spawned; the real-zone population is missing", spawnedRows)
	}
}
