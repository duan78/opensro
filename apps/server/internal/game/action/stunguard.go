/*
===========================================================================

stunguard.go - the Warlock's Stun Link: attackers of the covered member
roll a reactive stun

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s43). The live SOULA_STUNLINK tiers past the cap are the anonymous lnks
pair with the st block: "covers one member of the party with a mask of
horror... enemies who are weak lose consciousness" - once a strike on
the covered member commits, each live guard rolls its stun on the
attacker (the level ceiling first, then the native chance stream inside
the monster state). No native row pins the guard, so no native strike
ever rolls it - void-safe by construction, as the s42 redirect.

===========================================================================
*/

package action

import (
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/world/monster"
)

/*
================
rollStunGuards

Every live stun guard linked to the victim rolls its block on the
striking monster. The strike has already committed: the guard is a
consequence, never a shield.
================
*/
func (rt *Runtime) rollStunGuards(division string, victim *enterworld.Character, attacker monster.Instance, now int64) {
	if rt.effects == nil || rt.Monsters == nil {
		return
	}
	guards, _ := rt.effects.StunGuards(division, victim.Name, now)
	for _, guard := range guards {
		if uint32(attacker.Ref.Level) > guard.GuardCeilingLevel {
			continue
		}
		rt.Monsters.ApplyStunGuard(division, attacker.Gid, guard.SourceGID, guard.SourceName,
			guard.GuardDurationMs, guard.GuardChance, guard.GuardLevel, now)
	}
}
