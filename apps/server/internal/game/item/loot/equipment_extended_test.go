/*
===========================================================================

equipment_extended_test.go - the live-2026 supplement, both settings

Extended content (isro-live-2026), port-only, not v1.150-native. The
installed supplement puts the degree 10-12 wearables into the ordinary
drop path at their real levels; the shared process catalogue (what a
flag-off process serves) never sees them.
===========================================================================
*/
package loot

import (
	"sort"
	"strings"
	"testing"
)

/*
================
TestExtendedEquipmentSupplementInstallsItsWindows
================
*/
func TestExtendedEquipmentSupplementInstallsItsWindows(t *testing.T) {
	c := loadEquipmentCatalog()
	if err := installEquipmentExtended(&c); err != nil {
		t.Fatal(err)
	}

	// The items: every degree 10-12 bucket carries its rows, native and
	// extended never mixed.
	extendedRows := 0
	for key, bucket := range c.buckets {
		for _, ref := range bucket.refs {
			if !strings.Contains(ref.Codename, "_10_") && !strings.Contains(ref.Codename, "_11_") &&
				!strings.Contains(ref.Codename, "_12_") {
				continue
			}
			extendedRows++
			if key.group != ref.Group || key.rare != ref.Rare || key.country != ref.Country {
				t.Fatalf("extended row keyed wrong: %+v under %+v", ref, key)
			}
		}
	}
	if extendedRows != 308 {
		t.Fatalf("extended rows catalogued = %d, want 308", extendedRows)
	}

	// The windows: the real requirement levels open their tiers (DG10-A
	// at 90, DG11-A at 101, DG12 tiers at 111/114/118) and levels beyond
	// 120 keep the highest open window. Level 101 is the live transition:
	// the native DG10-C window (its own last provisioned row) and the
	// live DG11-A window are both open and roll against each other.
	for level, wantGroup := range map[int]int{91: 27, 95: 28, 100: 29, 110: 30, 111: 33, 115: 34, 120: 35, 140: 35} {
		group, ok := localEquipmentGroup(&c, level, false)
		if !ok || group != wantGroup {
			t.Fatalf("level %d selects group %d (ok=%v), want %d", level, group, ok, wantGroup)
		}
	}
	group, ok := localEquipmentGroup(&c, 101, false)
	if !ok || group != 29 {
		t.Fatalf("level 101 rolls from group %d (ok=%v), want the native 29 first", group, ok)
	}
	hasLiveWindow := false
	for _, threshold := range c.classes[0][100] {
		if threshold.group == 30 {
			hasLiveWindow = true
		}
	}
	if !hasLiveWindow {
		t.Fatal("level 101 lacks the live DG11-A window beside the native one")
	}

	// The witness: every extended row is selectable at its own window -
	// the same production guarantee the native rows carry.
	for key, bucket := range c.buckets {
		for _, ref := range bucket.refs {
			if !strings.Contains(ref.Codename, "_10_") && !strings.Contains(ref.Codename, "_11_") &&
				!strings.Contains(ref.Codename, "_12_") {
				continue
			}
			kind := 0
			if key.rare {
				kind = 1
			}
			found := false
			for level := int(ref.Level); level <= 140 && !found; level++ {
				for _, threshold := range c.classes[kind][level-1] {
					if threshold.group != key.group {
						continue
					}
					// The tier window is open at this level: the bucket
					// resolves (the fallback only walks DOWN, and every
					// group below is natively populated).
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("extended row has no open window: %+v", ref)
			}
		}
	}
}

/*
================
TestExtendedEquipmentLeavesTheProcessCatalogueNative

A flag-off process never installs the supplement: the shared catalogue
has no degree 10+ row and level 90+ selection keeps the native
lower-class fallback.
================
*/
func TestExtendedEquipmentLeavesTheProcessCatalogueNative(t *testing.T) {
	if err := installEquipmentExtended(&equipmentCatalog{buckets: map[equipmentKey]*equipmentBucket{}}); err == nil {
		t.Fatal("installing into an empty catalogue must refuse (the loader's invariants)")
	}
	for key, bucket := range equipment.buckets {
		for _, ref := range bucket.refs {
			if strings.Contains(ref.Codename, "_10_") || strings.Contains(ref.Codename, "_11_") ||
				strings.Contains(ref.Codename, "_12_") {
				t.Fatalf("flag-off catalogue carries extended row %+v under %+v", ref, key)
			}
		}
	}
	// Level 90: the native window names the (natively empty) group 27;
	// the production fallback walks down to the populated native tier.
	group, ok := EquipmentGroup(90, false, 0)
	if !ok || group != 27 {
		t.Fatalf("native level-90 window = %d (ok=%v), want 27", group, ok)
	}
	if equipment.buckets[equipmentKey{0, 27, false}] != nil {
		t.Fatal("level 90's group-27 bucket is populated natively")
	}
}

/*
================
localEquipmentGroup

EquipmentGroup's walk over a named catalogue (the test's own copy).
================
*/
func localEquipmentGroup(c *equipmentCatalog, level int, rare bool) (int, bool) {
	kind := 0
	if rare {
		kind = 1
	}
	if level < 1 || level > len(c.classes[kind]) {
		return 0, false
	}
	row := c.classes[kind][level-1]
	i := sort.Search(len(row), func(i int) bool { return row[i].threshold >= 1 })
	if i == len(row) {
		return 0, false
	}
	return row[i].group, true
}
