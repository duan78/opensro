/*
===========================================================================

skilllinkredirect_extended_test.go - the redirection link's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended GUARDA tiers past the cap pin the lkdr redirection link; the
family's cap-90 tiers - and every other v1.150 row - stay unpinned
exactly as before, which is also what makes the runtime hook void-safe
by construction.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestLinkedRedirectPinsOnlyPastCap

The nine tiers past the cap (PHYSICAL A_12..A_15, MAGIC A_10..A_14)
pin the redirection link with the lane mask and the scaling share; the
tiers at or under it stay out.
================
*/
func TestLinkedRedirectPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		// The BLOCK variant is another family (the tail's map); the
		// redirection link is the PHYSICAL and MAGIC _A_ lines' own.
		if strings.Contains(row.Codename, "BLOCK") ||
			!strings.Contains(row.Codename, "GUARDA_PHYSICAL") && !strings.Contains(row.Codename, "GUARDA_MAGIC") {
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
		if !link.Present || !link.Redirect || link.RedirectPercent < 33 || link.RedirectPercent > 75 ||
			link.RedirectMask != 4 && link.RedirectMask != 8 || link.Mana || link.Threat {
			t.Fatalf("%s: %+v", row.Codename, link)
		}
	}
	if pinned != 9 {
		t.Fatalf("pinned %d redirection links, want the nine past-cap tiers", pinned)
	}
}

/*
================
TestLinkedRedirectLeavesNativeRowsUnpinned

No shipped v1.150 row pins the redirection link: the mastery floor is
the void proof (no native row reaches it), and by construction no
native game can ever hold one - the strike-path hook stays inert for
every native strike.
================
*/
func TestLinkedRedirectLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.TimedEffect.Pinned && row.TimedEffect.Link.Redirect {
			t.Fatalf("%s pinned the redirection link", row.Codename)
		}
	}
}
