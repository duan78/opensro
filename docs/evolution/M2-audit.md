# Audit M2 — progression 140 (charter V2)

Session du 2026-10-09, worktree `opensro-w-evolution`, branche
`evolution/extended-content-m2` (base `evolution/extended-content` @
582e7608). Mission : `docs/extended-content-mission.md` §M2.

## Livré

- `progression.Runtime.LevelCap` (zéro = natif 90, pattern `Growth`) :
  `walkExpCurve` reçoit le cap effectif et son gel
  « requirement−1 » s'applique au cap en vigueur. La constante
  `LevelCap = 90` ne bouge pas ; le mode étendu lève le cap au wiring.
- `enterworld.ExtendedLevels` (extended_levels.go, port-only, not
  native) : la courbe du client live 2026 depuis la projection scellée —
  mêmes sémantiques de colonnes que le lecteur natif (niveau 0, XP 1, SP
  2, base mob 5, jobs 6-8, mesurées sur le fichier livré), base d'or
  depuis le `dg.txt` 2026 (le même rôle que le lecteur natif
  CDropGoldData). Implémente `LevelDataSource` + `JobLevelDataSource` +
  la base d'or : remplace la table native sans changer un appelant.
  Échec de parse = source empoisonnée (fail-closed), jamais un
  demi-chargement.
- Extraction/build : `dg.txt` ajouté à l'extraction et à la projection
  (`goldcurve.json`, scellé au manifeste — nouveau manifeste
  `3c3c5d1d…`, rebuild reproductible, cap toujours dérivé 140).
- Wiring (une seule décision) : `wiring.go` charge la projection scellée
  quand le mode est on (boot en échec sur manifeste dérivé/drifté),
  construit `ExtendedLevels` + cap et les pose sur `DevPaths`
  (`LevelOverride`, `LevelCap`) → la courbe remonte partout : marche de
  progression (`stats.LevelCap`), pourcentage d'XP de présentation,
  bornes de création (`agentapi.Config.LevelCap` via
  `effectiveLevelCap`). Mode off ou projection absente : rien ne change
  (warning de boot M0 inchangé).

## Mesures réelles (données du disque, session)

- leveldata 2026 : 150 lignes ; niveau 1 → 118 XP, niveau 90 →
  200 532 065 (courbe 2026 plus raide que la 1.150 : 281 672 373),
  niveau 140 → 578 982 029 973 906, base mob col 5 identique au natif
  (24/94/259/6949/30462), jobs col 6-8 identiques (70875×3 puis -1).
- dg.txt 2026 : 140 lignes, or 28 (niv. 1) → 515 (niv. 140).
- Probe réel (test jetable non committé) : `ExtendedLevels` sur la
  projection réelle → ExpRequired(140)=578982029973906,
  MonsterExpBasis(90)=6949, WithdrawalGoldBasis(140)=515, 150 lignes.

## Critères d'acceptation (mission M2) — vérifiés

- « Cap injecté, gel identique en natif » :
  `TestNativeWalkFreezesBelowNinety` (gel à requirement−1 au 90 natif,
  zéro injection) PASS.
- « Franchissement 90→91 impossible hors flag » : le cap n'est levé que
  par le wiring derrière `SRO_EXTENDED_CONTENT` + projection scellée ;
  `TestExtendedCapWalksPastNinetyAndFreezesAtOneForty` (cap 140 : 50
  franchissements, gel à requirement−1 à 140) et
  `TestExtendedCapRefusesAboveItsCurve` (courbe courte → refus
  fail-closed) PASS.
- « Courbes 2026 en étendu » : source `ExtendedLevels` branchée au
  wiring + tests fixture (valeurs réelles du fichier) + probe réel.
- « Tables 89-91 et 139-140 des deux sources » : valeurs citées
  ci-dessus (2026) ; natif documenté dans leveldata.go (90 →
  281 672 373) et inchangé (const + tests natifs verts).

## Tests et gates (sorties de session)

- `go test ./internal/game/progression/` → `ok` (suite complète, y
  compris les tests natifs du gel).
- `go test ./internal/game/enterworld/` → `ok` (3 tests ExtendedLevels :
  courbe réelle, fichier absent, JSON cassé).
- `go test ./internal/gamedata/` → `ok`.
- `node --test scripts/test/pipeline/extendedGameDataBundle.test.mjs`
  → 4/4 PASS (après ajout de dg.txt à la fixture).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 11.7s`.
- `pnpm task run check:server` (SRO_CHECK_FORCE=1) →
  `server gates: PASS (16 package workers, test cache on, 50.0s)`
  (tidy, gofmt, vet, golangci-lint, tests, race, govulncheck, release
  contract).

## Écarts connus

- Piège PATH vécu deux fois (leçon déjà consignée mission §0.4) :
  `pnpm check source` sans Go sur le PATH fait échouer `check-server` en
  0,7 s ; avec PATH, tout passe. Les sorties citées sont celles avec
  PATH.
- Maîtrises : le cap natif de maîtrise (120, plafonds Chine 300/Europe
  240) est inchangé en M2 — les tables 2026 les couvrent déjà
  (reqMasteryLv max = 120 mesuré) ; reprise en M3 si les données 2026
  en contiennent d'autres.

## Suite

M3 — contenu : items DG11-14, mobs 91-140 (stats characterdata, spawns
vSRO188/xSROMap), skills jusqu'à maîtrise 120, derrière le même flag.
