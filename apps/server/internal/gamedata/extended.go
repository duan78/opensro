/*
===========================================================================

extended.go - the extended-content switch (port-only, not native)

Extended content grafts the live-2026 client's data (level cap 140,
degree 11-14 gear, the post-1.150 zones) onto the native v1.150 core.
It is not v1.150-native: everything here ships behind one environment
flag whose unset value is the native game, exactly like
progression.SRO_BETA_GROWTH. The extended projection is built beside the
native one by scripts/build/server/buildExtendedGameDataBundle.mjs
(mission docs/extended-content-mission.md, milestone M1); this file only
answers two questions - is the mode on, and where is its projection. A
disabled process resolves no extended path at all, and an enabled process
with no built projection stays explicitly native, never guessing.

===========================================================================
*/
package gamedata

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvExtendedContent turns the extended content on ("on", "1", "true").
// Unset or any other value keeps the game native (level cap 90).
const EnvExtendedContent = "SRO_EXTENDED_CONTENT"

// EnvExtendedRoot overrides the extended projection root for deployments;
// local development uses the module-relative default below.
const EnvExtendedRoot = "SRO_EXTENDED_GAME_DATA_ROOT"

// extendedDefaultRoot is module-relative, the sibling of the native
// .generated/game-data/<version> directory (see defaultRoot in resolve.go).
var extendedDefaultRoot = filepath.Join(".generated", "game-data", "extended")

/*
================
ExtendedContentEnabled

True only for an explicit "on", "1" or "true"; every other value, like
progression.BetaGrowthFromEnv, is the native default.
================
*/
func ExtendedContentEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(EnvExtendedContent))) {
	case "on", "1", "true":
		return true
	default:
		return false
	}
}

/*
================
ResolveExtended

The extended projection root when the mode is on and the projection is
built, ok=false otherwise. The flag is checked before anything touches
the filesystem, so a native process opens no extended path. A missing
projection is reported as ok=false with a nil error: the caller logs and
stays native, it is not a failure of the native game. Any other filesystem
error is returned.
================
*/
func ResolveExtended() (root string, ok bool, err error) {
	if !ExtendedContentEnabled() {
		return "", false, nil
	}
	root = strings.TrimSpace(os.Getenv(EnvExtendedRoot))
	if root == "" {
		moduleRoot, moduleErr := mainModuleRoot()
		if moduleErr != nil {
			return "", false, fmt.Errorf("%s is required outside the source checkout: %w", EnvExtendedRoot, moduleErr)
		}
		root = filepath.Join(moduleRoot, extendedDefaultRoot)
	}
	absoluteRoot, absErr := filepath.Abs(root)
	if absErr != nil {
		return "", false, fmt.Errorf("resolve extended projection: %w", absErr)
	}
	if _, statErr := os.Stat(absoluteRoot); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return absoluteRoot, false, nil
		}
		return "", false, fmt.Errorf("inspect extended projection %s: %w", absoluteRoot, statErr)
	}
	return absoluteRoot, true, nil
}
