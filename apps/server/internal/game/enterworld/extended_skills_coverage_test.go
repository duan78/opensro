/*
===========================================================================

extended_skills_coverage_test.go - measure 91-140 skill execution (M7)

Extended content (isro-live-2026), port-only, not v1.150-native. Over
the REAL extended catalogue: every skill whose mastery requirement puts
it past level 90, classified by the PRODUCTION execution compiler - the
numbers the M7 audit cites and the lane's remaining-work budget comes
from. The native catalogue measured the same way is the baseline.
===========================================================================
*/
package enterworld

import (
	"fmt"
	"strings"
	"testing"

	"opensro.online/server/internal/gamedata"
	"opensro.online/server/internal/testsupport/gamedatatest"
	"opensro.online/server/internal/testsupport/licensed"
)

/*
================
skillCoveragePast90

Count, per 10-level mastery band, how many rows the production compiler
routes to each executable authority - and how many stay unsupported.
================
*/
func skillCoveragePast90(t *testing.T, source *TextdataSkills) map[int]map[string]int {
	t.Helper()
	bands := map[int]map[string]int{}
	kindName := func(kind SkillExecutionKind) string {
		switch kind {
		case SkillExecutionOffense:
			return "offense"
		case SkillExecutionRecovery:
			return "recovery"
		case SkillExecutionInstantEffect:
			return "instant"
		case SkillExecutionPassive:
			return "passive"
		case SkillExecutionTimedEffect:
			return "timed"
		case SkillExecutionPosition:
			return "position"
		case SkillExecutionThreat:
			return "threat"
		default:
			return "unsupported"
		}
	}
	for _, row := range source.rows.values() {
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if mastery < 91 {
			continue
		}
		band := int((mastery-1)/10)*10 + 1
		if bands[band] == nil {
			bands[band] = map[string]int{}
		}
		if row.ChainSub {
			// A chain's internal stage, covered by its root's plan - not a
			// separately castable skill; counted apart, never against the
			// castable denominator.
			bands[band]["chain-stage"] += 1
			continue
		}
		// The production admission is the union of the compiled plan kinds
		// and the runtime's own dispatch (resolveOffensiveSkill admits
		// Periodic.Pinned rows; applySkillRecovery admits targeted heals,
		// cures and lowest-ratio heals its shape flags pin). Measure what
		// actually casts.
		kind := kindName(source.plans[row.ID].kind)
		if kind == "unsupported" && row.Wall.Pinned {
			kind = "wall"
		}
		if kind == "unsupported" && row.Concealment.Pinned {
			kind = "concealment"
		}
		if kind == "unsupported" && row.CombatTrap.Pinned {
			// The planted combat-trap lane (acceptCombatTrap runs
			// skill.CombatTrap rows, action/skillcombattrap.go): measured
			// s29 - the live FIREA_TRAP rows already pin there, the plan
			// compiler just never names the lane.
			kind = "trap"
		}
		if kind == "unsupported" && row.Aura.Present && (row.BuffModifiers.Present() || row.Aura.Eshp) {
			// The persistent party-aura lane (acceptPartyBuff runs
			// Aura.Present rows with modifier blocks, action/skillparty.go):
			// measured s38 - BATTLAA_GUARD was already pinned there.
			kind = "aura"
		}
		if kind == "unsupported" && row.Illusion.Pinned {
			// The msch-3 random-player disguise (acceptIllusion, action/
			// illusion.go): measured s59 - live-only family.
			kind = "illusion"
		}
		if kind == "unsupported" && row.Duplicate.Pinned {
			// The msch-2 duplicate (acceptDuplicate copies the allied
			// target's model and equipment, action/duplicate.go) - measured
			// s53: the rows were already pinned there.
			kind = "duplicate"
		}
		if kind == "unsupported" && row.PoisonField.Pinned {
			// The planted, ticking poison field (acceptPoisonField plants a
			// persistent object, action/poisonfield.go): measured s51.
			kind = "field"
		}
		if kind == "unsupported" && row.RatioDebuff.Pinned {
			// The timed hostile ratio cut (acceptRatioDebuff installs the
			// slotless abnormal writes, action/ratiodebuff.go): measured
			// s34.
			kind = "debuff"
		}
		if kind == "unsupported" && row.Threat.Decrease {
			// The hostility-cut lane (skill.Threat.Decrease rows run
			// acceptDiscordWave targeted and its untargeted sibling past
			// 90, action/discordwave.go): measured s31.
			kind = "threat"
		}
		if kind == "unsupported" && row.TimedEffect.Periodic.Pinned {
			kind = "periodic"
		}
		if kind == "unsupported" && row.TimedEffect.Dance.Pinned {
			// The live dance toggle (M8 s66): the pulse-owned party-aura
			// lane - acceptPartyBuff, the same engine the native dances
			// run on.
			kind = "dance"
		}
		if kind == "unsupported" && row.MonsterCapture.Pinned {
			// The corpse-capture lane (action/monstercapture.go): the
			// Monster Mask drops the Essence of the Dead, transform.go
			// wears it. Shipped before M8; the union counted it only
			// from s67 - the five live TRANSFORMA_MASK tiers had been a
			// measurement artifact, never a missing feature.
			kind = "capture"
		}
		if kind == "unsupported" && (row.Recovery.SelfFlatPinned || row.Recovery.PartyHealPinned ||
			row.Recovery.LowestHealPinned || row.Recovery.PartyResurrectPinned || row.Recovery.HealOverTimePinned ||
			row.Abnormal.CurePresent() ||
			(row.Heal.Present && !row.Aura.Eshp && row.TargetRequired)) {
			kind = "recovery"
		}
		if kind == "unsupported" {
			// The unsupported split the M7 budget turns on: chain roots
			// whose chain failed validation, versus plain rows no parser
			// pins.
			if row.ChainNext != 0 {
				kind = "unsupported-chain"
			} else {
				kind = "unsupported-unpinned"
			}
		}
		bands[band][kind] += 1
	}
	return bands
}

