/*
===========================================================================

skilldance_extended_test.go - the live "Dance with Music" toggle

Extended content (isro-live-2026), port-only, not v1.150-native (M8 s66,
owner order 2026-10-10): the live bard dance tiers author no efr and no
dru - their whole program is the toggle (ycdc/scls/reqc riders, the onff
rhythm, setv writes into the music-area slots). The family's native
ancestors admit through their efr aura; the live tiers pin on their own
evidence and run on the same pulse-owned party-aura engine.

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
TestDanceWithMusicPinsOnlyPastCap

The six live tiers parse the same facts; only the two past-cap tiers pin
(the ycdc/scls/reqc/setv riders carry the mastery floor).
================
*/
func TestDanceWithMusicPinsOnlyPastCap(t *testing.T) {
	extended := extendedProjection(t)
	live := NewTextdataSkills(extended.TextdataDir)
	if err := live.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range live.rows.values() {
		if !strings.HasPrefix(row.Codename, "SKILL_EU_BARD_DANCEA_WITH_MUSIC_") {
			continue
		}
		if !row.TimedEffect.Dance.Pinned {
			continue
		}
		pinned++
		if row.TimedEffect.Dance.RhythmMs != 5000 || row.TimedEffect.Dance.MusicArea != 5 {
			t.Fatalf("%s: dance = %+v", row.Codename, row.TimedEffect.Dance)
		}
		if row.TimedEffect.Dance.DanceRange != 3 && row.TimedEffect.Dance.DanceRange != 4 {
			t.Fatalf("%s: dance range = %d", row.Codename, row.TimedEffect.Dance.DanceRange)
		}
		if row.Aura.PulseMs != 5000 || row.Aura.PulseMP != 310 && row.Aura.PulseMP != 546 {
			t.Fatalf("%s: pulse = %d/%d", row.Codename, row.Aura.PulseMs, row.Aura.PulseMP)
		}
		if row.TimedEffect.Targeted || row.TimedEffect.Area.Present {
			t.Fatalf("%s: the toggle is targeted or area-shaped", row.Codename)
		}
	}
	if pinned != 2 {
		t.Fatalf("pinned %d dance tiers, want the two past-cap ones", pinned)
	}
}

/*
================
TestDanceHeaderWordAbsentFromNativePrograms

The ycdc word does not exist in any v1.150 row's program - the rider's
void premise, and with it the floor that keeps every native row off the
dance pin.
================
*/
func TestDanceHeaderWordAbsentFromNativePrograms(t *testing.T) {
	native := sharedShippedSkills(t)
	if err := native.Load(); err != nil {
		t.Fatal(err)
	}
	cells := readExtendedSkillCells(t, gamedatatest.TextdataDir(t))
	for _, row := range native.rows.values() {
		if row.TimedEffect.Dance.Pinned {
			t.Fatalf("%s: a native row pinned the dance", row.Codename)
		}
		program, err := CompileSkillProgram(cells[row.Codename])
		if err != nil {
			continue
		}
		for i := 0; i < program.Len(); i++ {
			if program.Instruction(i).Tag == tagDanceHeader {
				t.Fatalf("%s: the v1.150 program carries ycdc", row.Codename)
			}
		}
	}
}
