/*
===========================================================================

hwan.go - the "Aura of Blood": the party's berserk mode runs longer

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s63). The live SPELLP_PARTY_HWAN_UP rows cast a five-minute party aura
(the kind-1 efr, radius 300, eight recipients, party select) whose one
word hwdu extends every party member's berserk mode by its seconds -
the recorded inference (the naming evidence: the skill's HWAN_UP name,
the wire calling BerserkPoints "the persistent Hwan gauge", the mall
pills acting on pill DURATION; the row carries no puls, so the
extension applies once at the cast). Recipients outside the berserk
mode are untouched - the word extends a running mode, it does not
start one.

===========================================================================
*/

package action

import (
	"sync/atomic"

	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/statuseffect"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/world/simulation"
)

/*
================
acceptHwanAura

An untargeted self cast: admission and cost as every self effect, the
extension applied to each party recipient inside the door, the caster's
aura instance installed for the board and released at the dura.
================
*/
func (rt *Runtime) acceptHwanAura(division string, c, snapshot *enterworld.Character, cast wire.SkillAction, skill enterworld.SkillRow, now int64) OpResult {
	extension := skill.TimedEffect.HwanDurationMs
	if extension == 0 || cast.HasTarget || cast.HasGroundTarget ||
		!enterworld.CharacterAlive(snapshot) || !enterworld.SkillLearned(snapshot, skill.ID) {
		return OpResult{DiagnosticRefusal: "hwan-aura-admission-refused"}
	}
	if rt.skillCastPostureBlocked(division, snapshot, now) || rt.hasOpenSkillCast(division, snapshot.Name) {
		return OpResult{DiagnosticRefusal: "hwan-aura-action-busy"}
	}
	if code := rt.skillAdmission(division, snapshot, skill, now, nil, nil, admitExecution); code != 0 {
		return offensiveRefusal(code)
	}
	from := rt.liveSpawn(simulation.WorldKey(division, snapshot.Name), snapshot, now)
	// The row's Self column names the caster inside its party selection
	// (col 26 = 1), so the recipients are the party walk PLUS the caster.
	recipients := rt.partyMembersAround(division, snapshot, from, skill.TimedEffect.Area.Radius, false, now)
	names := make([]string, 0, len(recipients)+1)
	for _, gid := range recipients {
		if member := rt.findCharacterByGid(division, gid); member != nil {
			names = append(names, member.Name)
		}
	}
	names = append(names, snapshot.Name)

	token := atomic.AddUint32(&rt.castTokenCounter, 1)
	casterGID := enterworld.ObjectIDForCharacter(snapshot)
	var refusal uint16
	var installed []wire.Frame
	if !rt.deps.Update(c, "hwan-aura", func() bool {
		if !enterworld.CharacterAlive(c) || !enterworld.SkillLearned(c, skill.ID) {
			return false
		}
		cost, code := rt.offensivePhaseCost(division, c, skill, now, nil)
		if refusal = code; code != 0 {
			return false
		}
		rt.startSkillCast(division, c, skill, now)
		presentation := EffectPresentation{
			Phase: 1, AreaSourceGID: casterGID, AreaSourceName: c.Name,
		}
		var ok bool
		installed, ok = rt.commitCharacterEffect(division, c, skill, token, statuseffect.StateActive, true, presentation, now)
		if !ok {
			return false
		}
		rt.commitOffensivePhaseCost(division, c, skill, cost, now, false)
		return true
	}) {
		if refusal != 0 {
			return offensiveRefusal(refusal)
		}
		return OpResult{DiagnosticRefusal: "hwan-aura-commit-refused"}
	}

	// The extension: every recipient inside their own door; a recipient
	// not in the berserk mode is untouched (the word extends a running
	// mode, it never starts one), and the tick index follows the new
	// expiry exactly as the activation wrote it.
	for _, name := range names {
		member := rt.findCharacter(division, name)
		if member == nil {
			continue
		}
		var vitals *wire.Frame
		rt.deps.Update(member, "hwan-aura-extend", func() bool {
			if member.DeletePending || !enterworld.CharacterAlive(member) || member.BerserkUntilMs <= now {
				return false
			}
			member.BerserkUntilMs += int64(extension)
			rt.berserkActors.Store(simulation.WorldKey(division, member.Name), berserkExpiry{division, member.Name, member.BerserkUntilMs})
			frame := wire.Frame{Opcode: simulation.OpVitalsUpdate,
				Payload: simulation.VitalsRefreshWithSourcePayload(enterworld.ObjectIDForCharacter(member), simulation.VitalsSourceSkillRecovery, rt.publishedVitals(division, member))}
			vitals = &frame
			return true
		})
		if vitals != nil && rt.PushCharacterFrames != nil {
			rt.PushCharacterFrames(division, member.Name, []wire.Frame{*vitals})
		}
	}

	lifetime, _ := skill.ActionLifecycleMs()
	rt.queueSkillFinalize(division, snapshot.Name, casterGID, now+int64(lifetime), wire.SkillCastFinalizeFrame(token))
	open := wire.SkillCastAtTargetFrame(wire.SkillCastSuccess{SkillId: skill.ID, CasterGid: casterGID, InstanceToken: token})
	frames := append([]wire.Frame{open}, installed...)
	return OpResult{Frames: frames, Broadcast: frames}
}
