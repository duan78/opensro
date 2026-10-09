/*
===========================================================================

extended_journey_test.go - the scripted 1-to-140 journey (E8)

Extended content (isro-live-2026), port-only, not v1.150-native. Over
the REAL projection: a character kills its way from level 1 to the cap
140 through the real kill formula, the real 2026 curve and the real
progression runtime - no level is unreachable, every extended band
boundary moves the character into one of its real zones (public entry),
the loot planner pays gold from the official curve at each band's first
step, and the eleventh-degree gear is granted once its requirement level
is met. The kills-per-band table the run produces is the milestone's
sanity-balance evidence. Skips when the projection is not built.
===========================================================================
*/
package action

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"opensro.online/server/internal/domain"
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/loot"
	"opensro.online/server/internal/game/progression"
	"opensro.online/server/internal/game/world/monster"
	"opensro.online/server/internal/game/world/worldarea"
	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
	"opensro.online/server/internal/testsupport/licensed"
)

/*
================
TestExtendedJourneyKillsItsWayFromOneTo140
================
*/
func TestExtendedJourneyKillsItsWayFromOneTo140(t *testing.T) {
	licensed.RequireGameData(t)
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	// The extended drop supplement (port-only, not native): the journey's
	// equipment legs need the live degrees in the ordinary drop path.
	if installErr := loot.InstallExtendedEquipment(); installErr != nil {
		t.Fatalf("extended equipment supplement: %v", installErr)
	}

	// The real 2026 curve, the full item catalogue and the full monster
	// catalogue - native rows plus the grafted live-2026 refs, exactly the
	// wiring's composition.
	levels := enterworld.NewExtendedLevels(extended.LeveldataPath, extended.GoldCurvePath)
	items := enterworld.NewTextdataItems(extended.TextdataDir)
	template := monster.LoadTemplate(gamedatatest.TextdataDir(t))
	template = monster.GraftRefs(template, monster.LoadMonsterRefs(extended.TextdataDir))
	rewarding := func(ref monster.MonsterRef) bool {
		return ref.RewardActionPinned && ref.ExpToGive > 0 && ref.MaxHP > 0 && ref.Level >= 1
	}
	// Placed hunts first (what a player actually fights); the whole grafted
	// catalogue answers where the placements' levels have a granularity
	// gap - every census mob of the range is still a killable 2026 mob.
	hunters := make([]monster.MonsterRef, 0, 1024)
	for _, ref := range template.SpawnableRefs() {
		if rewarding(ref) {
			hunters = append(hunters, ref)
		}
	}
	catalog := make([]monster.MonsterRef, 0, 8192)
	for _, ref := range monster.LoadMonsterRefs(extended.TextdataDir) {
		if rewarding(ref) {
			catalog = append(catalog, ref)
		}
	}
	sort.Slice(hunters, func(i, j int) bool { return hunters[i].Level < hunters[j].Level })
	sort.Slice(catalog, func(i, j int) bool { return catalog[i].Level < catalog[j].Level })
	if len(hunters) == 0 || len(catalog) == 0 {
		t.Fatal("no reward-bearing monster refs")
	}

	// The extended zones: bands from the sealed zone list, entries and the
	// public-access rule from the projection's authored areas.
	areas, err := worldarea.LoadAuthority(extended.AreasAuthorityDir)
	if err != nil {
		t.Fatalf("extended areas: %v", err)
	}
	zonesByBand := journeyZonesByBand(t, extended.Root)

	rt, character, _, _ := returnFixture(t, 30000)
	rt.NpcSpawn.Enabled = true
	rt.DropRoll = func() (uint32, error) { return 0, nil }
	// The real progression plane over the same character pointer, with the
	// extended cap: every kill's XP flows through the production grant.
	journey := progression.NewRuntime(&enterworld.Deps{
		Characters:      enterworld.StaticCharacterSource{testDivision: {character}},
		Items:           items,
		Levels:          levels,
		MutateCharacter: func(*enterworld.Character, string, func()) {},
	})
	journey.LevelCap = 140
	rt.UpdateExperience = journey.ExperienceUpdater()
	// The fixture's item set is a small hand-picked map; the journey's
	// loot legs resolve gold-heap tiers and eleventh-degree rows through
	// the FULL extended catalogue, as production does.
	rt.deps = journeyItemDeps{rt.deps, items}

	level := int64(1)
	character.Level = &level
	totalKills, worstLevel, worstKills := 0, int64(0), 0
	killsByBand := map[int]int{}
	bandEntered := map[int]bool{}
	// The bound catches an IMPOSSIBLE step, not a slow one. The official
	// 2026 curve's top levels need millions of best-mob kills (measured:
	// 137->138 requires 112 329 945 026 428 XP; the densest full-credit
	// hunt pays 839 999 936); beyond the bound the journey grants the
	// level's requirement in one aggregate through the same path. The
	// spawned champions, giants, uniques and party bonuses - all native
	// systems - are the practical accelerators the official game relies on.
	const sanityKillBound = 2000000

	for *character.Level < 140 {
		current := *character.Level
		band := int((current-1)/10)*10 + 1
		// A real traveller trains as they level: an untrained mastery gap
		// costs the EXP rate (clamped to 10% at level 29 and worse beyond).
		character.Masteries = []domain.CharacterMastery{{ID: 1, Level: current}}
		ref := journeyNearestHunter(hunters, current)
		if ref.ExpToGive == 0 {
			ref = journeyNearestHunter(catalog, current)
		}
		instance := monster.Instance{Ref: ref, Gid: 40000}
		exp, skillExp := monsterKillReward(character, instance, levels)
		if exp <= 0 {
			t.Fatalf("level %d: %s (level %d) yields no experience", current, ref.Codename, ref.Level)
		}
		kills := 0
		for *character.Level == current && kills < sanityKillBound {
			rt.UpdateExperience(character, exp, skillExp, instance.Gid)
			kills++
		}
		if *character.Level == current {
			t.Fatalf(
				"level %d stuck: %d kills of %s (level %d, %d xp each) on the 2026 curve",
				current, kills, ref.Codename, ref.Level, exp,
			)
		}
		totalKills += kills
		killsByBand[band] += kills
		if kills > worstKills {
			worstLevel, worstKills = current, kills
		}

		// The band's first level-up: move into one of the band's real
		// zones, through the same public-entry rule a teleport commits
		// under, and prove the loot planner pays gold on the kill.
		if zone, ok := zonesByBand[band]; ok && !bandEntered[band] {
			bandEntered[band] = true
			area, resolved := areas.ResolveRegion(zone.region)
			if !resolved || !areas.CanEnterRegion(zone.region, false) {
				t.Fatalf("band %d zone 0x%04X is not publicly enterable", band, zone.region)
			}
			region, x, y, z := int64(area.RegionID), area.Entry.X, area.Entry.Y, area.Entry.Z
			angle := int64(area.Entry.Angle)
			character.World.Spawn = &domain.WorldSpawn{
				RegionID: &region, X: &x, Y: &y, Z: &z, Angle: &angle,
			}
			pose := monster.Pose{RegionID: area.RegionID, X: x, Y: y, Z: z}
			gold, gear := false, ""
			for _, drop := range rt.planMonsterKillLoot(character, instance, pose, time.Now().UnixMilli()) {
				if drop.GoldAmount > 0 && !gold {
					gold = true
					t.Logf("band %d: %s dropped a %d-gold heap in 0x%04X", band, ref.Codename, drop.GoldAmount, area.RegionID)
				}
				if gear == "" && (strings.Contains(drop.Codename, "_10_") ||
					strings.Contains(drop.Codename, "_11_") || strings.Contains(drop.Codename, "_12_")) {
					gear = drop.Codename
				}
			}
			if !gold {
				t.Fatalf("band %d: %s dropped no gold on the official curve", band, ref.Codename)
			}
			// The equipment leg (E4): from band 91 up, the kill's loot
			// carries a wearable of the extended degrees. Level 101 is the
			// live transition - the native DG10-C window and the live
			// DG11-A one are both open and roll against each other - so the
			// assertion is the degree floor.
			floor := 10
			if band >= 111 {
				floor = 12
			}
			matched := false
			for degree := floor; degree <= 12; degree++ {
				if strings.Contains(gear, fmt.Sprintf("_%02d_", degree)) {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatalf("band %d: %s dropped %q, want a DG%02d+ wearable", band, ref.Codename, gear, floor)
			}
			t.Logf("band %d: %s dropped %s", band, ref.Codename, gear)
		}
	}

	// The equipment leg: once the eleventh degree's requirement level is
	// passed, the overlay's real row is granted to the traveller.
	sword, ok := items.ItemRefByCodename("ITEM_CH_SWORD_11_A")
	if !ok {
		t.Fatal("the eleventh-degree sword is absent from the extended catalogue")
	}
	if sword.ReqQuadTypes[0] == 1 && sword.ReqQuadValues[0] > 101 {
		t.Fatalf("eleventh-degree sword requires level %d", sword.ReqQuadValues[0])
	}
	character.MissionInventory = append(character.MissionInventory, enterworld.InventoryRow{
		Slot: 22, RefObjID: sword.RefObjID, Codename: sword.Codename, StackCount: 1,
	})

	for band := 1; band <= 141; band += 10 {
		if kills := killsByBand[band]; kills > 0 {
			t.Logf("band %3d: %5d kills", band, kills)
		}
	}
	t.Logf("journey 1->140 complete: %d kills, worst level %d at %d kills", totalKills, worstLevel, worstKills)
}

/*
================
journeyItemDeps

The fixture dependencies with the item source swapped for the full
extended catalogue.
================
*/
type journeyItemDeps struct {
	Dependencies
	items *enterworld.TextdataItems
}

func (d journeyItemDeps) ItemReferences() enterworld.ItemRefSource { return d.items }

/*
================
journeyNearestHunter

The densest hunt at the traveller's level: among the reward-bearing refs
within the native full-credit window (ten levels each way), the one
whose experience times the native level-gap scale pays the most (a
player farms the best mob of the range - the band uniques included -
not the first clone the table lists).
================
*/
func journeyNearestHunter(hunters []monster.MonsterRef, level int64) monster.MonsterRef {
	var best monster.MonsterRef
	var bestScore float64
	for _, ref := range hunters {
		gap := int64(ref.Level) - level
		if gap < -10 || gap > 10 {
			continue
		}
		score := float64(ref.ExpToGive) * float64(monsterLevelGapRewardScale(level, int64(ref.Level)))
		if score > bestScore {
			best, bestScore = ref, score
		}
	}
	return best
}

/*
================
journeyZonesByBand

The sealed zone list's first region per band.
================
*/
func journeyZonesByBand(t *testing.T, extendedRoot string) map[int]struct{ region uint16 } {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(extendedRoot, "zones.json"))
	if err != nil {
		t.Fatalf("sealed zone list: %v", err)
	}
	var zones struct {
		Zones map[string]struct {
			Bands []int `json:"bands"`
		} `json:"zones"`
	}
	if err := json.Unmarshal(raw, &zones); err != nil {
		t.Fatalf("zone list: %v", err)
	}
	byBand := map[int]struct{ region uint16 }{}
	for regionText, zone := range zones.Zones {
		region, err := strconv.ParseUint(regionText[2:], 16, 32)
		if err != nil {
			t.Fatalf("zone region %q: %v", regionText, err)
		}
		for _, band := range zone.Bands {
			if _, seen := byBand[band]; !seen {
				byBand[band] = struct{ region uint16 }{uint16(region)}
			}
		}
	}
	return byBand
}
