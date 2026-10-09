/*
===========================================================================

discordwave_test.go - the Bard's Discord Wave clears monsters' hostility

The shipped Discord Wave row is cast through HandleTargetInteract on the
Bard itself; the monsters chasing the Bard then run the real monster leg.
Native sends each victim one negative hate event sourced at the cast's
target (593D62..593E7B) through the ordinary ledger update (5473C0).

===========================================================================
*/

package action

import (
	"testing"

	"opensro.online/server/internal/game/combat"
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/world/monster"
	"opensro.online/server/internal/game/world/simulation"
)

const (
	// discordCode is Discord Wave's last tier, dtnt(21194,0) mwdt(850).
	discordCode = "SKILL_EU_BARD_FORGETA_AGGRO_A_11"
	// discordFlat is that tier's dtnt flat word.
	discordFlat = 21194
	// discordMaxTargets and discordRadius are its efr words.
	discordMaxTargets = 4
	discordRadius     = 100
	// discordFarOffset puts a monster beyond the radius plus both body
	// radii from the Bard.
	discordFarOffset = 200
	// discordSmallHostility is less than any cut; discordLargeHostility
	// is more than this tier's cut.
	discordSmallHostility = 100
	discordLargeHostility = 1000000
)

/*
================
TestDiscordWaveReleasesUpToFourMonstersAroundItsTarget

Six monsters chase the Bard: five within 100 of it, one 200 away. The
first holds more hostility than the cut. Discord Wave on the Bard cuts
exactly four of the five near ones: the first keeps less hostility, the next
three are left with none. 5473C0 clamps at zero and, when the primary record
reaches it, refuses the callback, so each keeps the target it chases until
another opponent's hate takes over. The fifth near one and the far one are
untouched.
================
*/
func TestDiscordWaveReleasesUpToFourMonstersAroundItsTarget(t *testing.T) {
	rt, clock, c, original := newCombatTestRuntime(t, 1000)
	skill := shippedOffense(t, discordCode)
	if !skill.Threat.Decrease || skill.Threat.DecreaseFlat != discordFlat || skill.Threat.Area.MaxTargets != discordMaxTargets ||
		skill.Threat.Area.Radius != discordRadius || skill.Threat.Area.Shape != 2 || skill.Threat.Area.Select != 16 || !skill.Targets.Self {
		t.Fatalf("catalog shape: %+v refusal %q", skill.Threat, skill.OffenseRefusal)
	}
	skills := rt.deps.SkillData().(staticSkillSource)
	attack := skills[temptationAttackSkill]
	attack.Attack.Min, attack.Attack.Max, attack.Attack.Percent = 1, 1, 100
	skills[temptationAttackSkill] = attack
	ref := original.Ref
	ref.DefaultSkillIDs[0], ref.RunSpeed, ref.WalkSpeed, ref.ScaleDenom = temptationAttackSkill, 22, 8, 100
	state := simulation.NewMonsterState(monster.TemplateFromParts(map[uint32]monster.MonsterRef{ref.RefObjID: ref}, nil))
	state.SetTimeSource(clock.Now)
	state.SetRandomSource(func() float64 { return 0 })
	state.SetAbnormalContext(monsterAbnormalContext{rt})
	rt.Monsters = state
	gid := enterworld.ObjectIDForCharacter(c)
	at := monster.Pose{RegionID: original.Spawn.RegionID, X: original.Spawn.X, Y: original.Spawn.Y, Z: original.Spawn.Z}
	var mobs []monster.Instance
	for i, offset := range []float64{10, 20, 30, 40, 50, discordFarOffset} {
		pose := at
		pose.X += offset
		mob, err := state.DevelopmentCreateLeader(testDivision, ref.RefObjID, pose, clock.NowMs()+100000)
		if err != nil {
			t.Fatal(err)
		}
		if !state.ArmRetaliation(testDivision, mob.Gid, gid) {
			t.Fatal("no retaliation")
		}
		hostility := int32(discordSmallHostility)
		if i == 0 {
			hostility = discordLargeHostility
		}
		rt.commitAggression(testDivision, mob.Gid, simulation.HostilityEvent{Attacker: gid, Aggression: hostility}, clock.NowMs())
		mobs = append(mobs, mob)
	}
	rt.Worlds.Update(simulation.WorldKey(testDivision, c.Name), func() simulation.WorldState { return simulation.SeedWorldState(c) },
		func(w *simulation.WorldState) {
			w.Spawn = simulation.Spawn{RegionID: at.RegionID, X: at.X, Y: at.Y, Z: at.Z}
			w.SpawnSet = true
		})
	*c.World.Spawn.X = at.X
	hostility := func(mob monster.Instance) int32 {
		live, _ := rt.Monsters.Get(testDivision, mob.Gid)
		for _, record := range live.Opponents {
			if record.GID == gid {
				return record.Aggression
			}
		}
		return 0
	}
	before := make([]int32, len(mobs))
	for i, mob := range mobs {
		before[i] = hostility(mob)
		if mover, _ := rt.Monsters.Mover(testDivision, mob.Gid); mover.TargetGID() != gid || before[i] <= 0 {
			t.Fatalf("monster %d does not fight the Bard: hostility %d, mover %+v", i, before[i], mover)
		}
	}

	rt.deps.SkillData().(staticSkillSource)[skill.ID] = skill
	c.RaceIndex = testInt64(enterworld.RaceEurope)
	c.ModelCodename = "CHAR_EU_MAN_NOBLE"
	c.Skills = []uint32{skill.ID}
	c.Intellect = testInt64(2000)
	c.CurrentMP = testInt64(100000)
	weapon := rt.deps.ItemReferences().(staticItemSource)[c.MissionInventory[0].Codename]
	weapon.TypeIDs[3] = bardHarpKind
	c.MissionInventory[0].TypeFlags = weapon.TypeFlags()
	mp := enterworld.CurrentMP(c)
	out := rt.HandleTargetInteract(testDivision, c, wire.SkillAction{ActionId: skill.ID}.Encode())
	if frame, ok := findFrame(out.Frames, wire.OpSkillCastResult); !ok || frame.Payload[0] != 1 {
		t.Fatalf("Discord Wave was refused: %q %+v", out.DiagnosticRefusal, out.Frames)
	}
	if enterworld.CurrentMP(c) >= mp {
		t.Fatalf("MP not charged: %d -> %d", mp, enterworld.CurrentMP(c))
	}

	// One planner step consumes the release.
	clock.Advance(temptationTick)
	lease, _ := rt.Monsters.ObjectPopulation(testDivision, mobs[0].Gid)
	viewer := simulation.SessionSnapshot{SessionID: "discord", DivisionID: testDivision, CharacterID: c.ID,
		Population: lease, WorldInstance: uint32(lease.ID), BodyRadius: 4, CombatEligible: true,
		World: simulation.SeedWorldState(c)}
	ops := &simulation.MonsterMoverOps{Monsters: state, Rand: func() float64 { return 0 }, AttackPlan: rt.MonsterAttackPlan, RunAction: rt.RunMonsterAction}
	ops.RunMonsterLeg(clock.NowMs(), []simulation.SessionSnapshot{viewer}, &summonTickPusher{})

	cut := 0
	for i, mob := range mobs {
		after := hostility(mob)
		mover, _ := rt.Monsters.Mover(testDivision, mob.Gid)
		keeps := mover.TargetGID() == gid
		if after < before[i] {
			cut++
		}
		switch {
		case i == 0:
			if after >= before[i] || after <= 0 || !keeps {
				t.Fatalf("the hostile monster: hostility %d -> %d, keeps the Bard %v", before[i], after, keeps)
			}
		case i < discordMaxTargets:
			if after != 0 || !keeps {
				t.Fatalf("monster %d: hostility %d -> %d, keeps the Bard %v", i, before[i], after, keeps)
			}
		default:
			if after != before[i] || !keeps {
				t.Fatalf("monster %d beyond the cap or the radius was touched: hostility %d -> %d, keeps the Bard %v", i, before[i], after, keeps)
			}
		}
	}
	if cut != discordMaxTargets {
		t.Fatalf("%d monsters lost hostility, want %d", cut, discordMaxTargets)
	}
}

