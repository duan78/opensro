/*
===========================================================================

registry_test.go - quest-trap scan, lifetime and target-isolation regressions

Drive time explicitly. Boundaries come from 48CEA0/48D690, including the
strict expiry and radius comparisons that ordinary timer helpers obscure.

===========================================================================
*/
package skillobject

import (
	"testing"

	"opensro.online/server/internal/game/item/wire"
	worldgeom "opensro.online/server/internal/game/world"
	"opensro.online/server/internal/game/world/instance"
)

/*
================
trapFixture
================
*/
func trapFixture(t *testing.T, registry *Registry) Object {
	t.Helper()
	object, err := registry.Create(Object{
		Division: "trap-test", Population: instance.Lease{ID: instance.Pack(1, 1), Generation: 1},
		OwnerGID: 100001, OwnerName: "Trapper", CreatedMs: 1000,
		Program: Program{SkillID: 7108, DurationMs: 600, ScanMs: 300, Radius: 20, Targets: [3]uint32{5867}},
		Spawn:   wire.SkillObjectSpawn{Region: 0x6454, X: 100, Y: 10, Z: 200},
	})
	if err != nil {
		t.Fatal(err)
	}
	return object
}

/*
================
TestTrapScansOnceAndRetiresBeforeQuestOutcome

A failed quest handler cannot leave the trap armed for repeated rolls.
================
*/
func TestTrapScansOnceAndRetiresBeforeQuestOutcome(t *testing.T) {
	var registry Registry
	object := trapFixture(t, &registry)
	target := Target{GID: 400001, RefID: 5867, Alive: true, Region: 0x6454, X: 100, Y: 10, Z: 200}
	if _, hit, retired := registry.Scan(object.Spawn.GID, 1299, true, []Target{target}); hit != 0 || retired {
		t.Fatal("trap scanned before its first native pulse")
	}
	if _, hit, retired := registry.Scan(object.Spawn.GID, 1300, true, []Target{target}); hit != target.GID || !retired {
		t.Fatalf("due scan did not capture and retire: %d %v", hit, retired)
	}
	if _, hit, retired := registry.Scan(object.Spawn.GID, 1600, true, []Target{target}); hit != 0 || retired {
		t.Fatal("retired trap replayed its capture event")
	}
	if len(registry.Snapshot()) != 0 {
		t.Fatal("retired trap survived in world publication")
	}
}

/*
================
TestTrapExpiryIsStrictAndOwnerLossIsImmediate
================
*/
func TestTrapExpiryIsStrictAndOwnerLossIsImmediate(t *testing.T) {
	var registry Registry
	object := trapFixture(t, &registry)
	if _, _, retired := registry.Scan(object.Spawn.GID, 1600, true, nil); retired {
		t.Fatal("trap expired at equality instead of elapsed > duration")
	}
	if _, _, retired := registry.Scan(object.Spawn.GID, 1601, true, nil); !retired {
		t.Fatal("trap survived its duration")
	}
	second := trapFixture(t, &registry)
	if second.Spawn.GID == object.Spawn.GID {
		t.Fatal("object identity reused")
	}
	if _, _, retired := registry.Scan(second.Spawn.GID, 1001, false, nil); !retired {
		t.Fatal("owner loss waited for the next scan")
	}
}

/*
================
TestTrapTargetFilterUsesOwnershipThreeDimensionsAndTerminatedList
================
*/
func TestTrapTargetFilterUsesOwnershipThreeDimensionsAndTerminatedList(t *testing.T) {
	var registry Registry
	object := trapFixture(t, &registry)
	base := Target{GID: 400001, RefID: 5867, Alive: true, Region: 0x6454, X: 100, Y: 10, Z: 200}
	for _, test := range []struct {
		name   string
		change func(*Target)
		want   bool
	}{
		{"unowned", func(*Target) {}, true},
		{"same owner", func(target *Target) { target.OwnerGID = object.OwnerGID }, true},
		{"foreign owner", func(target *Target) { target.OwnerGID = object.OwnerGID + 1 }, false},
		{"dead", func(target *Target) { target.Alive = false }, false},
		{"wrong model", func(target *Target) { target.RefID++ }, false},
		{"radius equality", func(target *Target) { target.Y += 20 }, false},
		{"radius inside", func(target *Target) { target.Y += 19 }, true},
		{"other plane", func(target *Target) { target.Region |= 0x8000 }, false},
		{"neighbor region", func(target *Target) { target.Region++; target.X -= 1920 }, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := base
			test.change(&target)
			if got := Matches(object, target); got != test.want {
				t.Fatalf("match=%v want=%v", got, test.want)
			}
		})
	}
	object.Program.Targets = [3]uint32{0, base.RefID}
	if Matches(object, base) {
		t.Fatal("target after the zero terminator matched")
	}
}

