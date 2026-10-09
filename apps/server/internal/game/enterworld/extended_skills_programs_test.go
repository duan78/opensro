/*
===========================================================================

extended_skills_programs_test.go - the unsupported programs, decoded (M8)

Extended content (isro-live-2026), port-only, not v1.150-native. A
diagnostic: for each unsupported player family past 90, read the row's
raw skilldata cells and compile its encoded tail through the production
program compiler, printing the instruction tags. The clusters this
reveals are the M8 implementation budget's shape. Folded into the
coverage test once the lane lands.
===========================================================================
*/
package enterworld

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/licensed"
)

/*
================
TestSkillProbeUnsupportedPrograms

One exemplar per unsupported family, its compiled tag sequence and the
family's row count past 90.
================
*/
func TestSkillProbeUnsupportedPrograms(t *testing.T) {
	licensed.RequireGameData(t)
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	live := NewTextdataSkills(extended.TextdataDir)
	if err := live.Load(); err != nil {
		t.Fatal(err)
	}

	// The unsupported family exemplars (production plans), then their raw
	// cells from the shards.
	type family struct {
		exemplar string
		count    int
	}
	families := map[string]*family{}
	for _, row := range live.rows.values() {
		name := row.Codename
		if row.ChainSub || row.ChainNext != 0 {
			continue
		}
		isPlayer := strings.HasPrefix(name, "SKILL_CH") || strings.HasPrefix(name, "SKILL_EU")
		if !isPlayer {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if mastery < 91 || live.plans[row.ID].kind != SkillExecutionUnsupported {
			continue
		}
		body := familyKey(name)
		if families[body] == nil {
			families[body] = &family{exemplar: name}
		}
		families[body].count++
	}
	raw := readExtendedSkillCells(t, extended.TextdataDir)

	names := make([]string, 0, len(families))
	for name := range families {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return families[names[i]].count > families[names[j]].count })
	for _, name := range names {
		if name == "" || families[name].count < 3 {
			continue
		}
		cells, ok := raw[families[name].exemplar]
		if !ok {
			t.Logf("%-40s n=%-3d %s -> no raw row", name, families[name].count, families[name].exemplar)
			continue
		}
		program, err := CompileSkillProgram(cells)
		if err != nil {
			t.Logf("%-40s n=%-3d %s -> COMPILE ERROR %v", name, families[name].count, families[name].exemplar, err)
			continue
		}
		tags := ""
		for i := 0; i < program.Len(); i++ {
			tag := program.Instruction(i).Tag
			tags += fmt.Sprintf(" %q", string([]byte{byte(tag), byte(tag >> 8), byte(tag >> 16), byte(tag >> 24)}))
		}
		t.Logf("%-40s n=%-3d %s ->%s", name, families[name].count, families[name].exemplar, tags)
	}
}

/*
================
familyKey

SKILL_EU_CLERIC_HEALA_TARGET_B_11 -> CLERIC_HEALA_TARGET.
================
*/
func familyKey(name string) string {
	body := name
	for _, prefix := range []string{"SKILL_CH_", "SKILL_EU_"} {
		if strings.HasPrefix(body, prefix) {
			body = body[len(prefix):]
			break
		}
	}
	for _, cut := range []string{"_A_", "_B_", "_C_", "_D_", "_E_"} {
		if i := strings.Index(body, cut); i > 0 {
			return body[:i]
		}
	}
	return body
}

/*
================
readExtendedSkillCells

codename -> the row's 118 raw cells, read straight from the shards.
================
*/
func readExtendedSkillCells(t *testing.T, dir string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "skilldata_") || !strings.HasSuffix(name, ".txt") {
			continue
		}
		for _, cells := range ReadTextdataFile(filepath.Join(dir, name)) {
			if len(cells) == 118 && len(cells[3]) > 9 {
				out[cells[3]] = cells
			}
		}
	}
	return out
}
