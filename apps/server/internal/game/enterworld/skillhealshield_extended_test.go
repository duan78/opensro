/*
===========================================================================

skillhealshield_extended_test.go - the party shield aura's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended RECOVERYA_HEALSHIELD tiers past the cap pin their kind-2
party area with the defp pair; every v1.150 row stays unpinned.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestHealShieldPinsOnlyPastCap

The three tiers past the cap pin the area with the defense pair; the
tiers at or under it stay out.
================
*/
func TestHealShieldPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "RECOVERYA_HEALSHIELD") {
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
		if !effect.Area.Present || effect.Area.Select != SelectParty && effect.Area.Select != SelectParty|SelectCaster || effect.Area.Radius != 300 ||
			!effect.Defense || effect.Physical == 0 || effect.Magical == 0 {
			t.Fatalf("%s: area %+v defense %v %d/%d", row.Codename, effect.Area, effect.Defense, effect.Physical, effect.Magical)
		}
	}
	if pinned != 3 {
		t.Fatalf("pinned %d party shields, want the three past-cap tiers", pinned)
	}
}

/*
================
TestHealShieldFloorIsTheVoid

The family's own tiers carry the proof: the tiers at or under the cap
author the same kind-2 defp shape and pin NOTHING (the first
TestHealShieldPinsOnlyPastCap asserts exactly that), so the floor is
the void. A broader native claim is not derivable post-parse - the
kind is not retained on the row, and native kind-1 party areas
(POISONA_GUARD's own line) have always pinned through the area path.
================
*/
func TestHealShieldFloorIsTheVoid(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	low := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "RECOVERYA_HEALSHIELD") {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if mastery < extendedPastCapMastery && row.TimedEffect.Pinned {
			t.Fatalf("%s (mastery %d) pinned below the floor", row.Codename, mastery)
		}
		if mastery < extendedPastCapMastery {
			low++
		}
	}
	if low == 0 {
		t.Fatal("the family lost its below-cap tiers - the void's premise changed")
	}
}
