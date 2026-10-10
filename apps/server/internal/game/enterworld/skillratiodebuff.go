/*
===========================================================================

skillratiodebuff.go - the timed hostile ratio cut of the Water line

Extended content (isro-live-2026), port-only, not v1.150-native. The
live WATER_CANCEL timed tiers past the cap (A: terd, the target's
evasion; B: drht, its hit rate; CANCEL2_A carries both) author a
damage-free timed debuff no native lane admits: the cut is a parameter
write, not an abnormal status - terd/drht are name-attack metadata
words (7F85A0), not Sources entries. The engine side (the block's
slotless timed modifiers, abnormal.ApplyRatioDebuff) carries the
effect; this parser only pins the authored shape.

===========================================================================
*/

package enterworld

const (
	// ratioDebuffEvasion is terd (0x74657264): the target's evasion cut,
	// parameter 9 - the parameter ElectricShock writes natively
	// (callbacks.go:97) and MonsterInstanceStats projects into both hit
	// rolls.
	ratioDebuffEvasion = 0x74657264
	// ratioDebuffHitRate is drht (0x74687264): the target's hit-rate cut,
	// parameter 11.
	ratioDebuffHitRate = 0x74687264
)

/*
================
SkillRatioDebuff

DurationMs is dura's word; EvasionDown and HitDown are terd's and
drht's cut percents - the installed write is the remaining factor
(100 - cut) in the factor-product channel, ElectricShock's arithmetic.
================
*/
type SkillRatioDebuff struct {
	Pinned                bool
	DurationMs            uint32
	EvasionDown, HitDown  uint32
	ThreatFlat, ThreatPct uint32
}

/*
==================
compileSkillRatioDebuff

Admit the complete timed shape: bbuf, one dura equal to the envelope's
duration, terd and/or drht, one tant - nothing else - on a targeted
hostile row (the monster columns Enemy_M; the client tooltips: "Decreases
the enemy's dodge ability" / "hitting ratios"). The v1.150 catalogue
authors the same shape on the family's cap-90 tiers, so admission
carries the extended mastery floor - no v1.150 row reaches it.
==================
*/
func compileSkillRatioDebuff(fields []string, row SkillRow) (SkillRatioDebuff, bool) {
	if len(fields) != 118 || fields[0] != "1" || fields[8] != "2" || fields[68] != "3" ||
		!row.TimingPinned || !row.Consumption.Pinned || !row.ActionRangePinned || row.ActionRange == 0 ||
		row.ChainSub || row.ChainNext != 0 || row.Attack.Present || !row.TargetRequired ||
		row.ActionCastingTimeMs != 0 || row.ActionDurationMs == 0 ||
		row.Consumption.HP != 0 || row.Consumption.HPPercent != 0 || !row.Targets.EnemyM {
		return SkillRatioDebuff{}, false
	}
	// The extended lane's floor: the caller runs this once the row is
	// complete, so Masteries is final here.
	mastery := row.Masteries[0].Level
	if row.Masteries[1].Level > mastery {
		mastery = row.Masteries[1].Level
	}
	if mastery < extendedPastCapMastery {
		return SkillRatioDebuff{}, false
	}
	// A hostile target only: no self, ally or party column. Column 21 is
	// the targeted cast's range word and stays, as the targeted contracts
	// keep it.
	for _, column := range []int{26, 27, 28} {
		if fields[column] != "0" {
			return SkillRatioDebuff{}, false
		}
	}
	for _, column := range []int{15, 16, 17, 19, 20, 24, 25, 31, 32, 33, 56} {
		if fields[column] != "0" {
			return SkillRatioDebuff{}, false
		}
	}
	program, err := CompileSkillProgram(fields)
	if err != nil {
		return SkillRatioDebuff{}, false
	}
	var out SkillRatioDebuff
	seen := make(map[uint32]bool)
	for i := 0; i < program.Len(); i++ {
		op := program.Instruction(i)
		if seen[op.Tag] {
			return SkillRatioDebuff{}, false
		}
		seen[op.Tag] = true
		switch op.Tag {
		case tagBbuf:
		case tagDura:
			if op.Count != 1 || op.Arguments[0] == 0 || op.Arguments[0] != row.EffectDurationMs {
				return SkillRatioDebuff{}, false
			}
			out.DurationMs = op.Arguments[0]
		case ratioDebuffEvasion:
			// Extended (isro-live-2026, M8 s58): the live CANCEL2_A tier
			// authors 105 - the rebalance's over-one-hundred cuts. The bound
			// stays 200: a percent word, never an arbitrary u32.
			if op.Count != 1 || op.Arguments[0] == 0 || op.Arguments[0] > 200 {
				return SkillRatioDebuff{}, false
			}
			out.EvasionDown = op.Arguments[0]
		case ratioDebuffHitRate:
			if op.Count != 1 || op.Arguments[0] == 0 || op.Arguments[0] > 200 {
				return SkillRatioDebuff{}, false
			}
			out.HitDown = op.Arguments[0]
		case tagStatusThreat:
			// tant rides every damage-free hostile cast; the flat word is
			// the debuff's authored aggression.
			if op.Count != 2 {
				return SkillRatioDebuff{}, false
			}
			out.ThreatFlat, out.ThreatPct = op.Arguments[0], op.Arguments[1]
		default:
			return SkillRatioDebuff{}, false
		}
	}
	if !seen[tagBbuf] || !seen[tagDura] || !seen[tagStatusThreat] ||
		!seen[ratioDebuffEvasion] && !seen[ratioDebuffHitRate] || out.DurationMs == 0 {
		return SkillRatioDebuff{}, false
	}
	out.Pinned = true
	return out, true
}
