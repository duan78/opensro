/*
===========================================================================

extended_cap_test.go - the injected cap owns the freeze point

The native walk freezes just below level 90's requirement; the extended
content (isro-live-2026, port-only, not native) wires 140 with its curve
and the same freeze moves with it. Zero on the Runtime keeps the native
constant, so a native process never sees a different ceiling.

===========================================================================
*/
package progression

import (
	"testing"

	"opensro.online/server/internal/game/enterworld"
)

// stubLevels answers flat requirements from a map, good enough to watch
// the walk's cap behaviour without the shipped curve.
type stubLevels map[int64]int64

func (s stubLevels) SkillPointCost(level int64) (int64, bool) {
	return 1, true
}

func (s stubLevels) ExpRequired(level int64) (int64, bool) {
	req, ok := s[level]
	return req, ok
}

func (s stubLevels) MonsterExpBasis(level int64) (int64, bool) {
	return 24, true
}

var capStub = stubLevels{}

func init() {
	for level := int64(1); level <= 150; level++ {
		capStub[level] = 1000
	}
}

func TestNativeWalkFreezesBelowNinety(t *testing.T) {
	rt := &Runtime{}
	if got := rt.effectiveLevelCap(); got != 90 {
		t.Fatalf("zero injection keeps the native cap, got %d", got)
	}
	walk := walkExpCurve(capStub, 90, 999, 5_000_000, rt.effectiveLevelCap())
	if walk.ok != true || walk.level != 90 || walk.exp != 999 || walk.levelsGained != 0 {
		t.Fatalf("native freeze: level=%d exp=%d gained=%d ok=%v", walk.level, walk.exp, walk.levelsGained, walk.ok)
	}
}

func TestExtendedCapWalksPastNinetyAndFreezesAtOneForty(t *testing.T) {
	rt := &Runtime{LevelCap: 140}
	if got := rt.effectiveLevelCap(); got != 140 {
		t.Fatalf("injected cap wins, got %d", got)
	}
	walk := walkExpCurve(capStub, 90, 999, 5_000_000, rt.effectiveLevelCap())
	if !walk.ok || walk.level != 140 {
		t.Fatalf("extended walk reaches the cap, got level=%d ok=%v", walk.level, walk.ok)
	}
	if walk.levelsGained != 50 {
		t.Fatalf("90 -> 140 crosses fifty boundaries, gained %d", walk.levelsGained)
	}
	if walk.exp != 999 {
		t.Fatalf("the freeze holds requirement-1 at the injected cap too, exp=%d", walk.exp)
	}
}

func TestExtendedCapRefusesAboveItsCurve(t *testing.T) {
	// A curve that stops one row short of the injected cap must refuse
	// the crossing grant - fail closed, exactly like the native posture
	// on a missing row: the walk needs required(140) after the first
	// crossing and the whole grant refuses when it is absent.
	short := stubLevels{}
	for level := int64(1); level <= 139; level++ {
		short[level] = 1000
	}
	walk := walkExpCurve(short, 139, 500, 10_000, 140)
	if walk.ok || walk.levelsGained != 0 {
		t.Fatalf("a missing curve row refuses the grant, got ok=%v gained=%d", walk.ok, walk.levelsGained)
	}
}

// The stub must satisfy the same interface the runtime feeds the walk.
var _ enterworld.LevelDataSource = stubLevels{}
