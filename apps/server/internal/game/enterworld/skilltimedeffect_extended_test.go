/*
===========================================================================

skilltimedeffect_extended_test.go - the timed evasion block's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended catalogue's JIPJUNG tiers past the cap pin the er self buff;
the family's cap-90 tiers - and every other learned-skill row of the
v1.150 catalogue - stay unpinned exactly as before.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestTimedEvasionPinsOnlyPastCap

The extended Lightning tiers past the cap pin a self evasion buff; the
tiers at or under it (the v1.150 catalogue's own JIPJUNG rows, which
author the same word) stay out.
================
*/
func TestTimedEvasionPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "LIGHTNING_JIPJUNG") {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if got := row.TimedEffect.Pinned; got != (mastery >= extendedPastCapMastery) {
			t.Fatalf("%s (mastery %d): pinned %v", row.Codename, mastery, got)
		}
		if row.TimedEffect.Pinned {
			pinned++
			if !row.TimedEffect.Evasion.Present || row.TimedEffect.Evasion.Percent != 0 ||
				row.TimedEffect.Evasion.Flat == 0 || row.TimedEffect.ItemProgram {
				t.Fatalf("%s: %+v", row.Codename, row.TimedEffect.Evasion)
			}
		}
	}
	if pinned != 9 {
		t.Fatalf("pinned %d evasion buffs, want the nine D_04..E_03 tiers", pinned)
	}
}

/*
================
TestTimedEvasionLeavesNativeLearnedRowsUnpinned

No shipped v1.150 LEARNED skill row pins the er block: the mastery
floor is the void proof, and the native catalogue holds no row at or
past it. The item lane's er programs (the evasion scrolls and socket
stones, column 8 = 1) are its own contract and stay admitted.
================
*/
func TestTimedEvasionLeavesNativeLearnedRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.TimedEffect.Evasion.Present && !row.TimedEffect.ItemProgram {
			t.Fatalf("%s pinned the learned-skill evasion block", row.Codename)
		}
		if strings.Contains(row.Codename, "LIGHTNING_JIPJUNG") && row.TimedEffect.Pinned {
			t.Fatalf("%s pinned a timed effect through er", row.Codename)
		}
	}
}
