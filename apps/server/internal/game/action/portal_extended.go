/*
===========================================================================

portal_extended.go - the live-2026 teleport plane graft (E7)

Extended content (isro-live-2026), port-only, not v1.150-native. The
extended textdata tree carries the live client's full teleport plane -
107 buildings, 332 destinations, 337 links (measured 2026-10-09), whose
destination coordinates include the post-1.150 zones this world lane
serves. The graft merges that plane behind the native catalogue: every
native identity (destination id, source ref, link pair, building
codename) keeps its native row, only new identities enter, and rows
naming a world the v1.150 instance package never defined - the live
table's modern dungeon worlds - are skipped with a counter, because the
native engine has no plane to admit them into. Link rows may carry two
trailing unknown columns (21 native, 23 live); the first 21 are the
contract, the tail is tolerated. Shape violations fail the boot: a
drifted table is never a silent gap.
===========================================================================
*/
package action

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"

	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/world/instance"
	"opensro.online/server/internal/game/world/simulation"
)

// PortalMergeStats reports what one extended merge did, for the boot log.
type PortalMergeStats struct {
	BuildingsAdded        int
	DestinationsAdded     int
	LinksAdded            int
	SkippedNativeIdentity int
	SkippedUnknownWorld   int
	SkippedUnresolvedLink int
	SkippedUnsupported    int
}

