package enterworld

import "fmt"

// SkillInstruction is one completely decoded native parameter block. Arguments
// preserve the original 32-bit words, including signed values' bit patterns.
type SkillInstruction struct {
	Tag       uint32
	Column    uint8
	Count     uint8
	Arguments [6]uint32
}

// SkillProgram owns the validated stream. Callers receive values, never its
// backing storage. Decoding does not confer execution or lifecycle support.
type SkillProgram struct{ instructions []SkillInstruction }

func (p SkillProgram) Len() int                           { return len(p.instructions) }
func (p SkillProgram) Instruction(i int) SkillInstruction { return p.instructions[i] }

// CompileSkillProgram is shared by production admission and coverage export.
// Native indexers skip zero words; ssou stops indexing. Unknown instructions
// cannot be treated as zero-argument operations in an executable plan.
func CompileSkillProgram(fields []string) (SkillProgram, error) {
	var p SkillProgram
	if len(fields) != 118 {
		return p, fmt.Errorf("skill program: expected 118 fields, got %d", len(fields))
	}
	for col := skilldataColEncodedTail; col < len(fields); {
		n, ok := textdataInt(fields[col])
		if !ok || n < 0 || n > 0xffffffff {
			return SkillProgram{}, fmt.Errorf("skill program: invalid tag at %d", col)
		}
		if n == 0 {
			col++
			continue
		}
		arity, known := spawnParamSpec(uint32(n))
		// Extended content (isro-live-2026), port-only, not v1.150-native:
		// the three one-word instructions the live client adds and the
		// v1.150 engine has no reader for (spawnParamSpec is native
		// hash-pinned evidence, so the riders are admitted here instead).
		// Measured on the live rows (2026-10-09): psog {1|2} rides every
		// 2026 attack and defense row, repl {1} and srpc {skillId} ride the
		// bard dance programs (srpc's word is a skill id, 9932..). The
		// v1.150 engine ignores them; the admission walks tolerate the
		// riders so the row's native shape decides executability.
		if !known && (n == 0x70736f67 || n == 0x7265706c || n == 0x72706373) {
			arity, known = 1, true
		}
		// Extended content (isro-live-2026), port-only, not v1.150-native
		// (M8 s63): hwdu - the live "Aura of Blood" rows' word (the
		// HWAN-duration family; absent from the v1.150 data). Arity one.
		if !known && n == 0x68776475 {
			arity, known = 1, true
		}
		// Extended content (isro-live-2026), port-only, not v1.150-native
		// (M8 s66): ycdc - the live "Dance with Music" programs' header
		// word, absent from the v1.150 data and from spawnParamSpec's
		// hash-pinned table. Zero arguments: the walk consistency of every
		// live dance tier (ycdc then scls {n}) leaves no other reading.
		if !known && n == 0x79636463 {
			arity, known = 0, true
		}
		if !known {
			return SkillProgram{}, fmt.Errorf("skill program: unknown instruction %x at %d", n, col)
		}
		if arity > 6 || col+arity >= len(fields) {
			return SkillProgram{}, fmt.Errorf("skill program: truncated instruction %x at %d", n, col)
		}
		i := SkillInstruction{Tag: uint32(n), Column: uint8(col), Count: uint8(arity)}
		for a := 0; a < arity; a++ {
			v, valid := textdataDword(fields[col+1+a])
			if !valid {
				return SkillProgram{}, fmt.Errorf("skill program: invalid argument at %d", col+1+a)
			}
			i.Arguments[a] = v
		}
		p.instructions = append(p.instructions, i)
		if i.Tag == 0x73736f75 {
			break
		}
		col += 1 + arity
	}
	return p, nil
}
