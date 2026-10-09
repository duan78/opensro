/*
===========================================================================

skilldisperse_extended_test.go - the damage-divide link's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended DAMAGE_DIVIDE tiers past the cap pin the lkdd disperse link;
every v1.150 row stays unpinned exactly as before.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestDispersePinsOnlyPastCap

The four tiers past the cap (SWORD, SPEAR, BOW, GUARDA) pin the
disperse link with the 60 percent share; the tiers at or under it
stay out.
================
*/
func TestDispersePinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "DAMAGE_DIVIDE") && !strings.Contains(row.Codename, "GUARDA_DIVIDE") {
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
		// The weapon lines share 60 percent; the Warrior's own guard 75.
		if !link.Present || !link.Disperse || link.DispersePercent != 60 && link.DispersePercent != 75 || link.Group != 0 {
			t.Fatalf("%s: %+v", row.Codename, link)
		}
	}
	if pinned != 4 {
		t.Fatalf("pinned %d disperse links, want the four past-cap tiers", pinned)
	}
}

/*
================
TestDisperseLeavesNativeRowsUnpinned

No shipped v1.150 row pins the disperse link: the party-without-ally
target shape and the anonymous group-0 link are the live rows' own,
both under the mastery floor.
================
*/
func TestDisperseLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.TimedEffect.Pinned && row.TimedEffect.Link.Disperse {
			t.Fatalf("%s pinned the disperse link", row.Codename)
		}
	}
}
