/*
===========================================================================

skillcripple_extended_test.go - the crippled-soul aura's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended REBIRTHA_SPECIAL_B_BUFF_02 (the skill 10276 the rmut revival
starts) pins its timed aura; the family's native BUFF tiers at or
under the cap stay unpinned exactly as before.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestCrippleAuraPinsOnlyPastCap

The one tier past the cap pins the crippled-soul aura - the attack
penalty, the incoming reduction, the recovery pair and the max-HP
raise; the tiers at or under it stay out.
================
*/
func TestCrippleAuraPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "REBIRTHA_SPECIAL") || !strings.Contains(row.Codename, "BUFF") {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if got := row.TimedEffect.Pinned; got != (mastery >= extendedPastCapMastery) {
			t.Fatalf("%s (mastery %d): pinned %v", row.Codename, mastery, got)
		}
		if !row.TimedEffect.Pinned {
			continue
		}
		pinned++
		effect := row.TimedEffect
		if !effect.Recovery.Present || effect.Recovery.HP != 50 || effect.Recovery.MP != 50 ||
			!effect.IncomingReduction || !effect.Attributes.MaxHP || effect.Attributes.HPPercent != 75 ||
			!effect.Attributes.DamagePenalty || effect.Attributes.PhysicalPenalty != 50 || effect.Attributes.MagicalPenalty != 50 {
			t.Fatalf("%s: recovery %+v attrs %+v inc %v", row.Codename, effect.Recovery, effect.Attributes, effect.IncomingReduction)
		}
	}
	if pinned != 1 {
		t.Fatalf("pinned %d crippled auras, want the one past-cap tier", pinned)
	}
}

/*
================
TestCrippleAuraLeavesNativeRowsUnpinned

No shipped v1.150 LEARNED row pins the aura's recovery pair: the
words are the live state's own and the floor keeps every native row
out. The item lane's own recovery programs (the potions, column 8 = 1)
are its contract and stay admitted.
================
*/
func TestCrippleAuraLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		// WATER_HARMONY is s16's own lane: its cgri recovery pairs were
		// admitted with their own void proof long before these words. The
		// claim here is only that no OTHER native row joins them.
		if strings.Contains(row.Codename, "WATER_HARMONY") {
			continue
		}
		if row.TimedEffect.Pinned && row.TimedEffect.Recovery.Present && !row.TimedEffect.ItemProgram {
			t.Fatalf("%s pinned a learned-skill recovery pair", row.Codename)
		}
	}
}
