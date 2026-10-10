/*
===========================================================================

skillchargearea_extended_test.go - the charge-area tolerance's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
FRENZYA B tiers past the cap carry tel3 AND an efr together; the void
proof measured on 2026-10-10 is that the pair is absent from the
v1.150 data entirely (zero native rows), so the tolerance admits only
live rows.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"

	"opensro.online/server/internal/testsupport/gamedatatest"
)

/*
================
TestChargeAreaPinsOnlyPastCap

The B tiers past the cap admit through the tolerance with their area;
the A tiers (single-target) were always admitted; every at-or-under
tier stays on the native contract.
================
*/
func TestChargeAreaPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pastArea := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "FRENZYA_TOUNT_SPRINT") {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if !row.OffensiveStagePinned {
			t.Fatalf("%s (mastery %d): the rush attack is not staged", row.Codename, mastery)
		}
		if mastery >= extendedPastCapMastery && row.OffensiveArea.Radius != 0 {
			pastArea++
		}
	}
	if pastArea != 5 {
		t.Fatalf("%d rush tiers carry an area, want the five B tiers", pastArea)
	}
}

/*
================
TestChargeAreaPairAbsentFromNativeData

The tel3+efr pair does not exist in the v1.150 data at all: the
tolerance's void premise, frozen.
================
*/
func TestChargeAreaPairAbsentFromNativeData(t *testing.T) {
	native := sharedShippedSkills(t)
	cells := readExtendedSkillCells(t, gamedatatest.TextdataDir(t))
	for _, row := range native.rows.values() {
		fields := cells[row.Codename]
		if len(fields) == 0 {
			continue
		}
		program, err := CompileSkillProgram(fields)
		if err != nil {
			continue
		}
		hasCharge, hasEfr := false, false
		for i := 0; i < program.Len(); i++ {
			switch program.Instruction(i).Tag {
			case 0x74656c33:
				hasCharge = true
			case 0x656672:
				hasEfr = true
			}
		}
		if hasCharge && hasEfr {
			t.Fatalf("%s carries the charge-area pair - the void premise is gone", row.Codename)
		}
	}
}
