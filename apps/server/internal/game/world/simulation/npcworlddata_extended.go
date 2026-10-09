/*
===========================================================================

npcworlddata_extended.go - the live-2026 teleport gate graft

Extended content (isro-live-2026), port-only, not v1.150-native. The
live client's teleportbuilding table carries the full gate set - the
native cities' gates plus the gates of the post-1.150 zones this world
lane serves. The graft walks it over the extended textdata tree and adds
ONLY gates whose ref the native pass did not already place: a ref both
tables carry keeps its native spawn (native wins, same rule as every
extended overlay), and the new ids stay in the same 250000 static band
so every consumer's identity contract is unchanged.
===========================================================================
*/
package simulation

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"opensro.online/server/internal/domain"
)

/*
================
AppendExtendedTeleportGates

The extended building table appended to an already-native roster. Row
shape violations still fail loudly - a drifted table is a build error,
never a silent gap. Two live-data shapes are skipped and counted, not
errors: the region-zero dynamic instance gates (as native) and the
marker rows that reuse the building table with zero bounds (job-structure
spawn markers like STORE_HUNTER_SPAWN, measured 2026-10-09) - they are
legitimate live rows, just not functional gates.
================
*/
func AppendExtendedTeleportGates(dir string, roster []NpcDef) ([]NpcDef, int, error) {
	names, fortresses := readTeleportNamesAndFortresses(dir)
	rows := readNpcTabbed(filepath.Join(dir, "teleportbuilding.txt"))
	if len(rows) == 0 {
		return nil, 0, fmt.Errorf("missing extended teleportbuilding table")
	}
	gated := make(map[uint32]bool, len(roster))
	for _, row := range roster {
		if row.Teleport != nil {
			gated[row.RefObjID] = true
		}
	}
	result := append([]NpcDef(nil), roster...)
	added := 0
	skippedMarkers := 0
	for index, c := range rows {
		if len(c) == 0 || c[0] != "1" {
			continue
		}
		gate, ref, skip, err := parseTeleportBuildingRow(c, names, fortresses)
		if err != nil {
			return nil, 0, fmt.Errorf("extended teleportbuilding row %d %s", index+1, err)
		}
		if skip || gated[ref] {
			continue
		}
		if gate.Teleport.Height <= 0 || gate.Teleport.Radius <= 0 {
			skippedMarkers++
			continue
		}
		gated[ref] = true
		result = append(result, gate)
		added++
	}
	log.Infof(
		"npc roster: extended building table grafted %d live-2026 gates, skipped %d zero-bound marker rows (not v1.150-native)",
		added,
		skippedMarkers,
	)
	return result, added, ValidateNpcRoster(result)
}

/*
================
AppendExtendedNpcWorldRoster

The live towns' service NPCs (M7, port-only, not v1.150-native): the
extended textdata's own roster - its npcpos anchors joined to its
characterdata NPCs, with store groups from its refmappings - appended to
an already-native roster. A codename both rosters carry keeps its native
row (native wins, the same rule as every extended graft); the additions
are re-identified into a dedicated band of the NPC range (NPCGIDBase +
10000, far above any shipped roster and below the teleport gates'
250000 band), so no identity ever collides.
================
*/
func AppendExtendedNpcWorldRoster(dir string, roster []NpcDef) ([]NpcDef, int, error) {
	extended := LoadNpcWorldRoster(dir)
	if len(extended) == 0 {
		return nil, 0, fmt.Errorf("extended NPC roster is empty under %s", dir)
	}
	known := make(map[string]bool, len(roster))
	for _, row := range roster {
		known[row.Codename] = true
	}
	result := append([]NpcDef(nil), roster...)
	added := 0
	for _, row := range extended {
		if known[row.Codename] {
			continue
		}
		known[row.Codename] = true
		row.ObjectID = domain.NPCGIDBase + 10000 + uint32(added) + 1
		result = append(result, row)
		added++
	}
	return result, added, ValidateNpcRoster(result)
}
