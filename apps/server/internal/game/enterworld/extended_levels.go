/*
===========================================================================

extended_levels.go - the live-2026 level curve (port-only, not native)

Extended content (isro-live-2026). Reads the sealed projection the
extended bundle build produced instead of the native textdata: the same
CRefLevel column semantics the v1.150 reader documents (leveldata.go) -
level 0, exp requirement 1, SP cost 2, monster exp basis 5, job exp 6-8 -
over the live client's steeper curve, and the gold-walk basis from the
live dg.txt (the table the v1.150 CDropGoldData reader owns natively).
Cell tolerance mirrors TextdataLevels (short rows and empty cells skip,
a missing row refuses the mutation); a file that fails to parse sticks
as a load error the boot wiring reports. Nothing here is read unless
SRO_EXTENDED_CONTENT is on (gamedata.ResolveExtended).

===========================================================================
*/
package enterworld

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// The live-2026 leveldata cells this reader uses, pinned by measurement
// against the shipped file (2026-10-09): row 1 carries exp 118 at column
// 1 and monster basis 24 at column 5, and the job cells 6-8 are
// identical to the v1.150 table (70875 x3 at row 1, -1 from row 11).
const (
	extendedLevelColumn       = 0
	extendedExpColumn         = 1
	extendedSkillPointColumn  = 2
	extendedMonsterExpColumn  = 5
	extendedJobExpColumn      = 6
	extendedJobExpColumnCount = 3
)

/*
================
ExtendedLevels

One lazy owner over the extended projection's leveldata.json and
goldcurve.json. Implements LevelDataSource, JobLevelDataSource and the
gold basis TextdataLevels exposes, so it replaces the native table
without any call site changing.
================
*/
type ExtendedLevels struct {
	leveldataPath string
	goldCurvePath string

	once             sync.Once
	loadErr          error
	spCostByLvl      map[int64]int64
	expByLvl         map[int64]int64
	mobExpBasisByLvl map[int64]int64
	goldBasisByLvl   map[int64]int64
	jobExpByLvl      map[int64][extendedJobExpColumnCount]int64
}

/*
================
NewExtendedLevels
================
*/
func NewExtendedLevels(leveldataPath, goldCurvePath string) *ExtendedLevels {
	return &ExtendedLevels{leveldataPath: leveldataPath, goldCurvePath: goldCurvePath}
}

/*
================
load

Both tables parse or the whole source refuses: an extended curve a row
short would walk characters into a wall.
================
*/
func (e *ExtendedLevels) load() {
	type curveFile struct {
		Rows [][]string `json:"rows"`
	}
	var leveldata curveFile
	if raw, err := os.ReadFile(e.leveldataPath); err != nil {
		e.loadErr = fmt.Errorf("extended leveldata: %w", err)
		return
	} else if err := json.Unmarshal(raw, &leveldata); err != nil {
		e.loadErr = fmt.Errorf("extended leveldata: %w", err)
		return
	}
	var goldcurve curveFile
	if raw, err := os.ReadFile(e.goldCurvePath); err != nil {
		e.loadErr = fmt.Errorf("extended gold curve: %w", err)
		return
	} else if err := json.Unmarshal(raw, &goldcurve); err != nil {
		e.loadErr = fmt.Errorf("extended gold curve: %w", err)
		return
	}
	e.spCostByLvl = map[int64]int64{}
	e.expByLvl = map[int64]int64{}
	e.mobExpBasisByLvl = map[int64]int64{}
	e.goldBasisByLvl = map[int64]int64{}
	e.jobExpByLvl = map[int64][extendedJobExpColumnCount]int64{}
	for _, fields := range leveldata.Rows {
		if len(fields) <= extendedSkillPointColumn {
			continue
		}
		level, ok := textdataInt(fields[extendedLevelColumn])
		if !ok || level < 1 {
			continue
		}
		if cost, ok := textdataInt(fields[extendedSkillPointColumn]); ok && cost >= 0 {
			e.spCostByLvl[level] = cost
		}
		if exp, ok := textdataInt(fields[extendedExpColumn]); ok && exp > 0 {
			e.expByLvl[level] = exp
		}
		if len(fields) > extendedMonsterExpColumn {
			if basis, ok := textdataInt(fields[extendedMonsterExpColumn]); ok && basis > 0 {
				e.mobExpBasisByLvl[level] = basis
			}
		}
		if len(fields) > extendedJobExpColumn+extendedJobExpColumnCount-1 {
			var jobs [extendedJobExpColumnCount]int64
			valid := true
			for index := range jobs {
				value, ok := textdataInt(fields[extendedJobExpColumn+index])
				valid = valid && ok && value > 0
				jobs[index] = value
			}
			if valid {
				e.jobExpByLvl[level] = jobs
			}
		}
	}
	for _, fields := range goldcurve.Rows {
		if len(fields) < 2 {
			continue
		}
		level, validLevel := textdataInt(fields[extendedLevelColumn])
		basis, validBasis := textdataInt(fields[extendedExpColumn])
		if validLevel && validBasis && level > 0 && basis > 0 {
			e.goldBasisByLvl[level] = basis
		}
	}
}

/*
================
ready

The shared lazy gate: first failure sticks, exactly like TextdataLevels.
================
*/
func (e *ExtendedLevels) ready() bool {
	e.once.Do(e.load)
	return e.loadErr == nil
}

/*
================
SkillPointCost
================
*/
func (e *ExtendedLevels) SkillPointCost(level int64) (int64, bool) {
	if !e.ready() {
		return 0, false
	}
	cost, ok := e.spCostByLvl[level]
	return cost, ok
}

/*
================
ExpRequired
================
*/
func (e *ExtendedLevels) ExpRequired(level int64) (int64, bool) {
	if !e.ready() {
		return 0, false
	}
	exp, ok := e.expByLvl[level]
	return exp, ok
}

/*
================
MonsterExpBasis
================
*/
func (e *ExtendedLevels) MonsterExpBasis(level int64) (int64, bool) {
	if !e.ready() {
		return 0, false
	}
	basis, ok := e.mobExpBasisByLvl[level]
	return basis, ok
}

/*
================
JobExpRequired
================
*/
func (e *ExtendedLevels) JobExpRequired(grade int64, job uint8) (int64, bool) {
	if !e.ready() {
		return 0, false
	}
	row, ok := e.jobExpByLvl[grade]
	if !ok || job < 1 || job > extendedJobExpColumnCount {
		return 0, false
	}
	return row[job-1], true
}

/*
================
WithdrawalGoldBasis
================
*/
func (e *ExtendedLevels) WithdrawalGoldBasis(level int64) (int64, bool) {
	if !e.ready() {
		return 0, false
	}
	basis, ok := e.goldBasisByLvl[level]
	return basis, ok
}

/*
================
Len
================
*/
func (e *ExtendedLevels) Len() int {
	if !e.ready() {
		return 0
	}
	return len(e.expByLvl)
}

/*
================
Err

The load failure, for the boot wiring to fail loudly.
================
*/
func (e *ExtendedLevels) Err() error {
	e.once.Do(e.load)
	return e.loadErr
}
