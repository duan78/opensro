/*
===========================================================================

skillpoisonfield_extended_test.go - the poison field's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended POISONA_FIELD tiers past the cap pin the field; the family's
five cap-90 tiers - and every other v1.150 row - stay unpinned.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestPoisonFieldPinsOnlyPastCap

The five tiers past the cap pin the field with the measured words;
the tiers at or under it stay out.
================
*/
func TestPoisonFieldPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "POISONA_FIELD") {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if got := row.PoisonField.Pinned; got != (mastery >= extendedPastCapMastery) {
			t.Fatalf("%s (mastery %d): pinned %v", row.Codename, mastery, got)
		}
		if !row.PoisonField.Pinned {
			continue
		}
		pinned++
		field := row.PoisonField
		if field.DurationMs != 30000 || field.ScanMs != 3000 || field.Radius == 0 ||
			field.PoisonMs < 108 || field.PoisonMs > 288 || field.Chance != 80 || field.Damage < 120 || field.Damage > 1964 {
			t.Fatalf("%s: %+v", row.Codename, field)
		}
	}
	if pinned != 5 {
		t.Fatalf("pinned %d poison fields, want the five past-cap tiers", pinned)
	}
}

/*
================
TestPoisonFieldLeavesNativeRowsUnpinned

No shipped v1.150 row pins the field: the family's own cap-90 tiers
author the same shape and the floor keeps them out.
================
*/
func TestPoisonFieldLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.PoisonField.Pinned {
			t.Fatalf("%s pinned the poison field", row.Codename)
		}
	}
}
