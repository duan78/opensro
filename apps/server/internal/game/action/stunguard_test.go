/*
===========================================================================

stunguard_test.go - the reactive stun's roll, ceiling and install

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s43): the guard installed through the registry's ApplyLink, the
ceiling checked, the chance rolled on the native stream (deterministic
in the fixture), the stun landed through the block transaction.

===========================================================================
*/

package action

import (
	"testing"

	"opensro.online/server/internal/game/abnormal"
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/statuseffect"
	"opensro.online/server/internal/game/world/monster"
)

func newStunGuardFixture(t *testing.T) (*Runtime, *enterworld.Character, monster.Instance) {
	t.Helper()
	rt, _, member, mob := newCombatTestRuntime(t, 1000)
	warlock := nearbyCharacter(rt, member, 21, "warlock", 30)
	now := int64(1_000_000)
	if code := rt.effects.ApplyLink(statuseffect.Link{
		DivisionID: testDivision, SourceName: warlock.Name, TargetName: member.Name,
		SourceGID: enterworld.ObjectIDForCharacter(warlock), TargetGID: enterworld.ObjectIDForCharacter(member),
		SourceToken: 1, TargetToken: 2, SkillID: 100, SkillGroup: 9, Group: 0,
		GuardDurationMs: 5000, GuardChance: 100, GuardLevel: 10, GuardCeilingLevel: 150,
		ExpiresAtMs: now + 120_000, StartedAtMs: now,
	}); code != 0 {
		t.Fatalf("ApplyLink refused: %x", code)
	}
	return rt, member, mob
}

func TestStunGuardRollsOnTheAttackerBelowTheCeiling(t *testing.T) {
	rt, member, mob := newStunGuardFixture(t)
	rt.rollStunGuards(testDivision, member, mob, 1_000_000)
	live, _ := rt.Monsters.Get(testDivision, mob.Gid)
	if live.Abnormal == nil || !live.Abnormal.Has(abnormal.Stun) {
		t.Fatalf("the attacker was not stunned: %+v", live.Abnormal)
	}
}

func TestStunGuardSkipsAttackersAboveTheCeiling(t *testing.T) {
	rt, member, mob := newStunGuardFixture(t)
	// The ceiling is 150: a level-151 attacker never rolls.
	tall := mob
	tall.Ref.Level = 151
	rt.rollStunGuards(testDivision, member, tall, 1_000_000)
	live, _ := rt.Monsters.Get(testDivision, mob.Gid)
	if live.Abnormal != nil {
		t.Fatalf("an attacker above the ceiling was stunned: %+v", live.Abnormal)
	}
}
