/*
===========================================================================

linkedredirect_test.go - the redirection hook's arithmetic and debit

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s42): the link is installed through the registry's own ApplyLink, the
victim's resolved formulas are scaled by the masked lane's share, and
the warrior pays inside its own door - held at one HP (the recorded
v1 boundary).

===========================================================================
*/

package action

import (
	"testing"

	"opensro.online/server/internal/game/combat"
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/statuseffect"
)

func TestLinkedRedirectScalesMaskedLaneAndDebitsWarrior(t *testing.T) {
	warrior := testCharacter()
	warrior.Name = "warrior"
	warrior.ID = 11
	victim := testCharacter()
	victim.Name = "victim"
	victim.ID = 12
	hpWarrior, hpVictim := int64(100000), int64(100000)
	warrior.CurrentHP, victim.CurrentHP = &hpWarrior, &hpVictim
	deps := &enterworld.Deps{
		Characters: enterworld.StaticCharacterSource{testDivision: {warrior, victim}},
		Items:      testItems(),
		Skills:     staticSkillSource{},
	}
	rt := NewRuntime(deps, nil)
	now := int64(1_000_000)
	if code := rt.effects.ApplyLink(statuseffect.Link{
		DivisionID: testDivision, SourceName: warrior.Name, TargetName: victim.Name,
		SourceGID: enterworld.ObjectIDForCharacter(warrior), TargetGID: enterworld.ObjectIDForCharacter(victim),
		SourceToken: 1, TargetToken: 2, SkillID: 100, SkillGroup: 9, Group: 1,
		Redirect: true, RedirectMask: 4, RedirectPercent: 40,
		ExpiresAtMs: now + 60_000, StartedAtMs: now,
	}); code != 0 {
		t.Fatalf("ApplyLink refused: %x", code)
	}
	strike := playerStrike{division: testDivision, victim: victim, now: now}
	strike.formulas = []combat.Result{
		{Damage: 1000},                    // physical: masked
		{Damage: 500, MagicalDamage: 500}, // magical: not masked
		{Damage: 100, Blocked: true},      // blocked physical: masked too
	}
	redirects := rt.scaleLinkedRedirect(testDivision, victim, &strike, now)
	if len(redirects) != 1 || redirects[0].source != warrior.Name || redirects[0].amount != 440 {
		t.Fatalf("redirects: %+v", redirects)
	}
	if strike.formulas[0].Damage != 600 || strike.formulas[1].Damage != 500 || strike.formulas[2].Damage != 60 {
		t.Fatalf("scaled formulas: %+v", strike.formulas)
	}
	before := enterworld.CurrentHP(warrior)
	rt.commitLinkedRedirect(testDivision, redirects[0], now)
	after := enterworld.CurrentHP(warrior)
	if after != before-440 {
		t.Fatalf("warrior hp %d -> %d, want -440", before, after)
	}
	// The v1 boundary: the diversion never kills the protector.
	tiny := warrior
	*tiny.CurrentHP = 1
	rt.commitLinkedRedirect(testDivision, redirects[0], now)
	if enterworld.CurrentHP(warrior) != 1 {
		t.Fatalf("the protector died to a redirect: %d", enterworld.CurrentHP(warrior))
	}
}
