/*
===========================================================================

skillthreatdecrease_extended_test.go - the caster-centred cut's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended catalogue's CONFUSIONA_AGGROLOW B_03..B_11 pin the untargeted
hostility cut; the same family's cap-90 tiers - and every other v1.150
row - stay unpinned exactly as the targeted contract refuses them.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestUntargetedThreatDecreasePinsOnlyPastCap

The extended Warlock tiers past the cap pin Decrease with the
caster-centred area; the tiers at or under it (native's own Mirage and
Phantasma shape) stay out.
================
*/
func TestUntargetedThreatDecreasePinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "CONFUSIONA_AGGROLOW") {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if row.Threat.Decrease != (mastery >= extendedPastCapMastery) {
			t.Fatalf("%s (mastery %d): decrease %v", row.Codename, mastery, row.Threat.Decrease)
		}
		if row.Threat.Decrease {
			pinned++
			if row.Threat.Area.Shape != 1 || row.Threat.Area.Radius != 300 ||
				row.Threat.Area.MaxTargets != 8 || row.Threat.Area.Select != 16 ||
				row.Threat.DecreaseFlat == 0 || row.TargetRequired {
				t.Fatalf("%s: %+v", row.Codename, row.Threat)
			}
		}
	}
	if pinned != 9 {
		t.Fatalf("pinned %d untargeted cuts, want the nine B_03..B_11 tiers", pinned)
	}
}

/*
================
TestUntargetedThreatDecreaseLeavesNativeRowsUnpinned

No shipped v1.150 row pins the untargeted cut: the mastery floor is the
void proof, and the native catalogue holds no row at or past it. The
Warlock's own cap-90 Mirage and Phantasma tiers author the same shape
and must stay refused exactly as before.
================
*/
func TestUntargetedThreatDecreaseLeavesNativeRowsUnpinned(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.Threat.Decrease && !row.TargetRequired {
			t.Fatalf("%s pinned the untargeted cut", row.Codename)
		}
	}
}
