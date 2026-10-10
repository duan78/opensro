/*
===========================================================================

skillillusion.go - the Warlock's Illusion

msch 3 {level} dura cks tant reqi: the Warlock randomly takes the look
of a character no higher than the level word ("Randomly transforms
yourself into a character lower level than you"). The instance lives on
the caster exactly as the Duplicate's does (transform.go); the level
word is the pick's ceiling. The vanilla client varies the model from
viewer to viewer; this port keeps one model per cast for every viewer
- the recorded v1 boundary (a per-connection replication lane does not
exist in the v1.150 wire this port owns).

Extended content (isro-live-2026), port-only, not v1.150-native: the
family exists ONLY in the live catalogue (measured: zero v1.150 rows
carry msch mode 3), so no mastery floor is needed - no native row can
ever reach this parser.

===========================================================================
*/

package enterworld

// SkillIllusion is the executable msch-3 program.
type SkillIllusion struct {
	Pinned   bool
	MaxLevel uint32
}

func parseSkillIllusion(fields []string, row *SkillRow) {
	gate := row.CastGate
	if !gate.MschPresent || gate.MschMode != 3 || gate.MschLevel == 0 {
		return
	}
	if !row.ReplacementPinned || !row.TimingPinned || !row.ActionCastingTimePinned || row.ActionCastingTimeMs != 0 ||
		!row.Consumption.Pinned || row.ChainNext != 0 || row.ActionHandler != SkillActionPersistent {
		return
	}
	if row.TargetRequired {
		return
	}
	program, err := CompileSkillProgram(fields)
	if err != nil {
		return
	}
	duration := false
	for i := 0; i < program.Len(); i++ {
		switch op := program.Instruction(i); op.Tag {
		case 0x6d736368: // msch
		case tagDura:
			duration = op.Arguments[0] != 0
		case tagSkc, tagTimedStunOverride, tagStatusThreat, tagReqi:
		default:
			return
		}
	}
	if !duration {
		return
	}
	row.Illusion = SkillIllusion{Pinned: true, MaxLevel: gate.MschLevel}
}
