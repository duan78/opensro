/*
===========================================================================

skillfireshield_extended_test.go - the status shield's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended FIRE_SHIELD tiers past the cap pin their real resistance
block (installed by fileEffectResistance) with the bgra rider; the
family's cap-90 tiers stay unpinned exactly as before.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestFireShieldPinsOnlyPastCap

The three tiers past the cap pin the resistance block; the tiers at
or under it stay out.
================
*/
func TestFireShieldPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "FIRE_SHIELD") {
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
		real := row.TimedEffect.Real
		if real.Mask != 25145280 || real.Flat != 30 || real.Grade < 10 || real.Grade > 12 {
			t.Fatalf("%s: real %+v", row.Codename, real)
		}
	}
	if pinned != 3 {
		t.Fatalf("pinned %d status shields, want the three past-cap tiers", pinned)
	}
}

/*
================
TestFireShieldLeavesNativeRowsUnpinned

No shipped v1.150 row pins the shield through the bgra rider: the
family's own cap-90 tiers author the same shape and the floor keeps
them out.
================
*/
func TestFireShieldLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if strings.Contains(row.Codename, "FIRE_SHIELD") && row.TimedEffect.Pinned {
			t.Fatalf("%s pinned the status shield", row.Codename)
		}
	}
}
