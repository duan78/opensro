/*
===========================================================================

ratiodebuff.go - the timed hostile ratio cut of the Water line

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s34). The live WATER_CANCEL timed tiers past the cap cast a damage-free
debuff at a monster target: walk into reach like any targeted command,
then install the slotless timed ratio writes (the engine abnormal
ApplyRatioDebuff shipped in s33) and file tant's aggression. The
tooltips: "Decreases the enemy's dodge ability" (terd, parameter 9) and
"Decreases the enemy's hitting ratios" (drht, parameter 11); the
installed write is the remaining factor, ElectricShock's arithmetic.

===========================================================================
*/

package action

import (
	"sync/atomic"

	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/world/monster"
	"opensro.online/server/internal/game/world/simulation"
)

// The unified parameter graph's ids: parameter 9 is the evasion rate,
// parameter 11 the hit rate (combat/statWrites, the monster table's
// same order).
const (
	ratioParamEvasion uint16 = 9
	ratioParamHit     uint16 = 11
)

/*
================
acceptRatioDebuff

Validate the command, then own it as a debuff intent that walks the
caster into reach and casts once there (the capture intent's monster
lane; the support walk is player-targeted only).
================
*/
func (rt *Runtime) acceptRatioDebuff(division string, character, snapshot *enterworld.Character, cast wire.SkillAction, skill enterworld.SkillRow) OpResult {
	if !cast.HasTarget || cast.HasGroundTarget || cast.TargetGid == 0 {
		return offensiveRefusal(0x3011)
	}
	if !enterworld.CharacterAlive(snapshot) || !enterworld.SkillLearned(snapshot, skill.ID) || rt.Monsters == nil {
		return OpResult{DiagnosticRefusal: "ratio-debuff-admission-refused"}
	}
	intent := basicAttackIntent{DebuffCast: true, DivisionID: division, CharacterName: snapshot.Name, TargetGid: cast.TargetGid, SkillID: skill.ID}
	rt.setCombatIntent(intent)
	return rt.advanceRatioDebuffIntent(character, intent, rt.Now().UnixMilli())
}

/*
==================
advanceRatioDebuffIntent

One tick of the debuff intent: give up when the caster or the monster
is gone, keep walking while out of reach, else stop, face the target
and cast (monstercapture.go's walk).
==================
*/
func (rt *Runtime) advanceRatioDebuffIntent(character *enterworld.Character, intent basicAttackIntent, now int64) OpResult {
	division := intent.DivisionID
	snapshot := rt.characterSnapshot(division, character)
	skills := rt.deps.SkillData()
	if snapshot == nil || snapshot.DeletePending || !enterworld.CharacterAlive(snapshot) ||
		skills == nil || rt.skillCastPostureBlocked(division, snapshot, now) {
		rt.ClearCombatIntent(division, intent.CharacterName)
		return OpResult{}
	}
	skill, known := skills.SkillByID(intent.SkillID)
	target, found := rt.characterMonster(division, snapshot, intent.TargetGid)
	mover, moving := rt.Monsters.Mover(division, intent.TargetGid)
	if !known || !skill.RatioDebuff.Pinned || !found || !moving || target.CurrentHP == 0 {
		rt.ClearCombatIntent(division, intent.CharacterName)
		return offensiveRefusal(0x3006)
	}
	_, loadout, err := rt.playerCombatStats(division, snapshot)
	if err != nil {
		rt.ClearCombatIntent(division, intent.CharacterName)
		return OpResult{}
	}
	spacing, ok := rt.playerToMonsterCombatSpacing(snapshot, target, rt.playerActionReach(division, snapshot, skill, loadout))
	if !ok {
		rt.ClearCombatIntent(division, intent.CharacterName)
		return OpResult{}
	}
	pose := mover.LivePoseAt(now, nil)
	to := simulation.Spawn{RegionID: pose.RegionID, X: pose.X, Y: pose.Y, Z: pose.Z}
	worldKey := simulation.WorldKey(division, snapshot.Name)
	from := rt.liveSpawn(worldKey, snapshot, now)
	if !spacing.Contains(from, to) {
		if !rt.pursuitSteerDue(intent, worldKey, snapshot, to, now) {
			return OpResult{}
		}
		return rt.approachIntentTarget(character, snapshot, intent, spacing, from, to, now)
	}
	rt.ClearCombatIntent(division, intent.CharacterName)
	transition, transitioned := rt.enterBasicAttackRange(character, snapshot, worldKey, to, now)
	if !transitioned {
		return OpResult{}
	}
	return prependOpResult(transition, rt.executeRatioDebuff(division, character, skill, target, to, now))
}