/*
================
MergeExtendedPortalDir

Merge the extended teleport tables into the configured portal catalogue.
The native catalogue must already be loaded (ConfigurePortals); the
native row wins every shared identity.
================
*/
func (rt *Runtime) MergeExtendedPortalDir(dir string) (PortalMergeStats, error) {
	var stats PortalMergeStats
	if rt.portals == nil {
		return stats, fmt.Errorf("portal catalogue is not configured")
	}
	c := rt.portals

	// Buildings: the classification maps cover every extended ref (native
	// refs included, so a new destination at a native gate resolves its
	// gate kind), while the catalogue's codename map only gains new names.
	buildingRows := enterworld.ReadTextdataFile(filepath.Join(dir, "teleportbuilding.txt"))
	if len(buildingRows) == 0 {
		return stats, fmt.Errorf("extended teleportbuilding is absent or empty")
	}
	isBuilding := map[uint32]bool{}
	fortressGate := map[uint32]bool{}
	gateKind := map[uint32]uint8{}
	for i, r := range buildingRows {
		if r[0] != "1" {
			continue
		}
		if len(r) < 13 {
			return stats, fmt.Errorf("extended teleportbuilding row %d is truncated", i+1)
		}
		id, e := strconv.ParseUint(r[1], 10, 32)
		if e != nil {
			return stats, fmt.Errorf("extended teleportbuilding identity at row %d", i+1)
		}
		isBuilding[uint32(id)] = true
		if len(r) > 12 && r[9] == "4" && r[10] == "1" && r[11] == "1" {
			if kind, e := strconv.ParseUint(r[12], 10, 8); e == nil {
				gateKind[uint32(id)] = uint8(kind)
			}
			fortressGate[uint32(id)] = r[12] == "1"
		}
		if len(r) > 2 {
			if _, exists := c.buildings[r[2]]; exists {
				stats.SkippedNativeIdentity++
			} else {
				c.buildings[r[2]] = uint32(id)
				stats.BuildingsAdded++
			}
		}
	}

	// Destinations: new ids only, native sources keep their gate, and a
	// world the native engine never defined admits nowhere - counted.
	dataRows := enterworld.ReadTextdataFile(filepath.Join(dir, "teleportdata.txt"))
	if len(dataRows) == 0 {
		return stats, fmt.Errorf("extended teleportdata is absent or empty")
	}
	for i, r := range dataRows {
		if r[0] != "1" {
			continue
		}
		if len(r) < 13 {
			return stats, fmt.Errorf("extended teleportdata row %d is truncated", i+1)
		}
		id, e := strconv.ParseUint(r[1], 10, 32)
		if e != nil || id == 0 {
			return stats, fmt.Errorf("extended teleportdata identity at row %d", i+1)
		}
		if _, exists := c.destinations[uint32(id)]; exists {
			stats.SkippedNativeIdentity++
			continue
		}
		ref, e := strconv.ParseUint(r[3], 10, 32)
		if e != nil {
			return stats, fmt.Errorf("extended teleportdata ref at row %d", i+1)
		}
		region, e := strconv.ParseInt(r[5], 10, 32)
		if e != nil || region < -32768 || region > 65535 {
			return stats, fmt.Errorf("extended teleportdata region at row %d", i+1)
		}
		xyz := [3]float64{}
		for j := range xyz {
			v, e := strconv.ParseFloat(r[6+j], 64)
			if e != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return stats, fmt.Errorf("extended teleportdata coordinate at row %d", i+1)
			}
			xyz[j] = v
		}
		recall, e := strconv.ParseUint(r[10], 10, 32)
		if e != nil || recall > 1 {
			return stats, fmt.Errorf("extended teleportdata recall eligibility at row %d", i+1)
		}
		world, e := strconv.ParseUint(r[12], 10, 16)
		if e != nil {
			return stats, fmt.Errorf("extended teleportdata world at row %d", i+1)
		}
		if _, known := instance.Lookup(instance.DefinitionID(world)); !known {
			stats.SkippedUnknownWorld++
			continue
		}
		if ref != 0 {
			if _, exists := c.sources[uint32(ref)]; exists {
				stats.SkippedNativeIdentity++
				continue
			}
		}
		c.destinations[uint32(id)] = portalDestination{
			id: uint32(id), ref: uint32(ref), code: r[2], building: isBuilding[uint32(ref)],
			fortressGate: fortressGate[uint32(ref)], gateKind: gateKind[uint32(ref)], recall: recall == 1,
			world: instance.DefinitionID(world),
			spawn: simulation.Spawn{RegionID: uint16(region), X: xyz[0], Y: xyz[1], Z: xyz[2]},
		}
		if ref != 0 {
			c.sources[uint32(ref)] = uint32(id)
		}
		stats.DestinationsAdded++
	}

	// Links: new pairs only; both ends must resolve in the merged catalogue
	// and the native engine must support the row's conditions. The live
	// table keeps the native column order for source/target/fee and the
	// five condition triplets, but its scheduling columns drifted: the
	// native r[4]=1/r[5]=0 pair reads 0/0 there and a 1 rides in the first
	// trailing column (measured 2026-10-09 on the Jangan-Donwha row both
	// tables carry, fee 5000 identical). The merge therefore reads the
	// triplets from r[6..20] and ignores the drifted pair and the tail.
	linkRows := enterworld.ReadTextdataFile(filepath.Join(dir, "teleportlink.txt"))
	if len(linkRows) == 0 {
		return stats, fmt.Errorf("extended teleportlink is absent or empty")
	}
	for i, r := range linkRows {
		if r[0] != "1" {
			continue
		}
		// 21 columns native; the live table appends two unknown ones.
		if len(r) != 21 && len(r) != 23 {
			return stats, fmt.Errorf("extended teleportlink row %d shape", i+1)
		}
		values := make([]uint32, 20)
		for j := range values {
			v, e := strconv.ParseUint(r[j+1], 10, 32)
			if e != nil {
				return stats, fmt.Errorf("extended teleportlink row %d: %w", i+1, e)
			}
			values[j] = uint32(v)
		}
		// The native scheduling/combination contract lives in values[3]/[4];
		// the drifted live pair is ignored, and only a nonzero combination
		// the native engine cannot express still skips the row.
		if values[4] != 0 {
			stats.SkippedUnsupported++
			continue
		}
		source, target := values[0], values[1]
		if _, ok := c.destinations[source]; !ok {
			stats.SkippedUnresolvedLink++
			continue
		}
		if _, ok := c.destinations[target]; !ok {
			stats.SkippedUnresolvedLink++
			continue
		}
		key := [2]uint32{source, target}
		if _, ok := c.links[key]; ok {
			stats.SkippedNativeIdentity++
			continue
		}
		link := portalLink{source: source, target: target, fee: int64(values[2])}
		unsupported := false
		for j := 5; j < 20; j += 3 {
			k, a, b := values[j], values[j+1], values[j+2]
			if k == 0 {
				if a != 0 || b != 0 {
					unsupported = true
					break
				}
				continue
			}
			if k > 2 || k == 1 && a > b || k == 2 && (a != 0 || b != 0) {
				unsupported = true
				break
			}
			link.conditions = append(link.conditions, portalCondition{k, a, b})
		}
		if unsupported {
			stats.SkippedUnsupported++
			continue
		}
		c.links[key] = link
		stats.LinksAdded++
	}
	return stats, nil
}
