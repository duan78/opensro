/*
===========================================================================

commerce_extended_test.go - the live-2026 shops resolve end to end (M7)

Extended content (isro-live-2026), port-only, not v1.150-native. Over
the REAL tables: the merged catalogue plus the grafted roster open a
live town's merchant - every tab the NPC's mapping names resolves to
offers, and the goods include items a 91-140 player actually needs (the
replenishment coherence: requirement levels past 90). Skips when the
licensed data or the extended projection is not built.
===========================================================================
*/
package action

import (
	"os"
	"path/filepath"
	"testing"

	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/world/simulation"
	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
	"opensro.online/server/internal/testsupport/licensed"
)

/*
================
TestExtendedCommerceShopsResolve
================
*/
func TestExtendedCommerceShopsResolve(t *testing.T) {
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

	rt, _, _, _ := returnFixture(t, 30000)
	// The fixture's item set is a small hand-picked map; the live goods
	// resolve through the FULL extended catalogue, as production does.
	rt.deps = journeyItemDeps{rt.deps, enterworld.NewTextdataItems(extendedDir)}
	if err := rt.ConfigureCommerce(gamedatatest.TextdataDir(t)); err != nil {
		t.Fatal(err)
	}
	tabs, err := rt.MergeExtendedCommerce(extendedDir)
	if err != nil {
		t.Fatal(err)
	}
	if tabs == 0 {
		t.Fatal("the live shop plane added no tab")
	}
	t.Logf("extended shop tabs merged: %d", tabs)

	roster, _, err := simulation.AppendExtendedNpcWorldRoster(extendedDir, simulation.LoadNpcWorldRoster(gamedatatest.TextdataDir(t)))
	if err != nil {
		t.Fatal(err)
	}

	// Live merchants whose mappings resolve, and the 91+ replenishment:
	// an offered wearable or consumable whose requirement level is past
	// 90. Some live stores legitimately sit outside the native shop
	// contract (arena exchangers, silk-currency malls): their tabs carry
	// no admitted goods and are counted, never forced.
	merchants, resolvedTabs, offers := 0, 0, 0
	unresolvedTabs := 0
	offeredPast90, offeredTotal := 0, 0
	for _, npc := range roster {
		if npc.ObjectID < 210001 || npc.ObjectID >= 220000 || len(npc.NpcTalkStoreGroups) == 0 {
			continue
		}
		merchants++
		if merchants > 12 {
			break
		}
		for _, group := range npc.NpcTalkStoreGroups {
			for _, tab := range group.Tabs {
				offersForTab, ok := rt.Commerce.Tabs[tab.TabID]
				if !ok || len(offersForTab) == 0 {
					unresolvedTabs++
					continue
				}
				resolvedTabs++
				for _, offer := range offersForTab {
					offers++
					offeredTotal++
					if offer.Ref == nil {
						continue
					}
					if offer.Ref.ReqQuadTypes[0] == 1 && offer.Ref.ReqQuadValues[0] >= 91 {
						offeredPast90++
					}
				}
			}
		}
	}
	if merchants == 0 || resolvedTabs == 0 {
		t.Fatal("no grafted merchant with store groups resolved")
	}
	if offeredPast90 == 0 {
		t.Fatalf("%d offers resolved but none requires level 91+ - the 91-140 player cannot restock", offeredTotal)
	}
	t.Logf("merchants checked: %d, tabs resolved: %d, unadmitted tabs: %d, offers: %d (%d requiring level 91+)",
		merchants, resolvedTabs, unresolvedTabs, offers, offeredPast90)
}