/*
================
TestUntargetedThreatDecreaseCutsAroundTheCaster

Extended content (isro-live-2026), port-only, not v1.150-native: the
Warlock's caster-centred form (CONFUSIONA_AGGROLOW past 90). Three
monsters hold hostility toward the Warlock; two stand within the efr
radius, one beyond it. The prepared untargeted cast releases and cuts
the near ones' hostility toward the CASTER - the tooltip's "confused
monsters will reduce their hostility toward the caster" - leaving the
far one untouched. The envelope rides the shipped Fire Trap row (a
prepared untargeted Wizard cast); the cut words are the measured
AGGROLOW shape.
================
*/
func TestUntargetedThreatDecreaseCutsAroundTheCaster(t *testing.T) {
	rt, clock, c, original := newCombatTestRuntime(t, 1000)
	skill := shippedOffense(t, "SKILL_EU_WIZARD_FIREA_TRAP_A_01")
	skill.CombatTrap = enterworld.SkillCombatTrap{}
	skill.Threat = enterworld.SkillThreat{Decrease: true, DecreaseFlat: 20000,
		Area: enterworld.SkillOffensiveArea{Shape: 1, Radius: 300, MaxTargets: 8, Select: 16}}
	rt.deps.SkillData().(staticSkillSource)[skill.ID] = skill
	c.RaceIndex = testInt64(enterworld.RaceEurope)
	c.ModelCodename = "CHAR_EU_MAN_NOBLE"
	c.Skills = []uint32{skill.ID}
	c.Intellect = testInt64(2000)
	c.CurrentMP = testInt64(100000)
	weapon := rt.deps.ItemReferences().(staticItemSource)[c.MissionInventory[0].Codename]
	weapon.TypeIDs[3] = int64(skill.RequiredWeaponKinds[0])
	c.MissionInventory[0].TypeFlags = weapon.TypeFlags()
	gid := enterworld.ObjectIDForCharacter(c)
	at := monster.Pose{RegionID: original.Spawn.RegionID, X: original.Spawn.X, Y: original.Spawn.Y, Z: original.Spawn.Z}
	var mobs []monster.Instance
	for _, offset := range []float64{10, 20, 700} {
		pose := at
		pose.X += offset
		mob, err := rt.Monsters.DevelopmentCreateLeader(testDivision, original.Ref.RefObjID, pose, clock.NowMs()+100000)
		if err != nil {
			t.Fatal(err)
		}
		if !rt.Monsters.ArmRetaliation(testDivision, mob.Gid, gid) {
			t.Fatal("no retaliation")
		}
		rt.commitAggression(testDivision, mob.Gid, simulation.HostilityEvent{Attacker: gid, Aggression: discordLargeHostility}, clock.NowMs())
		mobs = append(mobs, mob)
	}
	rt.Worlds.Update(simulation.WorldKey(testDivision, c.Name), func() simulation.WorldState { return simulation.SeedWorldState(c) },
		func(w *simulation.WorldState) {
			w.Spawn = simulation.Spawn{RegionID: at.RegionID, X: at.X, Y: at.Y, Z: at.Z}
			w.SpawnSet = true
		})
	*c.World.Spawn.X = at.X
	hostility := func(mob monster.Instance) int32 {
		live, _ := rt.Monsters.Get(testDivision, mob.Gid)
		for _, record := range live.Opponents {
			if record.GID == gid {
				return record.Aggression
			}
		}
		return 0
	}
	mp := enterworld.CurrentMP(c)
	out := rt.HandleTargetInteract(testDivision, c, wire.SkillAction{ActionId: skill.ID}.Encode())
	if frame, ok := findFrame(out.Frames, wire.OpSkillCastResult); !ok || frame.Payload[0] != 1 {
		t.Fatalf("the untargeted cut was refused: %q %+v", out.DiagnosticRefusal, out.Frames)
	}
	release := clock.NowMs() + int64(skill.ActionCastingTimeMs) + 1
	if len(rt.advanceProjectileCasts(release)) == 0 {
		t.Fatal("the untargeted cut never released")
	}
	if enterworld.CurrentMP(c) >= mp {
		t.Fatalf("MP not charged: %d -> %d", mp, enterworld.CurrentMP(c))
	}
	for i, mob := range mobs {
		after := hostility(mob)
		if i < 2 && after != discordLargeHostility-20000 {
			t.Fatalf("near monster %d: hostility %d, want %d", i, after, discordLargeHostility-20000)
		}
		if i == 2 && after != discordLargeHostility {
			t.Fatalf("the monster beyond the radius was touched: hostility %d", after)
		}
	}
}

