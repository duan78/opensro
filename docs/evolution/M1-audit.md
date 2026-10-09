# Audit M1 — bundle de données étendues (charter V1)

Session du 2026-10-09, worktree `opensro-w-evolution`, branche
`evolution/extended-content-m1` (base `evolution/extended-content` @
c101ce0c). Mission : `docs/extended-content-mission.md` §M1. Toutes les
mesures ci-dessous sont des sorties réelles de la session.

## Livré

- `scripts/extract_extended_client_data.py` — extraction lecture seule
  (mmap) du client live 2026 (`SRO_EXTENDED_GAME_ROOT`, défaut
  `C:/Program Files (x86)/Silkroad`) : `leveldata.txt`, `levelgold.txt`,
  loaders + shards itemdata/characterdata, vers
  `<generated>/extended/source` avec inventaire scellé (sha256 des cinq
  archives, réutilisé par taille+mtime). Aucun exécutable lancé, aucune
  écriture dans l'install.
- `scripts/build/server/buildExtendedGameDataBundle.mjs` — projection
  `.generated/game-data/extended/` (paths.mjs
  `resolveExtendedGameDataRoot`, override `SRO_EXTENDED_GAME_DATA_ROOT`,
  publication atomique sous le lock generated-assets) : leveldata.json
  (150 lignes, colonnes documentées d'après
  `apps/server/internal/game/enterworld/leveldata.go`), levelgold.json
  (140), census.json (degrés items + bandes mobs), manifest.json scellé
  (5 archives, contentDigest, fichiers par sha256, jamais « 1.150 »).
- Règle de la chaîne complète (charter §4.5) implémentée
  (`deriveCapByCompleteChain`) : cap = min(XP, or, bande d'équipement),
  chaque bande de mobs 91→cap non vide, et REFUS si le résultat n'est pas
  le 140 scellé.
- `gamedata.LoadExtended` (Go) : parse + valide le manifeste (format,
  schéma, sourceClient, cap scellé = `ExtendedLevelCap` 140) et vérifie
  chaque fichier par taille + sha256. États explicites, jamais de repli
  silencieux sur le natif.
- Colonnes découvertes et épinglées en constantes documentées : itemdata
  codename col 2, ReqLevel col 33 (non-rare : DG11→101, DG12→111-118,
  paliers +4 par tier ; rare : 101 plat pour DG11+ ; DG13+ livrés
  rare-only) ; characterdata codename col 2, niveau col 57 (validé sur 5
  mobs témoins de niveau 1→190, niveaux croisés avec la table SRObro).

## Mesures réelles

- Extraction : 750 fichiers, 549 shards itemdata + 197 characterdata ;
  sha256 des archives : Data f594e5e2639d77ce…, Media 95c4034e204de25c…,
  Map d10ea0f2b66eb770…, Particles d482dde8a4fefc18…,
  Music daf0668c58418632…
- Build : `derivedCap: 140 (xp 150, gold 140, gear DG14, bands 19)` ;
  `counts: 21529 item rows, 14642 character rows, 150 levels` ;
  manifest sha256 `48580add63cc257efd8de469e9841fba9b4198df8ccdf3fe96efe8db40b6214a`.
- Reproductibilité (AC) : deux runs complets = manifeste identique
  (`48580add…` avant et après rebuild, voir commandes).
- Cross-check SRObro (AC) : `leveldata.txt: identical`,
  `levelgold.txt: identical` (comparaison sha256 avec leur extraction du
  même client) ; compteurs strictement égaux à leurs tables — items
  21 529 = items.json, characters 14 642 = characters.json.
- Chargement Go de la projection réelle (test jetable non committé) :
  `cap=140 chain=xp:150/gold:140/gear:DG14/bands:19
  counts=levels:150 gold:140 items:21529 chars:14642 archives:5` — PASS.

## Critères d'acceptation (mission M1) — vérifiés

- « Build reproductible » : oui (manifeste identique sur deux runs).
- « Rapport de cross-check » : ce document (identiques/égaux, zéro écart).
- « Flag off = chargement identique au natif » :
  `TestDisabledNeverResolvesAnExtendedPath` + `ResolveExtended` court-circuite
  avant tout accès disque ; le build natif n'est ni lu ni touché.
- « Le cap dérivé = 140 » : build réel + test pipeline + constante Go
  vérifiée par `TestLoadExtendedRefusesAMovedCap`.

## Tests et gates (sorties de session)

- `node --test scripts/test/pipeline/extendedGameDataBundle.test.mjs` →
  4/4 PASS : cap scellé écrit ; or court → `/redelivered cap 130/` ;
  bande 121-130 vide → `/mob band 121-130 is empty/` ; seal dérivé →
  `/drifted from its sealed digest/`.
- `go test ./internal/gamedata/ -v` → 19 PASS (9 étendus + 10 natifs,
  zéro régression), `ok opensro.online/server/internal/gamedata`.
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 35.4s`.
- `pnpm task run check:server` (SRO_CHECK_FORCE=1) →
  `server gates: PASS (16 package workers, test cache on, 11.9s)`
  (tidy, gofmt, vet, golangci-lint, tests, race, govulncheck, release
  contract — pass complet sur le code M1).

## Écarts connus

- Le lock generated-assets a attendu le build du checkout principal
  (pid 8484) avant de builder : coordination OK, aucun conflit.
- `leveldata.json`/`levelgold.json` portent les colonnes brutes ; la
  sémantique par colonne (XP, SP, base mob…) est consommée en M2.

## Suite

M2 — progression 140 : `LevelCap` injecté (90 natif / 140 étendu via
`LoadExtended`), courbes 2026 branchées dans `LevelDataSource`, gel d'XP
au cap identique, franchissement 90→91 impossible hors flag.
