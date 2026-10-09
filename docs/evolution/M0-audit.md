# Audit M0 — squelette de flags (charter V0)

Session du 2026-10-09, worktree `opensro-w-evolution`, branche
`evolution/extended-content-m0` (base `evolution/extended-content` @
95f293c4). Mission : `docs/extended-content-mission.md` §M0.

## Livré

- `apps/server/internal/gamedata/extended.go` — le switch
  `SRO_EXTENDED_CONTENT` (pattern `SRO_BETA_GROWTH` : absent/off = natif)
  et `ResolveExtended()` (racine de la projection étendue,
  `SRO_EXTENDED_GAME_DATA_ROOT` ou
  `.generated/game-data/extended` module-relative). Désactivé : aucun
  chemin étendu résolu (le flag est testé avant tout accès disque).
  Activé sans projection : état explicite « pas construite », pas une
  erreur — le natif reste natif.
- `apps/server/internal/gamedata/extended_test.go` — 5 tests : défauts
  natifs (off/0/false/bruit), activations (on/1/true/TRUE/' On '),
  désactivé n'ouvre aucun chemin, activé-sans-projection reste natif,
  racine construite résolue depuis l'env.
- `apps/server/cmd/services/sro-gameworld/wiring.go` — état du mode
  loggé une fois au boot : ON + chemin, ou avertissement explicite quand
  la projection demandée n'est pas construite (avec la commande de build).
- Client : `extendedContent` dans `ExperimentalOptions` (explicit-true
  seulement, bannière « not native ») + onglet « Content » dans la
  fenêtre Experimental (ligne « Extended content (cap 140) »).
- Tests client : nouveau
  `tests/runtime/experimental-extended.test.mjs` (off par défaut,
  explicit-true, ligne présente, ids uniques, brouillon≠sauvé) ; le test
  exhaustif existant `experimental-options.test.mjs` étendu (OFF,
  liste d'onglets, cycle de brouillon).

## Critères d'acceptation (mission M0) — vérifiés

- Gates : sorties citées ci-dessous, exit 0 dans la session.
- « Flag off n'ouvre aucun chemin étendu » :
  `TestDisabledNeverResolvesAnExtendedPath` PASS (le test pointe
  `SRO_EXTENDED_GAME_DATA_ROOT` vers un répertoire inexistant et prouve
  que la valeur désactivée n'y touche jamais).
- « La fenêtre Experimental affiche la ligne off » :
  `experimental-extended.test.mjs` 3/3 PASS ; le test exhaustif des
  onglets couvre la nouvelle ligne.

## Commandes exécutées (preuves)

- `go test ./internal/gamedata/ -run "Extended|Enabled" -v` → 5 PASS
  (`TestExtendedContentDefaultsToNative`,
  `TestExtendedContentEnabledReadsTheEnvironment`,
  `TestDisabledNeverResolvesAnExtendedPath`,
  `TestEnabledWithoutProjectionStaysNative`,
  `TestEnabledProjectionRootFromEnv`), `ok opensro.online/server/internal/gamedata`.
- `pnpm task run check:server` (Go sur PATH) →
  `server gates: PASS (16 package workers, test cache on, 231.5s)`
  (tidy, vet, golangci-lint, tests, race, govulncheck, release contract).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 31.0s`
  (typecheck, format, encoding, size, pipeline contracts… — check:server
  réutilise le stamp du run complet ci-dessus).
- `node tools/run-tests.mjs tests/runtime/experimental-options.test.mjs
  tests/runtime/experimental-extended.test.mjs` → `tests 9, pass 9, fail 0`.
- `node tools/verify-test-types.mjs` →
  `test types: PASS (271 ledger file(s) carry 2750 error(s); no new errors)`.

## Écarts connus (hors périmètre M0, pré-existants)

- `verify:delivery` et 6 tests de la suite client (458 fichiers, 452
  verts) échouent en lisant l'arbre **partagé**
  `C:/Users/Arnaud/opensro/.generated/client-public` : son manifeste
  publié (26 959 assets) n'a pas de `deliveryVersion` — publié par un
  code plus ancien que le validateur d'origin/main. Preuve A-B-A : le
  check échoue à l'identique avec mes changements stashés (baseline
  propre). Aucun des échecs ne lit un fichier touché par M0. La
  républication de l'arbre partagé appartient au checkout principal
  (concurrent actif dessus, cf. journal), pas à cette lane.
- Le client check complet exige `SRO_GAME_ROOT` et
  `SRO_GENERATED_ROOT` dans un worktree (leçons consignées dans la
  mission §0).

## Suite

M1 — bundle de données étendues : build pk2 2026, projection
`.generated/game-data/extended/`, manifest scellé, cap redérivé par la
règle de la chaîne complète (mission §M1).
