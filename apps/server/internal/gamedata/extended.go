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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

// ExtendedManifestFormat, ExtendedSourceClient and ExtendedSchemaVersion
// pin the projection's identity; the manifest never mentions 1.150.
const (
	ExtendedManifestFormat = "sro-extended-game-data"
	ExtendedSourceClient   = "isro-live-2026"
	ExtendedSchemaVersion  = 1
)

// ExtendedLevelCap is the owner-sealed cap (2026-10-09, charter 4.5): the
// complete-chain rule redelivered 140 from the live-2026 data. A
// projection whose derived cap differs is an owner decision to revisit,
// never a silent change, so the loader refuses it.
const ExtendedLevelCap int64 = 140

// ErrExtendedNotBuilt: the mode is on but no projection exists. The
// native game keeps running; the operator gets the build command.
var ErrExtendedNotBuilt = errors.New("extended projection not built (scripts/build/server/buildExtendedGameDataBundle.mjs)")

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

/*
================
Extended

The loaded, verified extended projection. The paths are inputs for the
progression layer (mission M2); the manifest is the sealed evidence.
================
*/
type Extended struct {
	Root          string
	Manifest      ExtendedManifest
	LeveldataPath string
	LevelgoldPath string
	CensusPath    string
}

/*
================
ExtendedManifest

Mirrors scripts/build/server/buildExtendedGameDataBundle.mjs. Every listed
file is verified by size and SHA-256 before the projection is usable:
the extended game never runs on data that drifted from its seal.
================
*/
type ExtendedManifest struct {
	Format        string                         `json:"format"`
	SchemaVersion int                            `json:"schemaVersion"`
	SourceClient  string                         `json:"sourceClient"`
	Archives      map[string]ExtendedArchiveSeal `json:"archives"`
	DerivedCap    int64                          `json:"derivedCap"`
	Chain         ExtendedChain                  `json:"chain"`
	Counts        ExtendedCounts                 `json:"counts"`
	ContentDigest string                         `json:"contentDigest"`
	Files         []ExtendedSealedFile           `json:"files"`
}

// ExtendedArchiveSeal is one source PK2 archive's identity.
type ExtendedArchiveSeal struct {
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

// ExtendedChain records the complete-chain inputs that derived the cap.
type ExtendedChain struct {
	XPMaxLevel   int64 `json:"xpMaxLevel"`
	GoldMaxLevel int64 `json:"goldMaxLevel"`
	GearDegree   int64 `json:"gearDegree"`
	MobBandCount int64 `json:"mobBandCount"`
}

// ExtendedCounts are the projection's row censuses.
type ExtendedCounts struct {
	LevelRows     int64 `json:"levelRows"`
	GoldRows      int64 `json:"goldRows"`
	ItemRows      int64 `json:"itemRows"`
	CharacterRows int64 `json:"characterRows"`
}

// ExtendedSealedFile is one projection file pinned by size and digest.
type ExtendedSealedFile struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

/*
================
LoadExtended

Resolve, parse and verify the extended projection. Only called with the
mode on; every other state is an explicit error, never a native fallback
that would quietly run a different game than the operator asked for.
================
*/
func LoadExtended() (Extended, error) {
	root, ok, err := ResolveExtended()
	if err != nil {
		return Extended{}, err
	}
	if !ok {
		return Extended{}, ErrExtendedNotBuilt
	}
	manifestBytes, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		return Extended{}, fmt.Errorf("read extended manifest: %w", err)
	}
	var manifest ExtendedManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return Extended{}, fmt.Errorf("parse extended manifest: %w", err)
	}
	if manifest.Format != ExtendedManifestFormat || manifest.SchemaVersion != ExtendedSchemaVersion {
		return Extended{}, fmt.Errorf("extended manifest %q schema %d is not %q schema %d",
			manifest.Format, manifest.SchemaVersion, ExtendedManifestFormat, ExtendedSchemaVersion)
	}
	if manifest.SourceClient != ExtendedSourceClient {
		return Extended{}, fmt.Errorf("extended manifest source client %q is not %q", manifest.SourceClient, ExtendedSourceClient)
	}
	if manifest.DerivedCap != ExtendedLevelCap {
		return Extended{}, fmt.Errorf("extended manifest derived cap %d is not the sealed %d: the owner lifts the cap, not a rebuild",
			manifest.DerivedCap, ExtendedLevelCap)
	}
	for _, sealed := range manifest.Files {
		bytes, err := os.ReadFile(filepath.Join(root, sealed.Path))
		if err != nil {
			return Extended{}, fmt.Errorf("read extended file %s: %w", sealed.Path, err)
		}
		if int64(len(bytes)) != sealed.Bytes {
			return Extended{}, fmt.Errorf("extended file %s is %d bytes, sealed %d", sealed.Path, len(bytes), sealed.Bytes)
		}
		digest := sha256.Sum256(bytes)
		if hex.EncodeToString(digest[:]) != sealed.SHA256 {
			return Extended{}, fmt.Errorf("extended file %s drifted from its sealed digest", sealed.Path)
		}
	}
	return Extended{
		Root:          root,
		Manifest:      manifest,
		LeveldataPath: filepath.Join(root, "leveldata.json"),
		LevelgoldPath: filepath.Join(root, "levelgold.json"),
		CensusPath:    filepath.Join(root, "census.json"),
	}, nil
}