/*
==================
executeRatioDebuff

Admission (execution mask), the cost and cooldown, the cast frames,
then the slotless timed writes and tant's aggression event.
==================
*/
func (rt *Runtime) executeRatioDebuff(division string, c *enterworld.Character, skill enterworld.SkillRow, target monster.Instance, at simulation.Spawn, now int64) OpResult {
	snapshot := rt.characterSnapshot(division, c)
	if snapshot == nil || rt.hasOpenSkillCast(division, snapshot.Name) {
		return OpResult{DiagnosticRefusal: "ratio-debuff-action-busy"}
	}
	if code := rt.skillAdmission(division, snapshot, skill, now, &admitTarget{motion: target.Motion.StateAt(now), at: at}, nil, admitExecution); code != 0 {
		return offensiveRefusal(code)
	}
	debuff := skill.RatioDebuff
	until := now + int64(debuff.DurationMs)
	writes := []simulation.RatioWrite{}
	if debuff.EvasionDown != 0 {
		writes = append(writes, simulation.RatioWrite{Param: ratioParamEvasion, RemainingPercent: float32(100 - debuff.EvasionDown), Until: until})
	}
	if debuff.HitDown != 0 {
		writes = append(writes, simulation.RatioWrite{Param: ratioParamHit, RemainingPercent: float32(100 - debuff.HitDown), Until: until})
	}
	token := atomic.AddUint32(&rt.castTokenCounter, 1)
	casterGID := enterworld.ObjectIDForCharacter(snapshot)
	var refusal uint16
	if !rt.deps.Update(c, "ratio-debuff", func() bool {
		if !enterworld.CharacterAlive(c) || !enterworld.SkillLearned(c, skill.ID) {
			return false
		}
		cost, code := rt.offensivePhaseCost(division, c, skill, now, nil)
		if refusal = code; code != 0 {
			return false
		}
		rt.startSkillCast(division, c, skill, now)
		rt.commitOffensivePhaseCost(division, c, skill, cost, now, false)
		return true
	}) {
		if refusal != 0 {
			return offensiveRefusal(refusal)
		}
		return OpResult{DiagnosticRefusal: "ratio-debuff-commit-refused"}
	}
	if !rt.Monsters.ApplyRatioDebuff(division, target.Gid, writes) {
		return OpResult{DiagnosticRefusal: "ratio-debuff-target-gone"}
	}
	// tant rides every damage-free hostile cast: one positive aggression
	// event toward the caster, as the status casts file theirs.
	rt.recordSkillHostility(division, target.Gid, []simulation.HostilityEvent{
		{Attacker: casterGID, Aggression: int32(debuff.ThreatFlat), Percent: int32(debuff.ThreatPct)},
	}, now)

	cast := wire.SkillCastAtTargetFrame(wire.SkillCastSuccess{SkillId: skill.ID, CasterGid: casterGID, InstanceToken: token, OwnerOrTargetGid: target.Gid})
	lifetime, _ := skill.ActionLifecycleMs()
	rt.queueSkillFinalize(division, snapshot.Name, casterGID, now+int64(skill.ActionCastingTimeMs), wire.SkillCastReleaseFrame(token, casterGID))
	rt.queueSkillFinalize(division, snapshot.Name, casterGID, now+int64(lifetime), wire.SkillCastFinalizeFrame(token))
	vitals := wire.Frame{Opcode: simulation.OpVitalsUpdate, Payload: simulation.VitalsRefreshWithSourcePayload(casterGID, simulation.VitalsSourceSkillRecovery, rt.publishedVitals(division, c))}
	frames := []wire.Frame{cast, vitals}
	return OpResult{Frames: frames, Broadcast: []wire.Frame{cast}, ActorPrivate: []wire.Frame{vitals}}
}