/*
================
TestRatioDebuffCutsTheTargetAndExpires

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s34): the timed hostile ratio cast installs the slotless writes on the
monster target - its hit rate drops to the remaining factor in the
live combat projection - files tant's aggression, and the abnormal
tick's own expiry restores the stat. The envelope rides the shipped
Fire Trap row; the cut words are the measured WATER_CANCEL B shape.
================
*/
func TestRatioDebuffCutsTheTargetAndExpires(t *testing.T) {
	rt, clock, c, original := newCombatTestRuntime(t, 1000)
	skill := shippedOffense(t, "SKILL_EU_WIZARD_FIREA_TRAP_A_01")
	skill.CombatTrap = enterworld.SkillCombatTrap{}
	skill.RatioDebuff = enterworld.SkillRatioDebuff{Pinned: true, DurationMs: 10000, HitDown: 96, ThreatFlat: 5831}
	rt.deps.SkillData().(staticSkillSource)[skill.ID] = skill
	c.RaceIndex = testInt64(enterworld.RaceEurope)
	c.ModelCodename = "CHAR_EU_MAN_NOBLE"
	c.Skills = []uint32{skill.ID}
	c.Intellect = testInt64(2000)
	c.CurrentMP = testInt64(100000)
	weapon := rt.deps.ItemReferences().(staticItemSource)[c.MissionInventory[0].Codename]
	weapon.TypeIDs[3] = int64(skill.RequiredWeaponKinds[0])
	c.MissionInventory[0].TypeFlags = weapon.TypeFlags()
	gid := enterworld.ObjectIDForCharacter(c)
	at := monster.Pose{RegionID: original.Spawn.RegionID, X: original.Spawn.X, Y: original.Spawn.Y, Z: original.Spawn.Z}
	mob, err := rt.Monsters.DevelopmentCreateLeader(testDivision, original.Ref.RefObjID, at, clock.NowMs()+100000)
	if err != nil {
		t.Fatal(err)
	}
	if !rt.Monsters.ArmRetaliation(testDivision, mob.Gid, gid) {
		t.Fatal("no retaliation")
	}
	rt.commitAggression(testDivision, mob.Gid, simulation.HostilityEvent{Attacker: gid, Aggression: 1000}, clock.NowMs())
	rt.Worlds.Update(simulation.WorldKey(testDivision, c.Name), func() simulation.WorldState { return simulation.SeedWorldState(c) },
		func(w *simulation.WorldState) {
			w.Spawn = simulation.Spawn{RegionID: at.RegionID, X: at.X, Y: at.Y, Z: at.Z}
			w.SpawnSet = true
		})
	*c.World.Spawn.X = at.X
	live, _ := rt.Monsters.Get(testDivision, mob.Gid)
	before, err := combat.MonsterInstanceStats(live)
	if err != nil || before.HitRate <= 0 {
		t.Fatalf("base stats: %+v err %v", before, err)
	}
	mp := enterworld.CurrentMP(c)
	out := rt.HandleTargetInteract(testDivision, c, wire.SkillAction{ActionId: skill.ID, HasTarget: true, TargetGid: mob.Gid}.Encode())
	if frame, ok := findFrame(out.Frames, wire.OpSkillCastResult); !ok || frame.Payload[0] != 1 {
		t.Fatalf("the ratio debuff was refused: %q %+v", out.DiagnosticRefusal, out.Frames)
	}
	if enterworld.CurrentMP(c) >= mp {
		t.Fatalf("MP not charged: %d -> %d", mp, enterworld.CurrentMP(c))
	}
	cut, _ := rt.Monsters.Get(testDivision, mob.Gid)
	if cut.Abnormal == nil {
		t.Fatal("no abnormal block installed")
	}
	stats, err := combat.MonsterInstanceStats(cut)
	if err != nil {
		t.Fatal(err)
	}
	if want := before.HitRate * 0.04; stats.HitRate > want+0.5 || stats.HitRate < want-0.5 {
		t.Fatalf("hit rate %v, want ~%v (4%% of %v)", stats.HitRate, want, before.HitRate)
	}
	after, _ := rt.Monsters.Get(testDivision, mob.Gid)
	for _, record := range after.Opponents {
		if record.GID == gid && record.Aggression != 1000+5831 {
			t.Fatalf("tant hostility %d, want %d", record.Aggression, 1000+5831)
		}
	}
	// The abnormal tick's own expiry (4A4390) restores the stat.
	deadline := clock.NowMs() + 15000
	expired := false
	for !expired {
		clock.Advance(temptationTick)
		rt.advanceMonsterAbnormals(clock.NowMs())
		restored, _ := rt.Monsters.Get(testDivision, mob.Gid)
		if restored.Abnormal == nil {
			final, err := combat.MonsterInstanceStats(restored)
			if err != nil {
				t.Fatal(err)
			}
			if final.HitRate < before.HitRate-0.5 || final.HitRate > before.HitRate+0.5 {
				t.Fatalf("hit rate after expiry %v, want ~%v", final.HitRate, before.HitRate)
			}
			expired = true
		}
		if clock.NowMs() > deadline {
			t.Fatal("the debuff never expired")
		}
	}
}
