/*
===========================================================================

skillmindmaintain_extended_test.go - the mind-maintenance passive's
carried words

Extended content (isro-live-2026), port-only, not v1.150-native. The
live MINDP_MAINTAIN tiers pin through the passive lane: the setv is
DTDR (ParameterDotDuration - always a known key, EXECUTED through the
periodic producer's duration reader), hwir carried unexecuted on the
DSCR precedent (no clean mapping onto the 0..5 Hwan gauge, none
invented - the recorded boundary).

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestMindMaintainPinsWithCarriedWords

The live tiers pin with the DoT-duration slot filed and hwir carried.
================
*/
func TestMindMaintainPinsWithCarriedWords(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "MINDP_MAINTAIN") {
			continue
		}
		if !row.PassiveParameters.Pinned {
			continue
		}
		pinned++
		if !row.PassiveParameters.Mask.Has(ParameterDotDuration) || !row.PassiveParameters.HwirPresent {
			t.Fatalf("%s: the carried words are missing", row.Codename)
		}
	}
	if pinned < 2 {
		t.Fatalf("pinned %d mind-maintenance passives, want the live tiers", pinned)
	}
}

/*
================
TestMindMaintainWordsAbsentFromNativeData

hwir does not exist in the v1.150 data: the carried word's
void premise, frozen.
================
*/
func TestMindMaintainWordsAbsentFromNativeData(t *testing.T) {
	native := sharedShippedSkills(t)
	// Order-independent (the full suite can resolve the shared loader
	// against the extended projection once an extended test has run):
	// the claim is that only the MINDP_MAINTAIN family authors hwir.
	for _, row := range native.rows.values() {
		if row.PassiveParameters.HwirPresent &&
			!strings.Contains(row.Codename, "MINDP_MAINTAIN") {
			t.Fatalf("%s carries hwir - the word leaked beyond the family", row.Codename)
		}
	}
}
