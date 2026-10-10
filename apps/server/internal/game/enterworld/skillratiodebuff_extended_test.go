/*
===========================================================================

skillratiodebuff_extended_test.go - the timed ratio cut's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended catalogue's timed WATER_CANCEL tiers past the cap pin the
ratio debuff; the family's cap-90 tiers - and every other v1.150 row -
stay unpinned exactly as before. The pin is engine-side only until its
cast executor ships, so these rows are NOT yet in the admission union.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestRatioDebuffPinsOnlyPastCap

The eight timed tiers past the cap (A_10..A_12, B_07..B_10,
CANCEL2_A_01) pin; the tiers at or under it stay out.
================
*/
func TestRatioDebuffPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !row.RatioDebuff.Pinned {
			continue
		}
		pinned++
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if mastery < extendedPastCapMastery || !row.TargetRequired || row.RatioDebuff.DurationMs != 10000 || row.RatioDebuff.EvasionDown > 200 || row.RatioDebuff.HitDown > 200 {
			t.Fatalf("%s (mastery %d): %+v", row.Codename, mastery, row.RatioDebuff)
		}
		// The A and B lines carry one cut word; CANCEL2_A carries both.
		if row.RatioDebuff.EvasionDown == 0 && row.RatioDebuff.HitDown == 0 {
			t.Fatalf("%s: no cut word, %+v", row.Codename, row.RatioDebuff)
		}
		if !strings.Contains(row.Codename, "WATER_CANCEL") || row.RatioDebuff.ThreatFlat == 0 {
			t.Fatalf("%s: unexpected shape %+v", row.Codename, row.RatioDebuff)
		}
	}
	if pinned != 9 {
		t.Fatalf("pinned %d ratio debuffs, want the nine timed tiers", pinned)
	}
}

/*
================
TestRatioDebuffLeavesNativeRowsUnpinned

No shipped v1.150 row pins the timed ratio cut: the mastery floor is
the void proof, and the native catalogue holds no row at or past it.
================
*/
func TestRatioDebuffLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.RatioDebuff.Pinned {
			t.Fatalf("%s pinned the timed ratio cut", row.Codename)
		}
	}
}
