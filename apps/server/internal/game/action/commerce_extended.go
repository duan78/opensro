/*
===========================================================================

commerce_extended.go - the live-2026 shop catalogue graft (M7)

Extended content (isro-live-2026), port-only, not v1.150-native. The
live client ships the whole shop plane in its textdata (six commerce
tables plus the two NPC mappings). The graft loads that plane over the
extended textdata and merges it behind the native catalogue: every
native tab id keeps its native offers, new tab ids enter, and the
extended goods resolve through the same item reference source the
runtime already uses (the overlay in an extended boot). An unconfigured
runtime keeps exactly the native catalogue.
===========================================================================
*/
package action

import (
	"fmt"

	"opensro.online/server/internal/game/item/commerce"
)

/*
================
MergeExtendedCommerce

Load the extended shop plane and merge it into the configured catalogue.
Called once at boot by the composition root when the extended content is
on; a failed merge is a boot error - a half-served shop is never shipped.
================
*/
func (rt *Runtime) MergeExtendedCommerce(dir string) (int, error) {
	if rt.Commerce == nil {
		return 0, fmt.Errorf("commerce catalogue is not configured")
	}
	extended, err := commerce.Load(dir, rt.deps.ItemReferences())
	if err != nil {
		return 0, fmt.Errorf("extended commerce: %w", err)
	}
	added := 0
	for tab, offers := range extended.Tabs {
		if _, native := rt.Commerce.Tabs[tab]; native {
			continue
		}
		rt.Commerce.Tabs[tab] = offers
		added++
	}
	return added, nil
}
