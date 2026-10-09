/*
===========================================================================

extended_population_test.go - the 91-140 mobs spawn behind the graft

Extended content (isro-live-2026, port-only, not native). Over the REAL
sealed projection: the native template knows none of the extended mobs,
the graft adds the live characterdata references, the seed area's
population composes into ordinary nests, and the production monster
registry spawns every band exemplar in the seed region. Flag off (no
graft) resolves nothing - the native world is unchanged.
===========================================================================
*/
package enterworld

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"opensro.online/server/internal/game/world/monster"
	"opensro.online/server/internal/game/world/simulation"
	"opensro.online/server/internal/game/world/worldarea"
	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
)

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

func TestExtendedMobsAreAbsentFromTheNativeTemplate(t *testing.T) {
	native := monster.LoadTemplate(gamedatatest.Paths(t).TextdataDir)
	for _, codename := range extendedSeedCodenames(t) {
		for id, ref := range native.Refs {
			if ref.Codename == codename {
				t.Fatalf("native template already knows %q (id %d); the graft must be the only way in", codename, id)
			}
		}
	}
}

func TestExtendedGraftSpawnsEveryBandExemplar(t *testing.T) {
	extended := extendedProjection(t)
	native := monster.LoadTemplate(gamedatatest.Paths(t).TextdataDir)
	grafted := monster.GraftRefs(native, monster.LoadTemplate(extended.TextdataDir).Refs)

	areas, err := worldarea.LoadAuthority(extended.AreasAuthorityDir)
	if err != nil {
		t.Fatal(err)
	}
	composed, err := appendAuthoredAreaPopulation(grafted, areas)
	if err != nil {
		t.Fatal(err)
	}
	seed := extendedSeedCodenames(t)
	if len(composed.Nests) != len(grafted.Nests)+len(seed) {
		t.Fatalf("composed nests = %d, want the grafted %d plus the %d seed rows", len(composed.Nests), len(grafted.Nests), len(seed))
	}

	registry := simulation.NewMonsterState(composed)
	registry.StartDivision("test")
	registry.AdvancePopulation(registry.CurrentTimeMillis())
	instances := registry.InstancesInRegions("test", []uint16{0x62a8})
	spawned := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		spawned[instance.Ref.Codename] = struct{}{}
	}
	for _, codename := range seed {
		if _, ok := spawned[codename]; !ok {
			t.Fatalf("seed mob %q did not spawn; spawned=%v", codename, spawned)
		}
	}
}

// extendedSeedCodenames reads the seed area's population from the built
// projection: the deterministic band exemplars the build recorded.
func extendedSeedCodenames(t *testing.T) []string {
	t.Helper()
	extended := extendedProjection(t)
	raw, err := os.ReadFile(filepath.Join(extended.AreasAuthorityDir, "areas", "catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
		Areas []struct {
			Slug       string `json:"slug"`
			Population []struct {
				Codename string `json:"codename"`
			} `json:"population"`
		} `json:"areas"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatal(err)
	}
	for _, area := range catalog.Areas {
		if area.Slug != "extended-content-lab" {
			continue
		}
		codenames := make([]string, 0, len(area.Population))
		for _, row := range area.Population {
			codenames = append(codenames, row.Codename)
		}
		if len(codenames) == 0 {
			t.Fatal("the seed area carries no population")
		}
		return codenames
	}
	t.Fatal("the seed area extended-content-lab is absent from the built projection")
	return nil
}
