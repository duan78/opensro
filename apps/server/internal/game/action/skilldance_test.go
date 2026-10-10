/*
===========================================================================

skilldance_test.go - the live dance toggle on the party-aura engine

Extended content (isro-live-2026), port-only, not v1.150-native (M8 s66,
owner order 2026-10-10): "Dance with Music" installs as the pulse-owned
toggle - the same engine every native dance runs on - and one pulse
charges the authored onff MP.

===========================================================================
*/

package action

import (
	"testing"

	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
)

func danceSkill100() enterworld.SkillRow {
	return enterworld.SkillRow{ID: 100, Group: 9, EffectDurationMs: 0, RequiredWeaponKinds: [2]uint8{0xff, 0xff},
		Consumption: enterworld.SkillConsumption{MP: 10, Pinned: true}, TimingPinned: true,
		ActionCastingTimePinned: true, ActionDurationPinned: true,
		Aura: enterworld.SkillAura{PulseMs: 5000, PulseMP: 5},
		TimedEffect: enterworld.SkillTimedEffect{Pinned: true,
			Dance: enterworld.SkillDance{Pinned: true, RhythmMs: 5000, MusicArea: 5, DanceRange: 3}}}
}

/*
================
TestDanceToggleInstallsAndPulsesOnThePartyAuraEngine

The toggle installs on the party-aura engine at the authored rhythm with
no authored radius (the dance's effect is the caster's own widened music,
carried on the row), and one pulse charges the authored MP.
================
*/
func TestDanceToggleInstallsAndPulsesOnThePartyAuraEngine(t *testing.T) {
	rt, character := newActiveEffectTestRuntime(t, staticSkillSource{})
	character.Skills = []uint32{100}
	character.CurrentMP = testInt64(100000)
	rt.deps = &enterworld.Deps{Characters: enterworld.StaticCharacterSource{testDivision: {character}},
		Items: testItems(), Skills: staticSkillSource{100: danceSkill100()}}
	skill, _ := rt.deps.SkillData().SkillByID(100)
	now := rt.Now().UnixMilli()

	out := rt.acceptPartyBuff(testDivision, character, character, wire.SkillAction{ActionId: 100}, skill)
	if out.DiagnosticRefusal != "" {
		t.Fatalf("the dance was refused: %q", out.DiagnosticRefusal)
	}
	for _, frame := range out.ActorPrivate {
		if frame.Opcode == wire.OpSkillCastResult {
			t.Fatalf("the dance was refused with a cast-result frame: % x", frame.Payload)
		}
	}
	found := false
	rt.partyAuraMu.Lock()
	for _, aura := range rt.partyAuras {
		if aura.skillID == 100 && aura.radius == 0 && aura.nextPulse-now >= 5000 && aura.nextPulse-now <= 5001 {
			found = true
		}
	}
	rt.partyAuraMu.Unlock()
	if !found {
		t.Fatal("the toggle was not registered on the pulse engine at the authored rhythm")
	}

	before := *character.CurrentMP
	if _, kept := rt.pulseAura(testDivision, character, skill); !kept {
		t.Fatal("the pulse did not hold")
	}
	if drained := before - *character.CurrentMP; drained <= 0 {
		t.Fatalf("the pulse charged no MP: %d", drained)
	}
}
