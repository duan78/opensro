/*
===========================================================================

merge_test.go - the native areas stay owned, the extended ones join

Extended content (isro-live-2026, port-only, not native): the merge adds
the extended catalog's areas beside the native ones and refuses a slug
or region the native catalog already owns - the seed never silently
replaces a native area.

===========================================================================
*/
package worldarea

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func writeAuthorityCatalog(t *testing.T, slug string, region uint16) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "areas"), 0o755); err != nil {
		t.Fatal(err)
	}
	area := `{"slug":"` + slug + `","regionId":` + strconv.Itoa(int(region)) +
		`,"access":"gm","entry":{"x":900,"y":0,"z":920,"angle":16384},"population":[]}`
	catalog := `{"format":"sro-server-world-area-catalog","version":1,"areas":[` + area + `]}`
	if err := os.WriteFile(filepath.Join(dir, "areas", "catalog.json"), []byte(catalog), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestMergeAddsTheExtendedAreasBesideTheNativeOnes(t *testing.T) {
	native, err := LoadAuthority(writeAuthorityCatalog(t, "manyang-lab", 0x7e7e))
	if err != nil {
		t.Fatal(err)
	}
	extended, err := LoadAuthority(writeAuthorityCatalog(t, "extended-content-lab", 0x62a8))
	if err != nil {
		t.Fatal(err)
	}
	merged, err := Merge(native, extended)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := merged.bySlug["manyang-lab"]; !ok {
		t.Fatal("the native area stays owned")
	}
	if _, ok := merged.bySlug["extended-content-lab"]; !ok {
		t.Fatal("the extended area joins the catalog")
	}
}

func TestMergeRefusesARegionTheNativeCatalogOwns(t *testing.T) {
	native, err := LoadAuthority(writeAuthorityCatalog(t, "manyang-lab", 0x7e7e))
	if err != nil {
		t.Fatal(err)
	}
	extended, err := LoadAuthority(writeAuthorityCatalog(t, "extended-content-lab", 0x7e7e))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Merge(native, extended); err == nil {
		t.Fatal("a region the native catalog owns must refuse the merge")
	}
}
