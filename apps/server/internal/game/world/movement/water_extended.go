/*
===========================================================================

water_extended.go - the extended world's movement authority chain

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended projection's movement/ mirror (built by
scripts/build_extended_world_resources.mjs over the live-2026 bundles)
carries the regions the v1.150 world never served. The native authority
answers first and always wins; the chain only resolves what the native
tree has no surface for, so a flag-off boot (which never installs the
chain) and every native region keep exactly today's answers. The chain is
installed once at boot, before the first query.
===========================================================================
*/
package movement

import (
	"fmt"
	"os"
	"path/filepath"
)

/*
================
SetExtendedAuthorityRoot

Install the extended movement mirror as the fallback authority. A named
mirror that cannot be read is a configuration error, never a silent
degrade: the operator asked for the extended world and the boot says what
is missing.
================
*/
func (v *WaterValidator) SetExtendedAuthorityRoot(movementRoot string) error {
	if v == nil {
		return fmt.Errorf("movement: nil validator cannot take an extended authority root")
	}
	if _, err := os.Stat(filepath.Join(movementRoot, "catalog.json")); err != nil {
		return fmt.Errorf("movement: extended movement mirror unreadable under %s: %w", movementRoot, err)
	}
	// The extended mirror IS the movement root (the native constructor
	// appends "movement" to a world-authority dir; the extended projection
	// has no world-authority parent).
	extended := newWaterValidator(movementRoot, "catalog.json", filepath.Join("dungeon", "dungeon-resources.json"))
	if extended.loadCatalog() == nil {
		return fmt.Errorf("movement: extended movement mirror has no readable catalog under %s", movementRoot)
	}
	v.extendedAuthority = extended
	return nil
}

/*
================
ExtendedMovementRegions

The chained mirror's outdoor region count, for the boot log. Zero when no
chain is installed.
================
*/
func (v *WaterValidator) ExtendedMovementRegions() int {
	if v == nil || v.extendedAuthority == nil {
		return 0
	}
	catalog := v.extendedAuthority.loadCatalog()
	if catalog == nil {
		return 0
	}
	count := 0
	for name := range catalog.RegionsByID {
		if parseRegionID(name) != 0 {
			count++
		}
	}
	return count
}
