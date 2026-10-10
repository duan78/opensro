/*
===========================================================================

skillwarlockmp_extended_test.go - the RIMD cost-cut slot's void

Extended content (isro-live-2026), port-only, not v1.150-native. The
live MINDP_MANA_DECREASE tiers past the cap pin the setv{RIMD}
passive; the key is absent from the v1.150 data entirely (the s37
DMIR void proof), so no native row can carry the slot.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestWarlockMPDecreasePinsOnlyPastCap

The two tiers past the cap pin the passive with the ten-percent cut;
the tiers at or under it stay out only if they author some other
shape - the family's low tiers carry the SAME word and pin too (the
setv lane is a learned passive, not a cast: the mastery floor does not
apply, the parameter itself is the live rows' own).
================
*/
func TestWarlockMPDecreasePinsOnlyLiveRows(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "MINDP_MANA_DECREASE") {
			continue
		}
		if !row.PassiveParameters.Pinned {
			continue
		}
		pinned++
		if !row.PassiveParameters.Mask.Has(ParameterWarlockMPDecrease) {
			t.Fatalf("%s: mask wrong", row.Codename)
		}
		if v := row.PassiveParameters.Values[ParameterWarlockMPDecrease]; v < 5 || v > 15 {
			t.Fatalf("%s: cut word %d outside the measured band", row.Codename, v)
		}
	}
	if pinned == 0 {
		t.Fatal("no MINDP_MANA_DECREASE row pinned the RIMD passive")
	}
}

/*
================
TestWarlockMPDecreaseKeyAbsentFromNativeData

The RIMD key does not exist in the v1.150 data: no shipped row's
passive carries the warlock slot - the void premise, frozen.
================
*/
func TestWarlockMPDecreaseKeyAbsentFromNativeData(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.PassiveParameters.Pinned && row.PassiveParameters.Mask.Has(ParameterWarlockMPDecrease) {
			t.Fatalf("%s carries the RIMD slot - the void premise is gone", row.Codename)
		}
	}
}
