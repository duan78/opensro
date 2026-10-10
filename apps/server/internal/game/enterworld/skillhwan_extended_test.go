/*
===========================================================================

skillhwan_extended_test.go - the berserk-duration aura's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
two live "Aura of Blood" tiers past the cap pin the hwdu word with the
party area; no v1.150 row carries the word (the rider list froze its
absence when the compiler admitted it).

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
TestHwanAuraPinsOnlyPastCap

The two live tiers pin with the measured seconds and the party area.
================
*/
func TestHwanAuraPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "SPELLP_PARTY_HWAN") {
			continue
		}
		if !row.TimedEffect.Pinned {
			t.Fatalf("%s: the aura does not pin", row.Codename)
		}
		pinned++
		if row.TimedEffect.HwanDurationMs != 30000 && row.TimedEffect.HwanDurationMs != 35000 ||
			!row.TimedEffect.Area.Present || row.TimedEffect.Area.Radius != 300 ||
			row.TimedEffect.Area.MaxTargets != 8 || row.TimedEffect.Area.Select != 5 {
			t.Fatalf("%s: hwan=%d area=%+v", row.Codename, row.TimedEffect.HwanDurationMs, row.TimedEffect.Area)
		}
	}
	if pinned != 2 {
		t.Fatalf("pinned %d auras of blood, want the two live tiers", pinned)
	}
}

/*
================
TestHwanWordAbsentFromNativePrograms

The hwdu word does not exist in any v1.150 row's program: the rider's
void premise, frozen.
================
*/
func TestHwanWordAbsentFromNativePrograms(t *testing.T) {
	native := sharedShippedSkills(t)
	cells := readExtendedSkillCells(t, gamedatatest.TextdataDir(t))
	for _, row := range native.rows.values() {
		program, err := CompileSkillProgram(cells[row.Codename])
		if err != nil {
			continue
		}
		for i := 0; i < program.Len(); i++ {
			if program.Instruction(i).Tag == 0x68776475 {
				t.Fatalf("%s carries the hwdu word - the void premise is gone", row.Codename)
			}
		}
	}
}
