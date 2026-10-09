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
	"path/filepath"
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
