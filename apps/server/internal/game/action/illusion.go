/*
===========================================================================

illusion.go - the Warlock's Illusion: a random lower-level player's look

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s59). The live CONFUSIONA_ILLUSION rows (msch mode 3 - a family with
ZERO v1.150 carriers, measured) transform the Warlock into a randomly
chosen character no higher than the level word: the pick runs over the
division's roster, the copy is the Duplicate's own (duplicateLook -
the model, the shape byte, the nine worn slots), and the instance
lives on the caster exactly as a Duplicate's does. The vanilla client
varies the model from viewer to viewer; this port keeps ONE model per
cast for every viewer - the recorded v1 boundary (a per-connection
replication lane is not in the v1.150 wire this port owns).

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
acceptIllusion

An untargeted self cast: admission and cost as every self effect, the
pick from the live roster at commit, the Duplicate's look installed
under msch mode 3.
================
*/
func (rt *Runtime) acceptIllusion(division string, c, snapshot *enterworld.Character, cast wire.SkillAction, skill enterworld.SkillRow, now int64) OpResult {
	if !skill.Illusion.Pinned || cast.HasTarget || cast.HasGroundTarget ||
		!enterworld.CharacterAlive(snapshot) || !enterworld.SkillLearned(snapshot, skill.ID) {
		return OpResult{DiagnosticRefusal: "illusion-admission-refused"}
	}
	if rt.skillCastPostureBlocked(division, snapshot, now) || rt.hasOpenSkillCast(division, snapshot.Name) {
		return OpResult{DiagnosticRefusal: "illusion-action-busy"}
	}
	if code := rt.skillAdmission(division, snapshot, skill, now, nil, nil, admitExecution); code != 0 {
		return offensiveRefusal(code)
	}
	// The pick: a live character of the division, no higher than the word,
	// never the caster. Inferred, recorded deliberately (M8 s59): with no
	// candidate the cast refuses - a disguise the roster cannot build is
	// not invented.
	candidate := rt.illusionCandidate(division, snapshot, skill.Illusion.MaxLevel, now)
	if candidate == nil {
		return OpResult{DiagnosticRefusal: "illusion-no-candidate"}
	}
	look, ok := rt.duplicateLook(candidate)
	if !ok {
		return OpResult{DiagnosticRefusal: "illusion-look-unavailable"}
	}
	token := atomic.AddUint32(&rt.castTokenCounter, 1)
	casterGID := enterworld.ObjectIDForCharacter(snapshot)
	var refusal uint16
	var installed []wire.Frame
	if !rt.deps.Update(c, "illusion", func() bool {
		if !enterworld.CharacterAlive(c) || !enterworld.SkillLearned(c, skill.ID) {
			return false
		}
		cost, code := rt.offensivePhaseCost(division, c, skill, now, nil)
		if refusal = code; code != 0 {
			return false
		}
		rt.startSkillCast(division, c, skill, now)
		rt.commitOffensivePhaseCost(division, c, skill, cost, now, false)
		var ok bool
		installed, ok = rt.commitCharacterEffect(division, c, skill, token, statuseffect.StateActive, false, look, now)
		return ok
	}) {
		if refusal != 0 {
			return offensiveRefusal(refusal)
		}
		return OpResult{DiagnosticRefusal: "illusion-commit-refused"}
	}
	lifetime, _ := skill.ActionLifecycleMs()
	rt.queueSkillFinalize(division, snapshot.Name, casterGID, now+int64(lifetime), wire.SkillCastFinalizeFrame(token))
	open := wire.SkillCastAtTargetFrame(wire.SkillCastSuccess{SkillId: skill.ID, CasterGid: casterGID, InstanceToken: token})
	vitals := wire.Frame{Opcode: simulation.OpVitalsUpdate, Payload: simulation.VitalsRefreshWithSourcePayload(casterGID, simulation.VitalsSourceSkillRecovery, rt.publishedVitals(division, c))}
	frames := append([]wire.Frame{open}, installed...)
	return OpResult{Frames: append(frames, vitals), Broadcast: frames}
}

/*
================
illusionCandidate

The division's live roster filtered to characters below the caster's
own level (the tooltip's "lower level than you") and no higher than
the word; the roll picks one on the combat stream.
================
*/
func (rt *Runtime) illusionCandidate(division string, caster *enterworld.Character, maxLevel uint32, now int64) *enterworld.Character {
	candidates := []*enterworld.Character{}
	for _, other := range rt.deps.CharactersForDivision(division) {
		if other == nil || other.Name == caster.Name || other.DeletePending || !enterworld.CharacterAlive(other) {
			continue
		}
		level := int64(0)
		if other.Level != nil {
			level = *other.Level
		}
		own := int64(0)
		if caster.Level != nil {
			own = *caster.Level
		}
		if level >= own || level > int64(maxLevel) {
			continue
		}
		candidates = append(candidates, other)
	}
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	roll, err := rt.CombatRoll()
	if err != nil {
		return candidates[0]
	}
	return candidates[int(roll)%len(candidates)]
}
