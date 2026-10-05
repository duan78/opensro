/*
===========================================================================

playervictim.go - what a landed hit does to a player, whoever struck it

SkillCombat_CalculateHitOutcome resolves each impact against the victim
(the wall split of 58EC6C, the status rolls of 590680) and
SkillCombat_ApplyResultRecipients (593800) commits them: the HP debit,
the standing wall's drain, the armour and shield wear of a struck
victim, then either the death (pkdeath.go) or the battle state, the
statuses and the skc damage cancellation. A monster's hit and a player's
hit take this one path; only the attacker's own rolls and bookkeeping
differ.

===========================================================================
*/

package action

import (
	"opensro.online/server/internal/game/abnormal"
	"opensro.online/server/internal/game/combat"
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/internal/vitals"
)

/*
================
playerStrike

One attacker's hits on one player, resolved before the victim's door.
================
*/
type playerStrike struct {
	division string
	victim   *enterworld.Character
	killer   deathKiller
	skill    enterworld.SkillRow
	formulas []combat.Result
	splits   []wallSplit
	wall     standingWall
	walled   bool
	records  []abnormal.Record
	owner    *playerAbnormalOwner
	now      int64
}

/*
================
playerStruck

What the commit did.
================
*/
type playerStruck struct {
	impacts, absorb  []wire.SkillCastTargetImpact
	fatal, struck    bool
	deathEffects     []wire.Frame
	deathProgression []wire.Frame
	battle           []wire.Frame
	wear             wearFrames
	owner            *playerAbnormalOwner
}

/*
================
planPlayerStrike

58EC6C and 590680 per impact against a player defender: the hit behind
the victim's standing wall, then the attacker's status roll (roll), which
a blocked or ck-killed impact skips (5905FB). False when an impact cannot
be resolved, or lands nothing on an unwalled victim.
================
*/
func (rt *Runtime) planPlayerStrike(s *playerStrike, actor criticalActor, attacker, defender combat.Stats,
	roll func(wall *enterworld.SkillWall) ([]abnormal.Record, error)) bool {
	s.wall, s.walled = rt.standingWallOf(s.division, s.victim.Name)
	var wallRule *enterworld.SkillWall
	if s.walled {
		wallRule = &s.wall.wall
	}
	for range s.skill.Attack.ImpactCount {
		split, err := rt.resolveCombatBehindWall(actor, s.skill, attacker, defender, wallRule)
		formula := split.Defender
		if err != nil || formula.Damage == 0 && !s.walled && !formula.Blocked && !formula.Slain {
			return false
		}
		s.splits = append(s.splits, wallSplit{absorbed: split.Absorbed, flags: formula.ResultFlags, covered: split.Covered})
		s.formulas = append(s.formulas, formula)
		if formula.Blocked || formula.Slain {
			continue
		}
		records, err := roll(wallRule)
		if err != nil {
			return false
		}
		s.records = append(s.records, records...)
	}
	if len(s.formulas) == 0 {
		return false
	}
	s.owner = rt.newPlayerAbnormalOwner(s.division, s.victim, s.now)
	s.owner.sources = rt.captureAbnormalSources(s.division, s.owner.block, s.records)
	return true
}

