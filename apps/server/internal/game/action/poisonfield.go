/*
===========================================================================

poisonfield.go - the rogue's planted, ticking poison field

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s51). The live POISONA_FIELD tiers past the cap plant a field that
PERSISTS: every due scan the ps poison block lands on the monsters
inside the kind-3 radius, the field dying at its own dura. The poison
outlives the field by far (its own duration word) and ticks on the
native 2-second damage-over-time update.

===========================================================================
*/

package action

import (
	"sync/atomic"

	"opensro.online/server/internal/game/abnormal"
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/world/simulation"
	"opensro.online/server/internal/game/world/skillobject"
)

/*
================
acceptPoisonField

Admission and preparation follow the combat-trap owner; the release
plants a PERSISTING object whose scans the tick driver feeds.
================
*/
func (rt *Runtime) acceptPoisonField(division string, c, snapshot *enterworld.Character, cast wire.SkillAction, skill enterworld.SkillRow, now int64, pending *pendingProjectileCast) (OpResult, skillCastDecision) {
	field := skill.PoisonField
	if !field.Pinned {
		return OpResult{DiagnosticRefusal: "poison-field-admission-refused"}, skillCastRefused
	}
	if out, decision, done := rt.beginUntargetedCast(division, c, snapshot, cast, skill, now, pending, func(p *pendingProjectileCast) { p.poisonField = true }); done {
		return out, decision
	}
	rt.clearCurrentSkillCommand(division, c.Name)
	lease, present := rt.EntryPopulationLease(division, c.Name)
	if !present {
		return OpResult{DiagnosticRefusal: "poison-field-population"}, skillCastRefused
	}
	gid := enterworld.ObjectIDForCharacter(c)
	at := rt.liveSpawn(simulation.WorldKey(division, c.Name), c, now)
	var refusal uint16
	var effects []wire.Frame
	if !rt.deps.Update(c, "release-poison-field", func() bool {
		if !enterworld.CharacterAlive(c) || !enterworld.SkillLearned(c, skill.ID) {
			return false
		}
		cost, code := rt.offensivePhaseCost(division, c, skill, now, pending)
		refusal = code
		if code != 0 {
			return false
		}
		rt.commitOffensivePhaseCost(division, c, skill, cost, now, true)
		object, err := rt.SkillObjects.Create(skillobject.Object{
			Division: division, Population: lease, OwnerGID: gid, OwnerName: c.Name, CreatedMs: now,
			Program: skillobject.Program{SkillID: skill.ID, DurationMs: field.DurationMs, ScanMs: field.ScanMs,
				Radius: field.Radius, Combat: true, Persist: true},
			Spawn: wire.SkillObjectSpawn{Region: at.RegionID, X: float32(at.X), Y: float32(at.Y), Z: float32(at.Z), Heading: at.Angle},
		})
		if err != nil {
			return false
		}
		var installed bool
		effects, installed = rt.commitCharacterEffect(division, c, skill, object.OwnerEffect, 0, false, EffectPresentation{Phase: 1}, now)
		return installed
	}) {
		if refusal != 0 {
			return offensiveRefusal(refusal), skillCastRefused
		}
		return OpResult{DiagnosticRefusal: "poison-field-commit-refused"}, skillCastRefused
	}
	casterGID := enterworld.ObjectIDForCharacter(c)
	token := atomic.AddUint32(&rt.castTokenCounter, 1)
	released := wire.SkillCastReleaseFrame(token, 0)
	rt.queueSkillCastClose(division, c.Name, casterGID, token, skill, 0, now+int64(skill.ActionDurationMs))
	vitals := wire.Frame{Opcode: simulation.OpVitalsUpdate, Payload: simulation.VitalsRefreshWithSourcePayload(casterGID, simulation.VitalsSourceSkillRecovery, rt.publishedVitals(division, c))}
	out := OpResult{Frames: append([]wire.Frame{released}, effects...), Broadcast: []wire.Frame{released}, ActorPrivate: append(effects, vitals)}
	return mergeOpResults(out, OpResult{ActorPrivate: []wire.Frame{vitals}}), skillCastAccepted
}

/*
================
applyPoisonFieldScan

One due scan of a persisting field: every monster the registry matched
rolls the ps block on the native chance stream; a pass installs the
poison record through the ordinary block transaction, whose 2-second
update owns the damage ticks afterwards.
================
*/
func (rt *Runtime) applyPoisonFieldScan(division string, owner *enterworld.Character, field enterworld.SkillPoisonField, victims []uint32, now int64) {
	for _, victim := range victims {
		rt.Monsters.ApplyPoisonField(division, victim, enterworld.ObjectIDForCharacter(owner), owner.Name,
			field.PoisonMs, field.Chance, field.Damage, now)
	}
	_ = abnormal.Sources
}