/*
================
TestCombatTrapMatchesAnyLivingVictimAndReplicatesToObservers

A combat trap ignores the quest target list and the victim's current
opponent; concealment changes presentation without removing replication.
================
*/
func TestCombatTrapMatchesAnyLivingVictimAndReplicatesToObservers(t *testing.T) {
	var registry Registry
	object := trapFixture(t, &registry)
	object.Program.Combat, object.Program.Hidden = true, true
	victim := Target{GID: 7, RefID: 1, OwnerGID: 999, Alive: true, Region: object.Spawn.Region, X: 105, Y: 10, Z: 200}
	if !Matches(object, victim) {
		t.Fatal("combat trap refused a living monster fighting someone else")
	}
	victim.Alive = false
	if Matches(object, victim) {
		t.Fatal("combat trap matched a corpse")
	}
	viewer := Viewer{Division: object.Division, Population: object.Population,
		Position: worldgeom.RegionXZ{RegionID: object.Spawn.Region, X: 100, Z: 200}}
	if !Visible(object, viewer) {
		t.Fatal("concealed trap missing from observer replication")
	}
	viewer.CharacterGID = object.OwnerGID
	if !Visible(object, viewer) {
		t.Fatal("hidden trap hidden from its owner")
	}
}

/*
================
TestPersistFieldScansAllAndRetiresOnExpiry

Extended content (isro-live-2026), port-only, not v1.150-native (M8
s51): a persisting field reports every matching monster at each due
scan and stays alive until its own duration passes.
================
*/
func TestPersistFieldScansAllAndRetiresOnExpiry(t *testing.T) {
	r := &Registry{}
	r.nextGID = 900
	object, err := r.Create(Object{Division: "d", OwnerGID: 7, OwnerName: "rog", Population: instance.Lease{ID: 4},
		Program: Program{SkillID: 1, DurationMs: 30000, ScanMs: 3000, Radius: 100, Combat: true, Persist: true},
		Spawn:   wire.SkillObjectSpawn{Region: 1, X: 0, Y: 0, Z: 0}})
	if err != nil {
		t.Fatal(err)
	}
	now := int64(3000)
	targets := []Target{{GID: 11, RefID: 1, Alive: true, Region: 1}, {GID: 12, RefID: 1, Alive: true, Region: 1, X: 500}}
	t.Logf("object gid=%d spawn=%+v matches11=%v matches12=%v", object.Spawn.GID, object.Spawn, Matches(object, targets[0]), Matches(object, targets[1]))
	if matched, retired := r.ScanPersist(object.Spawn.GID, now, true, targets); retired || len(matched) != 1 || matched[0] != 11 {
		t.Fatalf("first scan: matched %v retired %v", matched, retired)
	}
	// Not due yet: the next scan window waits.
	if matched, retired := r.ScanPersist(object.Spawn.GID, now+1000, true, targets); retired || matched != nil {
		t.Fatalf("early scan: matched %v retired %v", matched, retired)
	}
	if matched, retired := r.ScanPersist(object.Spawn.GID, now+3000, true, targets); retired || len(matched) != 1 {
		t.Fatalf("second scan: matched %v retired %v", matched, retired)
	}
	if _, retired := r.ScanPersist(object.Spawn.GID, now+31000, true, targets); !retired {
		t.Fatal("the field outlived its duration")
	}
}