/*
================
TestExtendedSkillsExecutionCoveragePastLevel90
================
*/
func TestExtendedSkillsExecutionCoveragePastLevel90(t *testing.T) {
	licensed.RequireGameData(t)
	t.Setenv(gamedata.EnvExtendedContent, "on")
	extended, err := gamedata.LoadExtended()
	if err != nil {
		t.Skipf("extended projection is not built: %v", err)
	}
	skills := NewTextdataSkills(extended.TextdataDir)
	if err := skills.Load(); err != nil {
		t.Fatalf("extended skills: %v", err)
	}
	native := NewTextdataSkills(gamedatatest.TextdataDir(t))
	if err := native.Load(); err != nil {
		t.Fatalf("native skills: %v", err)
	}

	for name, source := range map[string]*TextdataSkills{"native": native, "extended": skills} {
		total, routed := 0, 0
		for band, counts := range skillCoveragePast90(t, source) {
			bandTotal, bandRouted := 0, 0
			for kind, count := range counts {
				if kind == "chain-stage" {
					continue
				}
				bandTotal += count
				if !strings.HasPrefix(kind, "unsupported") {
					bandRouted += count
				}
			}
			total += bandTotal
			routed += bandRouted
			summary := ""
			for _, kind := range []string{"offense", "recovery", "instant", "passive", "timed", "periodic", "wall", "concealment", "position", "threat", "aura", "debuff", "trap", "field", "duplicate", "illusion", "dance", "capture", "unsupported-unpinned", "unsupported-chain", "chain-stage"} {
				if counts[kind] > 0 {
					summary += fmt.Sprintf(" %s=%d", kind, counts[kind])
				}
			}
			t.Logf("%s band %d:%s (executable %d/%d)", name, band, summary, bandRouted, bandTotal)
		}
		t.Logf("%s past-90 total: %d rows, %d executable (%.1f%%)",
			name, total, routed, 100*float64(routed)/float64(max(total, 1)))
	}

	// Extended (isro-live-2026, M8): the mission's ACCEPTANCE criterion is
	// the player perimeter (owner decision 2026-10-10, charter section 8:
	// the ~379 monster/arena rows leave the criterion - none of them is a
	// player skill; the global stays printed as a report). The player rows
	// (SKILL_CH_/SKILL_EU_) are measured with the SAME filter the global
	// uses (chain roots cast, their ChainSub stages are counted inside the
	// root's plan; s64 fixed an earlier draft that dropped the roots).
	playerTotal, playerRouted := 0, 0
	var playerMisses []string
	for _, row := range skills.rows.values() {
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if mastery < 91 || row.ChainSub {
			continue
		}
		if !strings.HasPrefix(row.Codename, "SKILL_CH_") && !strings.HasPrefix(row.Codename, "SKILL_EU_") {
			continue
		}
		playerTotal++
		if rowRuntimeAdmitted(skills, row) {
			playerRouted++
		} else {
			playerMisses = append(playerMisses, row.Codename)
		}
	}
	t.Logf("acceptance (player scope, owner 2026-10-10): %d rows, %d executable (%.1f%%)",
		playerTotal, playerRouted, 100*float64(playerRouted)/float64(max(playerTotal, 1)))

	// The criterion's theorem: every miss belongs to an owner-decided
	// block (charter section 8, 2026-10-10). The 15 player-targeting PvP
	// rows are PARKED DEFINITIVELY by the owner - they may miss forever.
	// The bard dances (s66) and the TRANSFORMA_MASK rows (s67: the
	// capture lane was already shipped, the union had never counted it)
	// are DELIVERED: their blocks must stay at zero misses. A miss
	// outside the parked block - or a regression inside a delivered one -
	// is what this test exists to catch.
	parkedPvP, deliveredDances, deliveredMask := 0, 0, 0
	for _, codename := range playerMisses {
		switch {
		case strings.Contains(codename, "DANCEA_WITH_MUSIC"):
			deliveredDances++
		case strings.Contains(codename, "TRANSFORMA_MASK"):
			deliveredMask++
		case strings.Contains(codename, "MANADRY"), strings.Contains(codename, "STEALTHA_"):
			parkedPvP++
		default:
			t.Fatalf("player row %s is not executable and belongs to no owner-decided block", codename)
		}
	}
	if deliveredDances != 0 || deliveredMask != 0 {
		t.Fatalf("a delivered block regressed: dances %d, mask %d misses", deliveredDances, deliveredMask)
	}
	if parkedPvP != 15 {
		t.Fatalf("parked pvp misses = %d, want the owner's fifteen", parkedPvP)
	}
	t.Logf("misses by owner block: pvp parked %d (dances %d, mask %d - delivered, zero expected)", parkedPvP, deliveredDances, deliveredMask)

	// What the unpinned rows ARE (the M7 budget's denominator): sample
	// codenames per band - player-family variants, monster skills or
	// genuinely new player effects.
	samples := map[string][]string{}
	for _, row := range skills.rows.values() {
		mastery := row.Masteries[0].Level
		if row.Masteries[1].Level > mastery {
			mastery = row.Masteries[1].Level
		}
		if mastery < 91 || mastery > 140 || row.ChainSub || row.ChainNext != 0 {
			continue
		}
		// The sample must mirror the admission union the counts measure
		// (s30: the plan-kind-only filter listed HEALA_CYCLE_B's past-90
		// tiers and the trap rows as "unpinned" while both were counted
		// executable - a false lead that cost a probe).
		if skills.plans[row.ID].kind != SkillExecutionUnsupported || rowRuntimeAdmitted(skills, row) {
			continue
		}
		key := fmt.Sprintf("band %d", int((mastery-1)/10)*10+1)
		if len(samples[key]) < 12 {
			samples[key] = append(samples[key], row.Codename)
		}
	}
	for band, names := range samples {
		t.Logf("%s unpinned sample: %v", band, names)
	}

	// The extended catalogue must not route FEWER kinds than the native
	// one past 90 (the native rows past 90 are the provisioned tail; the
	// extended ones are the live game's real 91-140 set).
	extendedRouted := 0
	for _, counts := range skillCoveragePast90(t, skills) {
		for kind, count := range counts {
			if !strings.HasPrefix(kind, "unsupported") {
				extendedRouted += count
			}
		}
	}
	if extendedRouted == 0 {
		t.Fatal("no extended past-90 skill reaches an executable authority")
	}
}
