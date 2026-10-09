/*
===========================================================================

skillstunguard_extended_test.go - the reactive stun link's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended SOULA_STUNLINK tiers past the cap pin the anonymous guard
link; every v1.150 row stays unpinned exactly as before - which is
what makes the runtime hook void-safe by construction.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestStunGuardPinsOnlyPastCap

The five tiers past the cap pin the guard with the st block's words
and the ceiling; the tiers at or under it stay out.
================
*/
func TestStunGuardPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "SOULA_STUNLINK") {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		link := row.TimedEffect.Link
		if got := row.TimedEffect.Pinned; got != (mastery >= extendedPastCapMastery) {
			t.Fatalf("%s (mastery %d): pinned %v", row.Codename, mastery, got)
		}
		if !row.TimedEffect.Pinned {
			continue
		}
		pinned++
		if !link.Present || !link.StunGuard.Present || link.Group != 0 ||
			link.StunGuard.DurationMs != 3000 && link.StunGuard.DurationMs != 5000 ||
			link.StunGuard.Chance != 35 && link.StunGuard.Chance != 50 ||
			link.StunGuard.Level < 10 || link.StunGuard.Level > 14 || link.StunGuard.CeilingLevel != 150 {
			t.Fatalf("%s: %+v", row.Codename, link)
		}
	}
	if pinned != 5 {
		t.Fatalf("pinned %d stun guards, want the five past-cap tiers", pinned)
	}
}

/*
================
TestStunGuardLeavesNativeRowsUnpinned

No shipped v1.150 row pins the guard link: anonymous (group 0) links
were refused by the native contract and the floor keeps them refused.
================
*/
func TestStunGuardLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.TimedEffect.Pinned && row.TimedEffect.Link.StunGuard.Present {
			t.Fatalf("%s pinned the stun guard", row.Codename)
		}
	}
}
