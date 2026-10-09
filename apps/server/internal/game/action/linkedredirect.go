/*
===========================================================================

linkedredirect.go - the Warrior's Guard: a linked member's taken damage
is diverted to the warrior

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s42). The live GUARDA tiers past the cap are the lnks pair of
skilllinkedeffect.go with lkdr (enterworld.SkillEffectLink.Redirect):
while the link holds, a share of the damage the protected member TAKES
of the masked lane (4 physical, 8 magical) lands on the warrior
instead - "you divert part of physical damage from one member to
yourself". The victim's formulas are scaled BEFORE the strike commits;
the warrior's debit runs in its own door once the strike landed. The
defensive mirror of linkedmana.go (fed from the attacker's side); no
native row pins lkdr, so no redirect link can exist natively and this
hook is inert for every native strike by construction.

===========================================================================
*/

package action

import (
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/world/simulation"
)

// redirectLanePhysical and redirectLaneMagical are lkdr's mask bits -
// att's own lane flags.
const (
	redirectLanePhysical uint32 = 4
	redirectLaneMagical  uint32 = 8
)

/*
================
scaleLinkedRedirect

Scale the victim's resolved formulas down by each live link's share of
its masked lane, and return the per-source debits. The commit has not
run yet: a strike that never lands diverts nothing (the debits are
applied by the caller only once the victim's commit succeeded).
================
*/
func (rt *Runtime) scaleLinkedRedirect(division string, victim *enterworld.Character, strike *playerStrike, now int64) []linkedRedirectDebit {
	if rt.effects == nil {
		return nil
	}
	links, _ := rt.effects.RedirectLinks(division, victim.Name, now)
	var debits []linkedRedirectDebit
	for _, link := range links {
		var amount uint32
		for i := range strike.formulas {
			formula := &strike.formulas[i]
			if formula.Damage == 0 {
				continue
			}
			lane := redirectLanePhysical
			if formula.MagicalDamage != 0 {
				lane = redirectLaneMagical
			}
			if link.RedirectMask&lane == 0 {
				continue
			}
			share := formula.Damage * link.RedirectPercent / 100
			if share == 0 || share >= formula.Damage {
				continue
			}
			formula.Damage -= share
			formula.MagicalDamage -= min(formula.MagicalDamage, share)
			amount += share
		}
		if amount > 0 {
			debits = append(debits, linkedRedirectDebit{source: link.SourceName, sourceGID: link.SourceGID, amount: amount})
		}
	}
	return debits
}

type linkedRedirectDebit struct {
	source    string
	sourceGID uint32
	amount    uint32
}

/*
================
commitLinkedRedirect

Debit the warrior its diverted share inside its own door and push its
private vitals. Inferred v1 boundary, recorded deliberately (M8 s42):
the diversion wounds the protector but never kills - a protector's
death rides the strike lane's full death machinery, which a side debit
must not fake; the debit is held at one HP.
================
*/
func (rt *Runtime) commitLinkedRedirect(division string, debit linkedRedirectDebit, now int64) {
	source := rt.findCharacter(division, debit.source)
	if source == nil || enterworld.ObjectIDForCharacter(source) != debit.sourceGID {
		return
	}
	var frame wire.Frame
	if !rt.deps.Update(source, "linked-damage-redirect", func() bool {
		if source.DeletePending || !enterworld.CharacterAlive(source) {
			return false
		}
		hp := enterworld.CurrentHP(source)
		if hp <= 1 {
			return false
		}
		take := min(uint32(hp-1), debit.amount)
		*source.CurrentHP = int64(hp) - int64(take)
		frame = wire.Frame{Opcode: simulation.OpVitalsUpdate,
			Payload: simulation.VitalsRefreshWithSourcePayload(debit.sourceGID, simulation.VitalsSourceSkillRecovery, rt.publishedVitals(division, source))}
		return true
	}) {
		return
	}
	if frame.Opcode != 0 && rt.PushCharacterFrames != nil {
		rt.PushCharacterFrames(division, source.Name, []wire.Frame{frame})
	}
}
