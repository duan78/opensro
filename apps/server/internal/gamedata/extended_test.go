/*
===========================================================================

extended_test.go - the extended switch is native unless explicit

The unset state is native, the on states parse like SRO_BETA_GROWTH, and
a disabled process never resolves an extended path even when the root
knob points somewhere hostile. An enabled process without a built
projection reports not-built, not an error: native stays native.

===========================================================================
*/
package gamedata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtendedContentDefaultsToNative(t *testing.T) {
	for _, value := range []string{"", "  ", "off", "0", "false", "yes please", "onion"} {
		t.Setenv(EnvExtendedContent, value)
		if ExtendedContentEnabled() {
			t.Fatalf("value %q must leave the game native", value)
		}
	}
}

func TestExtendedContentEnabledReadsTheEnvironment(t *testing.T) {
	for _, value := range []string{"on", "1", "true", "TRUE", " On "} {
		t.Setenv(EnvExtendedContent, value)
		if !ExtendedContentEnabled() {
			t.Fatalf("value %q must enable the extended content", value)
		}
	}
}

func TestDisabledNeverResolvesAnExtendedPath(t *testing.T) {
	t.Setenv(EnvExtendedContent, "")
	// A root that would be inspected, and fail, if the disabled path ever
	// touched the filesystem.
	t.Setenv(EnvExtendedRoot, filepath.Join(t.TempDir(), "never-built"))
	root, ok, err := ResolveExtended()
	if err != nil {
		t.Fatalf("disabled resolve must not fail: %v", err)
	}
	if ok || root != "" {
		t.Fatalf("disabled resolve must open no path, got root=%q ok=%v", root, ok)
	}
}

func TestEnabledWithoutProjectionStaysNative(t *testing.T) {
	t.Setenv(EnvExtendedContent, "on")
	t.Setenv(EnvExtendedRoot, filepath.Join(t.TempDir(), "not-built-yet"))
	root, ok, err := ResolveExtended()
	if err != nil {
		t.Fatalf("a missing projection is not an error: %v", err)
	}
	if ok {
		t.Fatalf("a missing projection must report not-built")
	}
	if root == "" {
		t.Fatalf("the reported root tells the operator what to build")
	}
}

func TestEnabledProjectionRootFromEnv(t *testing.T) {
	t.Setenv(EnvExtendedContent, "on")
	built := t.TempDir()
	t.Setenv(EnvExtendedRoot, built)
	root, ok, err := ResolveExtended()
	if err != nil || !ok {
		t.Fatalf("a built projection resolves, got ok=%v err=%v", ok, err)
	}
	if filepath.Clean(root) != filepath.Clean(built) {
		t.Fatalf("root %q must be the configured projection", root)
	}
}

/*
================
writeExtendedFixture

A minimal sealed projection in a temp directory: a manifest whose files
carry true sizes and digests. The happy path loads; a moved cap or a
drifted file refuses.
================
*/
func writeExtendedFixture(t *testing.T, mutate func(manifest map[string]any, files map[string][]byte)) string {
	t.Helper()
	files := map[string][]byte{
		"leveldata.json": []byte("{\"rows\":[[\"140\",\"578982029973906\"]]}"),
		"levelgold.json": []byte("{\"rows\":[[\"140\",\"515\"]]}"),
	}
	manifest := map[string]any{
		"format":        "sro-extended-game-data",
		"schemaVersion": 1,
		"sourceClient":  "isro-live-2026",
		"derivedCap":    140,
		"contentDigest": "0",
	}
	if mutate != nil {
		mutate(manifest, files)
	}
	root := t.TempDir()
	sealed := make([]map[string]any, 0, len(files))
	for name, bytes := range files {
		if err := os.WriteFile(filepath.Join(root, name), bytes, 0o600); err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(bytes)
		sealed = append(sealed, map[string]any{
			"path":   name,
			"bytes":  len(bytes),
			"sha256": hex.EncodeToString(digest[:]),
		})
	}
	manifest["files"] = sealed
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), manifestBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestLoadExtendedVerifiesTheSeal(t *testing.T) {
	t.Setenv(EnvExtendedContent, "on")
	t.Setenv(EnvExtendedRoot, writeExtendedFixture(t, nil))
	extended, err := LoadExtended()
	if err != nil {
		t.Fatalf("a sealed projection loads: %v", err)
	}
	if extended.Manifest.DerivedCap != ExtendedLevelCap || extended.LeveldataPath == "" {
		t.Fatalf("unexpected projection %+v", extended.Manifest)
	}
}

func TestLoadExtendedRefusesAMovedCap(t *testing.T) {
	t.Setenv(EnvExtendedContent, "on")
	root := writeExtendedFixture(t, func(manifest map[string]any, files map[string][]byte) {
		manifest["derivedCap"] = 150
	})
	t.Setenv(EnvExtendedRoot, root)
	if _, err := LoadExtended(); err == nil || !strings.Contains(err.Error(), "sealed") {
		t.Fatalf("a moved cap must refuse, got %v", err)
	}
}

func TestLoadExtendedRefusesDriftedFiles(t *testing.T) {
	t.Setenv(EnvExtendedContent, "on")
	root := writeExtendedFixture(t, nil)
	// Same length, one flipped byte: the digest branch, not the size one.
	drifted, err := os.ReadFile(filepath.Join(root, "leveldata.json"))
	if err != nil {
		t.Fatal(err)
	}
	drifted[5] ^= 1
	if err := os.WriteFile(filepath.Join(root, "leveldata.json"), drifted, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvExtendedRoot, root)
	if _, err := LoadExtended(); err == nil || !strings.Contains(err.Error(), "drifted") {
		t.Fatalf("a drifted file must refuse, got %v", err)
	}
}

func TestLoadExtendedWithoutProjectionIsExplicit(t *testing.T) {
	t.Setenv(EnvExtendedContent, "on")
	t.Setenv(EnvExtendedRoot, filepath.Join(t.TempDir(), "absent"))
	if _, err := LoadExtended(); !errors.Is(err, ErrExtendedNotBuilt) {
		t.Fatalf("a missing projection reports ErrExtendedNotBuilt, got %v", err)
	}
}
