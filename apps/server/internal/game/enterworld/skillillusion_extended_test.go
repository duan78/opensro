/*
===========================================================================

skillillusion_extended_test.go - the illusion's floor

Extended content (isro-live-2026), port-only, not v1.150-native. The
CONFUSIONA_ILLUSION family exists ONLY in the live catalogue (zero
v1.150 rows carry msch mode 3 - measured), so the pin needs no
mastery floor: no native row can ever reach the parser.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestIllusionPinsAllLiveTiers

Every tier of the live-only family pins (the parse is the family's
own - no native carrier exists to protect).
================
*/
func TestIllusionPinsAllLiveTiers(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	pinned := 0
	for _, row := range skills.rows.values() {
		if !strings.Contains(row.Codename, "CONFUSIONA_ILLUSION") {
			continue
		}
		if !row.Illusion.Pinned || row.Illusion.MaxLevel == 0 {
			t.Fatalf("%s: %+v", row.Codename, row.Illusion)
		}
		pinned++
	}
	if pinned != 11 {
		t.Fatalf("pinned %d illusions, want the family's eleven live tiers", pinned)
	}
}

/*
================
TestIllusionIsTheOnlyMsch3Family

Measured: the CONFUSIONA_ILLUSION family is the ONLY author of msch
mode 3 (the earlier probe found zero native carriers; the full-suite
loader can resolve the shared cache against the extended projection
once an extended test has run, so the assertion here is the
ordering-independent one - whichever catalogue the shared loader
holds, only this family carries the word).
================
*/
func TestIllusionIsTheOnlyMsch3Family(t *testing.T) {
	native := sharedShippedSkills(t)
	for _, row := range native.rows.values() {
		if row.CastGate.MschPresent && row.CastGate.MschMode == 3 &&
			!strings.Contains(row.Codename, "CONFUSIONA_ILLUSION") {
			t.Fatalf("%s carries msch mode 3 - the word leaked beyond the family", row.Codename)
		}
		if row.Illusion.Pinned && !strings.Contains(row.Codename, "CONFUSIONA_ILLUSION") {
			t.Fatalf("%s pinned the illusion", row.Codename)
		}
	}
}
