/*
===========================================================================

skillstatuscast_dmir_test.go - the DMIR tolerance's void proof, frozen

Extended content (isro-live-2026), port-only, not v1.150-native. The
status-cast lane tolerates the rebalance's DMIR getv key; the void
proof measured on 2026-10-09 (M8 s37) is that no native row passes the
contract through it - the 134 native carriers are the DoT family's own
tiers, refused by the contract's other gates. This test freezes that
fact: a native row admitting only because of the tolerance would be a
native behavior change.

===========================================================================
*/

package enterworld

import (
	"testing"

	"opensro.online/server/internal/testsupport/gamedatatest"
)

/*
================
TestDMIRToleranceAdmitsNoNativeStatusCast

The DMIR key is absent from the v1.150 data entirely: no shipped row's
program carries it, so the tolerance can never change a native
admission. The extended past-90 RAZEA tiers ride it instead.
================
*/
func TestDMIRToleranceAdmitsNoNativeStatusCast(t *testing.T) {
	native := sharedShippedSkills(t)
	cells := readExtendedSkillCells(t, gamedatatest.TextdataDir(t))
	carriers := 0 // kept for the extended side's count in a future assertion
	for _, row := range native.rows.values() {
		fields := cells[row.Codename]
		if len(fields) == 0 {
			continue
		}
		program, err := CompileSkillProgram(fields)
		if err != nil {
			continue
		}
		for i := 0; i < program.Len(); i++ {
			op := program.Instruction(i)
			if op.Tag == tagGetv && op.Arguments[0] == extendedDotScalingKey {
				t.Fatalf("%s: the v1.150 data carries the DMIR key - the tolerance's void proof is gone", row.Codename)
			}
		}
	}
	_ = carriers
}