/*
================
strikePlayerInDoor

593800's recipient side inside the victim's door. A record a standing wall
absorbs whole skips the recipient branch (593AE8); a struck victim's
landed hits wear its armour and its blocks its shield (593C9F/593CB1).
================
*/
func (rt *Runtime) strikePlayerInDoor(s playerStrike) playerStruck {
	c := s.victim
	out := playerStruck{owner: s.owner}
	hit := abnormal.HitContext{Attack: s.skill.ReplacementPinned && s.skill.Replacement.MatchesExecutionSelector}
	_, _, remaining, _ := rt.playerKeeperVitals(s.division, c)
	for _, formula := range s.formulas {
		hit.Magical = hit.Magical || formula.MagicalDamage != 0
		debit := int64(vitals.HitDebit(uint32(remaining), formula.Damage))
		if formula.Slain {
			debit = remaining // 58F778: the ck kill marks the target dead
		}
		remaining -= debit
		out.fatal = remaining == 0
		out.impacts = append(out.impacts, wire.SkillCastTargetImpact{
			ResultFlags: formula.ResultFlags,
			// Native 585664 serializes the full hit independently of HP.
			Damage:  formula.Damage,
			Fatal:   out.fatal,
			Blocked: formula.Blocked,
			Slain:   formula.Slain,
		})
		if out.fatal {
			break
		}
	}
	c.CurrentHP = &remaining
	out.struck = len(out.impacts) > 0
	if s.walled && !s.skill.WallBypass {
		var absorbed uint32
		out.absorb, absorbed = wallRecords(s.wall, s.splits, len(out.impacts))
		rt.drainWall(s.division, c.Name, s.wall.token, absorbed)
		out.struck = !allWallAbsorbed(out.absorb)
	}
	if out.struck {
		var tally wearTally
		for _, impact := range out.impacts {
			// 58F784 jumps past the landed count (58F79F): a ck kill wears nothing.
			if !impact.Slain {
				tally.note(impact.Blocked, false)
			}
		}
		for _, roll := range [...]struct{ mode, count uint8 }{{wearArmour, tally.armour}, {wearShield, tally.shield}} {
			taken := rt.rollEquipmentWear(s.division, c, roll.mode, roll.count)
			out.wear.actor = append(out.wear.actor, taken.actor...)
			out.wear.public = append(out.wear.public, taken.public...)
		}
	}
	if out.fatal {
		out.deathEffects, out.deathProgression = rt.settlePlayerDeathInDoor(s.division, c, s.killer, s.now)
		out.owner = rt.clearPlayerAbnormalInDoor(s.division, c, s.now)
	} else {
		out.battle = rt.enterBattleState(s.division, c, s.now)
		out.owner.applyHit(hit, s.records)
		// 58F72F: a landed hit tests the victim's skc damage masks.
		rt.cancelEffectsOnDamage(s.division, c, s.skill.Attack.Flags, s.now)
	}
	return out
}

/*
================
playerStruckFrames

The struck victim's publication after its door: statuses, wear, the death
(effects retired before the life change, the zero baseline, then 0x3122),
the seated recipient standing, and the battle state. public goes to every
observer, victim to the victim alone.
================
*/
func (rt *Runtime) playerStruckFrames(division string, victim *enterworld.Character, struck playerStruck, now int64) (public, private []wire.Frame) {
	statuses := rt.playerAbnormalPublication(division, victim, struck.owner)
	public = append(public, statuses.public...)
	private = append(private, statuses.actor...)
	public = append(public, struck.wear.public...)
	private = append(private, struck.wear.actor...)
	if struck.fatal {
		if rt.PushCharacterFrames != nil && rt.PushDivisionPeerFrames != nil {
			// Native death retires effects before publishing the life change.
			// Enqueue under the action lock, before a rebirth/new application
			// can reuse these wire tokens.
			rt.publishBodyStatus(division, victim.Name, struck.deathEffects)
		} else {
			public = append(public, struck.deathEffects...)
		}
		life := beginFatalLifePublication(enterworld.ObjectIDForCharacter(victim))
		// B245/B505 retains fatal damage until the client impact callback; the
		// death-sourced 0x33A6 advances the wire baseline to zero without
		// applying it twice, and 0x3122 owns the life-state transition.
		baseline := life.publishDeathBaseline()
		public = append(public, wire.Frame{Opcode: baseline.opcode, Payload: baseline.payload})
		dead := life.publishDead()
		public = append(public, wire.Frame{Opcode: dead.opcode, Payload: dead.payload})
	}
	// Entering battle (4E1DF0, from ProcessNormalHit) follows the hit.
	if struck.struck && !struck.fatal {
		public = append(public, rt.offensiveResultRecipient(division, victim, now)...)
	}
	public = append(public, struck.battle...)
	private = append(private, struck.deathProgression...)
	return public, private
}
