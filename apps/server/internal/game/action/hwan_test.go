/*
===========================================================================

hwan_test.go - the berserk-duration extension's runtime behavior

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s63): the aura extends every party recipient's RUNNING berserk mode by
the word's seconds; a recipient outside the mode is untouched.

===========================================================================
*/

package action

import (
	"testing"

	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
)

func hwanSkill100() enterworld.SkillRow {
	return enterworld.SkillRow{ID: 100, Group: 9, EffectDurationMs: 300000, RequiredWeaponKinds: [2]uint8{0xff, 0xff},
		Consumption: enterworld.SkillConsumption{Pinned: true}, TimingPinned: true,
		TimedEffect: enterworld.SkillTimedEffect{Pinned: true, HwanDurationMs: 30000,
			Area: enterworld.SkillRecipientArea{Present: true, Radius: 300, MaxTargets: 8, Select: 5}}}
}

func TestHwanAuraExtendsRunningBerserkOnly(t *testing.T) {
	rt, character := newActiveEffectTestRuntime(t, staticSkillSource{})
	inMode := testCharacter()
	inMode.Name = "inmode"
	inMode.ID = 51
	idle := testCharacter()
	idle.Name = "idle"
	idle.ID = 52
	character.Skills = []uint32{100}
	character.CurrentMP = testInt64(100000)
	rt.deps = &enterworld.Deps{Characters: enterworld.StaticCharacterSource{testDivision: {character, inMode, idle}},
		Items: testItems(), Skills: staticSkillSource{100: hwanSkill100()}}
	gid := enterworld.ObjectIDForCharacter(character)
	gids := []uint32{gid, enterworld.ObjectIDForCharacter(inMode), enterworld.ObjectIDForCharacter(idle)}
	rt.RewardParties = func(string) []RewardParty { return []RewardParty{{Members: gids}} }
	now := rt.Now().UnixMilli()
	// One recipient in the running mode, one outside it.
	inMode.NativeBodyStatus = 1
	inMode.BerserkUntilMs = now + 60000
	idle.NativeBodyStatus = 0

	skill, _ := rt.deps.SkillData().SkillByID(100)
	out := rt.acceptHwanAura(testDivision, character, character, wire.SkillAction{ActionId: 100}, skill, now)
	if out.DiagnosticRefusal != "" {
		t.Fatalf("the aura was refused: %q", out.DiagnosticRefusal)
	}
	if inMode.BerserkUntilMs != now+60000+30000 {
		t.Fatalf("the running mode was not extended: %d", inMode.BerserkUntilMs)
	}
	if idle.BerserkUntilMs != 0 {
		t.Fatalf("a recipient outside the mode was touched: %d", idle.BerserkUntilMs)
	}
}
