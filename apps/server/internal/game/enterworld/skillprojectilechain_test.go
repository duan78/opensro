/*
===========================================================================

skillprojectilechain_test.go - handler-1 chains on every concrete weapon

SkillAction_Projectile (5857B0) links stages for ANY launcher, and the
v1.150 table itself proves it beyond the bow/crossbow pairs: the sword's
GEOMGI C/C2/D/D2 flying blades are handler-1 chains on swords. The port's
earlier bow/crossbow-only restriction refused those native rows and, with
them, the live 91-140 families that extend the same shape - the spear
throw (SPEAR_SHOOT D/E), the later GEOMGI E/F series and SPECIAL E (M8
s64). The families must compile offense plans; the any-weapon monster
marker {255,255} stays refused by the ammunition gate that owns it.

===========================================================================
*/

package enterworld

import (
	"strings"
	"testing"
)

/*
================
TestProjectileChainsCompileOnEveryConcreteWeapon

Every past-90 chain root of the three live families compiles an offense
plan whose stages include its companion strike - the 51 rows the old
bow/crossbow-only envelope refused.
================
*/
func TestProjectileChainsCompileOnEveryConcreteWeapon(t *testing.T) {
	extended := extendedProjection(t)
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{
		"SKILL_CH_SPEAR_SHOOT_D_":   10,
		"SKILL_CH_SPEAR_SHOOT_E_":   11,
		"SKILL_CH_SWORD_GEOMGI_D_":  3,
		"SKILL_CH_SWORD_GEOMGI_E_":  7,
		"SKILL_CH_SWORD_GEOMGI_F_":  9,
		"SKILL_CH_SWORD_SPECIAL_E_": 11,
	}
	seen := map[string]int{}
	for _, row := range skills.rows.values() {
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if mastery < 91 {
			continue
		}
		for prefix := range want {
			if !strings.HasPrefix(row.Codename, prefix) || row.ChainSub || row.ChainNext == 0 {
				continue
			}
			plan := skills.ExecutionPlan(row.ID)
			if plan.Kind() != SkillExecutionOffense {
				t.Fatalf("%s: plan kind %d, want offense", row.Codename, plan.Kind())
			}
			if plan.Len() < 2 {
				t.Fatalf("%s: plan holds %d stages, want the companion strike", row.Codename, plan.Len())
			}
			seen[prefix]++
		}
	}
	for prefix, count := range want {
		if seen[prefix] != count {
			t.Errorf("%s*: %d chain roots compiled, want %d", prefix, seen[prefix], count)
		}
	}
}

/*
================
TestNativeSwordProjectileChainsCompile

The v1.150 sword GEOMGI C/D chains compile as offense - the native
evidence that settled the envelope - while the any-weapon monster marker
stays refused by the ammunition gate.
================
*/
func TestNativeSwordProjectileChainsCompile(t *testing.T) {
	native := sharedShippedSkills(t)
	if err := native.Load(); err != nil {
		t.Fatal(err)
	}
	compiled := 0
	for _, row := range native.rows.values() {
		if !strings.HasPrefix(row.Codename, "SKILL_CH_SWORD_GEOMGI_") {
			continue
		}
		if row.ChainSub || row.ChainNext == 0 {
			continue
		}
		if native.ExecutionPlan(row.ID).Kind() == SkillExecutionOffense {
			compiled++
		}
	}
	if compiled == 0 {
		t.Fatal("no native GEOMGI chain root compiles - the handler-1 envelope regressed to launcher-only")
	}
	monster, ok := native.byCodename["MSKILL_RM_TAHOMET_ATTACK01_03"]
	if !ok {
		t.Fatal("the native handler-1 monster row is gone from the table")
	}
	if kind := native.ExecutionPlan(monster).Kind(); kind != SkillExecutionUnsupported {
		t.Fatalf("the {255,255} monster row compiles kind %d - the ammunition gate no longer owns it", kind)
	}
}
