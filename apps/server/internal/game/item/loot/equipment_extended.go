/*
===========================================================================

equipment_extended.go - the live-2026 equipment supplement (E4)

Extended content (isro-live-2026), port-only, not v1.150-native. The
live client's degree 10-14 wearables (308 rows: DG10 complete at
90-100, DG11 tier A at 101, DG12 complete at 111-120; the flat-101
DG13/14 rare pre-provisioning stays uncatalogued) drop through the
ordinary equipment path once the supplement is installed. The window
rates are inferred - the evidence file and the generator's audit say so
- continuing the native evidence's uniform patterns. Installation is an
explicit boot decision behind the extended flag: an uninstalled process
serves the native catalogue byte-for-byte, and a native row wins every
shared identity (the generator refuses overlaps).
===========================================================================
*/
package loot

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed .generated/equipment-extended.json
var equipmentExtendedJSON []byte

var extendedInstall struct {
	once sync.Once
	err  error
}

// extendedEquipment names the supplement's rows: the drop path lets them
// fall without magic options while the live client's magic-assignment
// evidence is absent (the native items keep the fail-closed gate).
var extendedEquipment = map[string]bool{}

/*
================
IsExtendedEquipment
================
*/
func IsExtendedEquipment(code string) bool {
	return extendedEquipment[code]
}

/*
================
InstallExtendedEquipment

Merge the supplement into the process catalogue: items into their
buckets and windows into their level rows. Called once at boot by the
composition root when the extended content is on; a second call answers
the first attempt's outcome without redoing the work.
================
*/
func InstallExtendedEquipment() error {
	extendedInstall.once.Do(func() {
		extendedInstall.err = installEquipmentExtended(&equipment)
	})
	return extendedInstall.err
}

/*
================
installEquipmentExtended

The merge itself, over a named catalogue: the process install owns the
global, tests install their own load so the shared native catalogue a
flag-off process serves is never touched.
================
*/
func installEquipmentExtended(c *equipmentCatalog) error {
	var source struct {
		Version int
		Items   []equipmentRef
		Windows []struct {
			Level       int     `json:"level"`
			Group       int     `json:"group"`
			ApplyNormal bool    `json:"applyNormal"`
			NormalRate  float32 `json:"normalRate"`
			ApplyRare   bool    `json:"applyRare"`
			RareRate    float32 `json:"rareRate"`
		}
	}
	if err := json.Unmarshal(equipmentExtendedJSON, &source); err != nil {
		return fmt.Errorf("extended equipment supplement: %w", err)
	}
	if source.Version != 1 {
		return fmt.Errorf("extended equipment supplement: unsupported version %d", source.Version)
	}
	known := map[string]bool{}
	for _, bucket := range c.buckets {
		for _, ref := range bucket.refs {
			known[ref.Codename] = true
		}
	}
	for _, r := range source.Items {
		if known[r.Codename] || r.Codename == "" || r.Country > 1 || r.Group < 0 || r.Group >= 36 ||
			r.Weight == 0 || r.Absolute > 100 || r.Level == 0 {
			return fmt.Errorf("extended equipment supplement: invalid assignment %+v", r)
		}
		known[r.Codename] = true
		extendedEquipment[r.Codename] = true
		key := equipmentKey{r.Country, r.Group, r.Rare}
		b := c.buckets[key]
		if b == nil {
			b = &equipmentBucket{alternatives: map[string][]uint32{}}
			c.buckets[key] = b
		}
		weight := r.Weight
		if len(b.weights) > 0 {
			weight += b.weights[len(b.weights)-1]
		}
		b.refs = append(b.refs, r)
		b.weights = append(b.weights, weight)
		b.alternatives[r.Type] = append(b.alternatives[r.Type], uint32(len(b.refs)-1))
	}
	for _, window := range source.Windows {
		if window.Level < 1 || window.Level > len(c.rows[0]) || window.Group < 0 || window.Group >= 36 {
			return fmt.Errorf("extended equipment supplement: window outside the rows: %+v", window)
		}
		if window.NormalRate < 0 || window.NormalRate > 1 || window.RareRate < 0 || window.RareRate > 1 {
			return fmt.Errorf("extended equipment supplement: window rate out of range: %+v", window)
		}
		// The window rates APPLY (installation only ever happens in the
		// extended mode, where the live itemdata's own window shape wins
		// from level 91 up - at 101 the live degree-11 tier opens where
		// the native rows had provisioned the degree-10 C window). A
		// flag-off process never installs, so its rows stay native. The
		// generator emits one row per window with a per-kind apply flag:
		// a natively active cell (the ordinary DG10 rows run to 101 while
		// the rare ones stop at 97) is never touched.
		if window.ApplyNormal {
			row := c.rows[0][window.Level-1]
			row[window.Group] = window.NormalRate
			c.classes[0][window.Level-1] = classThresholds(row)
		}
		if window.ApplyRare {
			row := c.rows[1][window.Level-1]
			row[window.Group] = window.RareRate
			c.classes[1][window.Level-1] = classThresholds(row)
		}
	}
	return nil
}
