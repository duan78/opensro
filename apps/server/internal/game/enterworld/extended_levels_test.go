/*
===========================================================================

extended_levels_test.go - the live-2026 curve over the sealed projection

The reader keeps every native contract: real rows answer, missing rows
refuse, the job cells keep the shared table, and a file that fails to
parse poisons the whole source instead of half-loading a curve. The
fixture rows are the live-2026 file's own values (measured 2026-10-09).

===========================================================================
*/
package enterworld

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCurveFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	levelRows := [][]string{
		{"1", "118", "1", "0", "0", "24", "70875", "70875", "70875"},
		{"4", "1880", "2", "0", "0", "94", "-1", "-1", "-1"},
		{"90", "200532065", "352", "0", "0", "6949", "-1", "-1", "-1"},
		{"140", "578982029973906", "823", "0", "0", "30462", "-1", "-1", "-1"},
		{"150", "1312995891532097", "900", "0", "0", "33500", "-1", "-1", "-1"},
	}
	goldRows := [][]string{
		{"1", "28", "42"},
		{"140", "515", "772"},
	}
	if err := os.WriteFile(
		filepath.Join(dir, "leveldata.json"),
		[]byte("{\"rows\":"+marshalRows(t, levelRows)+"}"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, "goldcurve.json"),
		[]byte("{\"rows\":"+marshalRows(t, goldRows)+"}"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	return dir
}

func marshalRows(t *testing.T, rows [][]string) string {
	t.Helper()
	parts := make([]string, len(rows))
	for index, row := range rows {
		cells := make([]string, len(row))
		for cellIndex, cell := range row {
			cells[cellIndex] = "\"" + cell + "\""
		}
		parts[index] = "[" + strings.Join(cells, ",") + "]"
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func TestExtendedLevelsAnswersTheLiveCurve(t *testing.T) {
	dir := writeCurveFixture(t)
	levels := NewExtendedLevels(filepath.Join(dir, "leveldata.json"), filepath.Join(dir, "goldcurve.json"))
	if req, ok := levels.ExpRequired(140); !ok || req != 578982029973906 {
		t.Fatalf("level 140 requires the live curve's exp, got %d/%v", req, ok)
	}
	if req, ok := levels.ExpRequired(151); ok {
		t.Fatalf("a row past the table refuses, got %d", req)
	}
	if basis, ok := levels.MonsterExpBasis(1); !ok || basis != 24 {
		t.Fatalf("monster basis column matches the shared table, got %d/%v", basis, ok)
	}
	if cost, ok := levels.SkillPointCost(1); !ok || cost != 1 {
		t.Fatalf("skill point column, got %d/%v", cost, ok)
	}
	if job, ok := levels.JobExpRequired(1, 1); !ok || job != 70875 {
		t.Fatalf("job cells keep the shared table, got %d/%v", job, ok)
	}
	if _, ok := levels.JobExpRequired(4, 1); ok {
		t.Fatalf("the -1 job rows stay unset, exactly like the native loader")
	}
	if gold, ok := levels.WithdrawalGoldBasis(140); !ok || gold != 515 {
		t.Fatalf("gold basis reads the live dg curve, got %d/%v", gold, ok)
	}
	if levels.Len() != 5 {
		t.Fatalf("five priced rows, got %d", levels.Len())
	}
}

func TestExtendedLevelsRefusesWhenAFileIsMissing(t *testing.T) {
	dir := writeCurveFixture(t)
	levels := NewExtendedLevels(filepath.Join(dir, "absent.json"), filepath.Join(dir, "goldcurve.json"))
	if _, ok := levels.ExpRequired(1); ok {
		t.Fatalf("a source without its leveldata refuses every lookup")
	}
	if levels.Err() == nil {
		t.Fatalf("the load failure sticks for the boot wiring to report")
	}
}

func TestExtendedLevelsRefusesBrokenJson(t *testing.T) {
	dir := writeCurveFixture(t)
	if err := os.WriteFile(filepath.Join(dir, "leveldata.json"), []byte("{\"rows\":Truncated"), 0o600); err != nil {
		t.Fatal(err)
	}
	levels := NewExtendedLevels(filepath.Join(dir, "leveldata.json"), filepath.Join(dir, "goldcurve.json"))
	if _, ok := levels.ExpRequired(1); ok {
		t.Fatalf("broken json poisons the whole source")
	}
}
