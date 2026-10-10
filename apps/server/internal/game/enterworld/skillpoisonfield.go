/*
===========================================================================

skillpoisonfield.go - the rogue's planted poison field

Extended content (isro-live-2026), port-only, not v1.150-native. The
live POISONA_FIELD tiers past the cap plant a field that PERSISTS and
TICKS: every scan the ps poison block lands on the monsters inside the
kind-3 trigger radius, the field dying at its own dura - unlike the
combat trap, which explodes once and retires. The poison far outlives
the field (its own duration word). The v1.150 catalogue authors the
same shape on the family's cap-90 tiers, so admission carries the
mastery floor.

===========================================================================
*/

package enterworld

import "opensro.online/server/internal/game/abnormal"

const (
	poisonFieldTrigger = 0x00006566 // efr; kind checked in the predicate
	poisonFieldPulse   = 0x70756c73 // puls
	poisonFieldStatus  = 0x00007073 // ps: the poison block
)

/*
================
SkillPoisonField

ScanMs is puls' word, Radius the kind-3 efr's, DurationMs the field's
own dura; PoisonMs/Chance/Damage are the ps block's words - the poison
the field inflicts, outliving it.
================
*/
type SkillPoisonField struct {
	Pinned     bool
	DurationMs uint32
	ScanMs     uint32
	Radius     uint32
	PoisonMs   uint32
	Chance     uint32
	Damage     uint32
}

/*
==================
compilePoisonField

dura + puls + the kind-3 efr + the ps block + the caster's poison
getv words (RPDU/RPTU - KeyPoisonDamage/KeyPoisonDuration) + the
dagger reqi pair, on an untargeted prepared cast. Measured on every
tier of both catalogues (M8 s50): one uniform shape.
==================
*/
func compilePoisonField(fields []string, row SkillRow) (SkillPoisonField, bool) {
	if len(fields) != 118 || fields[0] != "1" || fields[8] != "2" || fields[68] != "3" ||
		!row.TimingPinned || !row.Consumption.Pinned || row.ChainSub || row.ChainNext != 0 ||
		row.Attack.Present || row.TargetRequired || row.ActionCastingTimeMs == 0 ||
		row.Consumption.HP != 0 || row.Consumption.HPPercent != 0 {
		return SkillPoisonField{}, false
	}
	// The extended lane's floor: the family's five cap-90 tiers author
	// the same shape and must stay refused.
	if textdataNonNegative(fields[skilldataColReqMasteryLv1]) < extendedPastCapMastery &&
		textdataNonNegative(fields[skilldataColReqMasteryLv2]) < extendedPastCapMastery {
		return SkillPoisonField{}, false
	}
	for _, column := range []int{15, 16, 17, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 56} {
		if fields[column] != "0" {
			return SkillPoisonField{}, false
		}
	}
	program, err := CompileSkillProgram(fields)
	if err != nil {
		return SkillPoisonField{}, false
	}
	var out SkillPoisonField
	seen := map[uint32]bool{}
	for i := 0; i < program.Len(); i++ {
		op := program.Instruction(i)
		if seen[op.Tag] && op.Tag != tagGetv && op.Tag != tagReqi {
			return SkillPoisonField{}, false
		}
		seen[op.Tag] = true
		switch op.Tag {
		case tagDura:
			if op.Count != 1 || op.Arguments[0] == 0 || op.Arguments[0] != row.EffectDurationMs {
				return SkillPoisonField{}, false
			}
			out.DurationMs = op.Arguments[0]
		case poisonFieldPulse:
			if op.Count != 1 || op.Arguments[0] == 0 {
				return SkillPoisonField{}, false
			}
			out.ScanMs = op.Arguments[0]
		case tagEfr:
			a := op.Arguments
			if a[0] != 3 || a[1] != 1 || a[2] == 0 || a[2] > 0xffff || a[3] == 0 || a[4] != 0 || a[5] != 24 {
				return SkillPoisonField{}, false
			}
			out.Radius = a[2]
		case poisonFieldStatus:
			if op.Count != 3 || op.Arguments[0] == 0 || op.Arguments[1] == 0 || op.Arguments[1] > 100 || op.Arguments[2] == 0 {
				return SkillPoisonField{}, false
			}
			out.PoisonMs, out.Chance, out.Damage = op.Arguments[0], op.Arguments[1], op.Arguments[2]
		case tagGetv:
			// RPDU/RPTU: the caster's poison damage and duration words -
			// KeyPoisonDamage/KeyPoisonDuration, the abnormal roll's own
			// constants; the plant reads them through the caster's stats.
			if op.Arguments[0] != abnormal.KeyPoisonDamage && op.Arguments[0] != abnormal.KeyPoisonDuration {
				return SkillPoisonField{}, false
			}
		case tagReqi:
		default:
			return SkillPoisonField{}, false
		}
	}
	if !seen[tagDura] || !seen[poisonFieldPulse] || !seen[tagEfr] || !seen[poisonFieldStatus] ||
		out.DurationMs == 0 || out.ScanMs == 0 || out.Radius == 0 {
		return SkillPoisonField{}, false
	}
	out.Pinned = true
	return out, true
}
