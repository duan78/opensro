/*
===========================================================================

extended_skills_measurements_test.go - the skill execution baselines (M7)

Extended content (isro-live-2026), port-only, not v1.150-native. The M7
measurements the audit cites: the overall and player-family parity
between the native and live catalogues, and the classification of the
unsupported 91-140 player families (the remaining-work budget). The
per-band coverage lives in extended_skills_coverage_test.go.
===========================================================================
*/
package enterworld

import (
	"sort"
	"strings"
	"testing"

	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
	"opensro.online/server/internal/testsupport/licensed"
)

func TestSkillProbeOverallParity(t *testing.T) {
	licensed.RequireGameData(t)
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	native := NewTextdataSkills(gamedatatest.TextdataDir(t))
	if err := native.Load(); err != nil {
		t.Fatal(err)
	}
	live := NewTextdataSkills(extended.TextdataDir)
	if err := live.Load(); err != nil {
		t.Fatal(err)
	}
	for _, which := range []struct {
		name string
		src  *TextdataSkills
	}{{"native", native}, {"live", live}} {
		castable, routed := 0, 0
		for _, row := range which.src.rows.values() {
			if row.ChainSub {
				continue
			}
			castable++
			if rowRuntimeAdmitted(which.src, row) {
				routed++
			}
		}
		t.Logf("%s ALL castable rows: %d, executable %d (%.1f%%)",
			which.name, castable, routed, 100*float64(routed)/float64(castable))
		// Player families only (the SKILL_CH/EU prefix), the way a player
		// experiences coverage.
		pCastable, pRouted := 0, 0
		for _, row := range which.src.rows.values() {
			if row.ChainSub {
				continue
			}
			name := row.Codename
			isPlayer := len(name) > 8 && name[:8] == "SKILL_CH" || len(name) > 8 && name[:8] == "SKILL_EU"
			if !isPlayer {
				continue
			}
			pCastable++
			if rowRuntimeAdmitted(which.src, row) {
				pRouted++
			}
		}
		t.Logf("%s PLAYER (SKILL_CH/EU) castable: %d, executable %d (%.1f%%)",
			which.name, pCastable, pRouted, 100*float64(pRouted)/float64(pCastable))
	}
}

func TestSkillProbeUnsupportedPlayerFamilies(t *testing.T) {
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
	families := map[string]int{}
	for _, row := range live.rows.values() {
		name := row.Codename
		if row.ChainSub || row.ChainNext != 0 {
			continue
		}
		isPlayer := len(name) > 8 && (name[:8] == "SKILL_CH" || name[:8] == "SKILL_EU")
		if !isPlayer {
			continue
		}
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if mastery < 91 || rowRuntimeAdmitted(live, row) {
			continue
		}
		// SKILL_EU_CLERIC_HEALA_TARGET_B_11 -> CLERIC_HEALA_TARGET
		body := name[8:]
		for _, cut := range []string{"_A_", "_B_", "_C_", "_D_", "_E_"} {
			if i := strings.Index(body, cut); i > 0 {
				body = body[:i]
				break
			}
		}
		families[body] += 1
	}
	type familyCount struct {
		name  string
		count int
	}
	list := make([]familyCount, 0, len(families))
	for name, count := range families {
		list = append(list, familyCount{name, count})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].count > list[j].count })
	for i, fc := range list {
		if i >= 14 {
			break
		}
		t.Logf("unsupported family %-34s %d rows", fc.name, fc.count)
	}
	t.Logf("unsupported player families past 90: %d families, %d rows", len(list), func() int {
		n := 0
		for _, c := range families {
			n += c
		}
		return n
	}())
}

/*
================
rowRuntimeAdmitted

The runtime admission union the coverage test measures: compiled plan
kinds, the wall lane (the cast dispatch admits Wall.Pinned rows
directly, action/skillwall.go), the planted combat-trap lane
(acceptCombatTrap runs skill.CombatTrap rows, action/skillcombattrap.go
- measured s29: the live FIREA_TRAP rows already pin there), the
hostility-cut lane (skill.Threat.Decrease rows run acceptDiscordWave
and its untargeted sibling, action/discordwave.go - measured s31), the
timed hostile ratio cut (RatioDebuff.Pinned rows run acceptRatioDebuff,
action/ratiodebuff.go - measured s34), the persistent party-aura lane
(Aura.Present rows with modifier blocks run acceptPartyBuff,
action/skillparty.go - measured s38: the live BATTLAA_GUARD rows were
already pinned there), the offensive periodic path, and the recovery
dispatch's shape flags. Keep the two tests measuring the
same predicate.
================
*/
func rowRuntimeAdmitted(source *TextdataSkills, row SkillRow) bool {
	return source.plans[row.ID].kind != SkillExecutionUnsupported ||
		row.Wall.Pinned ||
		row.Concealment.Pinned ||
		row.CombatTrap.Pinned ||
		row.RatioDebuff.Pinned ||
		row.PoisonField.Pinned ||
		row.Duplicate.Pinned ||
		row.Illusion.Pinned ||
		row.Threat.Decrease ||
		row.Aura.Present && (row.BuffModifiers.Present() || row.Aura.Eshp) ||
		row.TimedEffect.Periodic.Pinned ||
		row.TimedEffect.Dance.Pinned ||
		row.Recovery.SelfFlatPinned || row.Recovery.PartyHealPinned ||
		row.Recovery.LowestHealPinned || row.Recovery.PartyResurrectPinned ||
		row.Recovery.HealOverTimePinned || row.Abnormal.CurePresent() ||
		(row.Heal.Present && !row.Aura.Eshp && row.TargetRequired)
}
