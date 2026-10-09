/*
===========================================================================

skillstance_extended_test.go - the sword-and-shield stance's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended SHIELDPD tiers past the cap pin the spda trade-off; the
family's cap-90 tiers - the twenty-two v1.150 rows that author the
same word - stay unpinned exactly as before.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestShieldStancePinsOnlyPastCap

The thirteen tiers past the cap (D_05..D_06, E_01..E_06, F_01..F_05)
pin the stance; the tiers at or under it stay out.
================
*/
func TestShieldStancePinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "SWORD_SHIELDPD") {
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
			stance := row.TimedEffect.Stance
			if !stance.Present || stance.AttackPercent < 76 || stance.AttackPercent > 110 ||
				stance.DefenseCutPercent < 23 || stance.DefenseCutPercent > 76 {
				t.Fatalf("%s: %+v", row.Codename, stance)
			}
		}
	}
	if pinned != 13 {
		t.Fatalf("pinned %d stances, want the thirteen past-cap tiers", pinned)
	}
}

/*
================
TestShieldStanceLeavesNativeRowsUnpinned

No shipped v1.150 row pins the stance: the mastery floor is the void
proof, and the native catalogue holds no row at or past it - not even
the family's own cap-90 tiers that author the same spda word.
================
*/
func TestShieldStanceLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.TimedEffect.Pinned && row.TimedEffect.Stance.Present {
			t.Fatalf("%s pinned the stance", row.Codename)
		}
	}
}
