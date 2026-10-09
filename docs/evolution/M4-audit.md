# Audit M4 (partie 1) — placements réels et zones dérivées (charter V4)

Session du 2026-10-09 (suite), worktree `opensro-w-evolution`, branche
`evolution/extended-content-m4` (base `evolution/extended-content` @
d373fd0e, merge upstream 2942a6a6). Mission :
`docs/extended-content-mission.md` §M4. Cette livraison couvre la
dérivation **données** du monde étendu ; la lane de rendu/mouvement des
régions 2026 est le reste du jalon (documenté ci-dessous).

## Découverte structurante

Le client live 2026 embarque son propre **`npcpos.txt`** (23 195
ancres, format natif exact : RefObjID, région, x, y, z — coordonnées
locales de région). C'est la source de placements **authoritative**,
meilleure que toute donnée communautaire : **2 014 régions** peuplées
en 2026 contre **1 207** au natif 1.150 → **809 régions nouvelles**,
dont **698 zones extérieures** portant du contenu 91-140.

## Livré

- Extraction : `npcpos.txt` 2026 ajouté (1 133 fichiers extraits).
- Builder (`buildExtendedGameDataBundle.mjs`) :
  - le census characterdata enregistre les mobs par RefObjID (join
    npcpos) ;
  - `buildExtendedAreas` dérive les **vraies zones** : chaque ancre
    npcpos dont la région n'existe pas dans le npcpos natif, dont le mob
    est 91-140, nouveau (absent du characterdata natif, hors familles
    non-combat), devient une ligne de population — groupée en **une
    aire authorée par région réelle** (max 24 ancres/zone, bornes
    0-1920 du contrat d'aire, régions donjon 0x8000 exclues pour cette
    partie extérieure) ;
  - `zones.json` (format `sro-extended-zone-list`) : la liste de zones
    dérivée — régions, ancres, bandes couvertes ;
  - l'aire de démonstration M3 (« extended-content-lab ») est
    **remplacée** par ces zones réelles ;
  - manifeste : `counts.extendedZones` / `extendedAnchors`, scellés
    (manifeste final `72ae3aa9…`, deux runs = même manifeste).
- `devdeps` : la greffe de refs passe par `monster.LoadMonsterRefs`
  seul — le `LoadTemplate` complet sur le textdata 2026 aurait joint
  les 23k ancres aux tables d'évidence v1.188 (panique attrapée par le
  test, corrigée avant tout boot).

## Mesures réelles

- **698 zones réelles** (régions 2026 jamais peuplées par le natif),
  **6 341 ancres npcpos**, bandes couvertes : 91, 101, 111, 121, 131
  (la trajectoire complète 91→140).
- Répartition mesurée des régions : 2 014 (2026) − 1 207 (natif) = 807
  nouvelles peuplées ; 698 extérieures avec du contenu 91-140 après
  filtres (donjons 0x8000, hors-bornes, mobs natifs/non-combat).
- Sample de spawn : chaque zone voit **toute sa population apparaître
  dans sa propre région** via le registre de production
  (`simulation.NewMonsterState` + `AdvancePopulation`).

## Critères d'acceptation — couverts par cette partie

- « Liste de zones dérivée des données, publiée dans le manifeste » :
  `zones.json` + counts scellés ; dérivation = npcpos 2026 ∖ npcpos
  natif, documentée dans le descripteur.
- « Activation des spawns de M3 sur les zones » : les 6 341 placements
  réels remplacent l'aire semée ;
  `TestExtendedZonesSpawnInTheirRealRegions` PASS (toutes les zones,
  toutes leurs populations, vraies régions).
- « Chaque tranche couverte » : `TestExtendedZonesCoverEveryTrajectoryBand`
  PASS (bandes 91..131, trajectoire complète).
- « Flag off » : `TestExtendedZoneMobsAreAbsentFromTheNativeTemplate`
  PASS (aucun mob placé n'existe nativement ; native = inchangé).

## Non couvert (reste du jalon M4)

- **Traversabilité** (E6) : les régions 2026 n'ont pas encore de données
  monde (heightmaps/bundles région + navmesh objet + rendu client) — la
  lane `buildOutdoorWorldRegionResources` (qui parse déjà le `.o2`)
  doit tourner pour ces régions vers l'arbre étendu ; les aires restent
  `access: "gm"` jusqu'à leur ouverture.
- **Donjons** (régions 0x8000) exclus de cette partie ; **téléports**
  (E7) : dépendent des zones servies — après la lane monde.
- Navmesh objet : le client 2026 embarque ses `.nvm` (extraits par
  SRObro) ; la conversion suit la lane monde.

## Tests et gates (sorties de session)

- `go test ./internal/game/enterworld/ -run "ExtendedZones|ExtendedZone"
  -count=1` (projection réelle) → 3 PASS.
- `node --test scripts/test/pipeline/extendedGameDataBundle.test.mjs`
  → 4/4 PASS (fixture npcpos, counts.extendedZones épinglés).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 46.9s`.
- `pnpm task run check:server` (SRO_CHECK_FORCE=1) →
  `server gates: PASS (16 package workers, test cache on, 37.9s)`.

## Écarts connus

- 24 ancres max par zone (borne mesurée pour cette partie ; les zones
  complètes suivent la lane monde — le npcpos complet est extrait et
  scellé).
- Les ancres hors bornes [0,1920) sont écartées (contrat d'aire).
- Merge upstream 2942a6a6 résolu (render scale + onglet Lighting de
  #370 coexistent avec l'onglet Content ; tests exhaustifs mis à jour,
  13/13).

---

# Audit M4 (partie 2) — la lane monde : traversabilité, minimaps, chaîne mouvement

Session du 2026-10-09 (suite), worktree `opensro-w-evolution`, branche
`evolution/extended-content-m4` (base : la lane `evolution/extended-content`
@ `5390a46f`). Mission : `docs/extended-content-mission.md` §M4. Cette
livraison couvre la lane monde que la partie 1 avait laissée : les données
monde 2026 extraites, construites et servies des deux côtés (autorité de
mouvement serveur + rendu/navigation client), plus les minimaps. Les
téléports (E7) et la preuve navigateur de traversée restent (plan en
« non couvert »).

## Livré

- **Extraction** (`scripts/extract_extended_world_data.py`, nouveau) : pour
  les régions de la liste scellée, le plan de données complet depuis les
  pk2 2026 en lecture mmap stricte — Map.pk2 (`.m`/`.t`/`.o2` par région,
  `object.ifo`, `tile2d.ifo`, tout `tile2d/`), Data.pk2 (`nv_*.nvm`,
  `mapinfo.mfo`, `objectstring.ifo`, fermeture transitive des objets :
  composé → BSR → BMT/BMS → DDJ), Media.pk2 (tuiles minimap). Sortie sous
  `<privé>/extended/game/extracted` (contrat game-root/extracted natif,
  provenance machine-indépendante). Un BSR 2026 dont l'entrée mesh porte un
  drapeau u32 a dérailé mon premier port du repli `readBsrMeshEntry` : le
  drapeau se lit à l'offset de DÉBUT d'entrée, corrigé avant tout build.
- **Lane de build** (`scripts/build/world/buildExtendedWorldRegionResources.mjs`
  + CLI `scripts/build_extended_world_resources.mjs`, nouveaux) : 687
  bundles région v5 (même constructeur que le natif), stock d'objets
  partagé étendu (index + meshes adressés par contenu), index
  `/assets/world/extended/world-regions.json`, **catalogue overlay client**
  `/assets/world/extended/world-region-catalog.json` (entrées
  `area: "outdoor"`, `source: "extended-outdoor-live-2026"`), référence au
  shared-render ciel/eau NATIF par chemin public (zéro duplication), tuiles
  minimap publiées au chemin natif + catalogue
  `/assets/data/extended-minimap.json`, **miroir movement** dans la
  projection étendue (`movement/`, projections `projectRegionBundle` /
  object-nav réutilisées du builder natif, exportées pour l'occasion), et
  **manifeste monde scellé** (`world-manifest.json`, digests de contenu).
  Le store d'objets partagés couvre TOUJOURS la liste complète des secteurs
  et se réutilise sinon (un run `--region` partiel ne rétrécit jamais
  l'index — bug attrapé et corrigé pendant la session).
- **Serveur** : la chaîne d'autorité de mouvement
  (`movement/water_extended.go`) — `SetExtendedAuthorityRoot` installe le
  miroir en repli de `surfaceForRegion`, le natif répond toujours en
  premier ; le preload de boot chauffe aussi le miroir ; le câblage passe
  la racine quand la projection porte `movement/catalog.json` (absent :
  avertissement, jamais un effacement silencieux). Les aires étendues
  passent `access: "public"` dans le builder de données (catalogue chargé
  seulement derrière le flag — les grades natifs ne le voient jamais). Le
  builder de données PORTE `movement/` et `world-manifest.json` à travers
  son remplacement d'arbre (le rebuild données avait effacé le miroir —
  attrapé par des tests qui sautaient, corrigé).
- **Client** : `world.ts` fusionne l'overlay catalogue derrière la ligne
  `extendedContent` (annonce serveur = l'arbre servi porte l'overlay ; un
  serveur natif ne le sert pas et le client reste exactement natif — le
  repli de la charte, sans wire nouveau) ; le worker de navigation consulte
  l'overlay (+ catalogue natif) seulement quand le bundle demandé vit sous
  `/assets/world/extended/` ; `hud/minimap.ts` surcharge l'ensemble d'art
  minimap depuis le catalogue étendu ; le servage (`published-assets.mjs`)
  fait union : racine principale d'abord, puis l'arbre privé nommé par
  `SRO_EXTENDED_ASSETS_ROOT` (variable absente = comportement d'aujourd'hui
  bit pour bit).

## Mesures réelles

- **687/698 régions** ont le plan de données complet dans le client 2026 ;
  11 régions (bord est, `0x73cb`…`0x7ecc`) n'ont ni terrain ni navmesh —
  écartées et scellées dans le manifeste ; les bandes restent couvertes
  (zones par bande : 91→133, 101→311, 111→151, 121→203, 131→86).
- **59 035 placements** `.o2` parsés ; **866 ids d'objets** → fermeture
  mesurée `{compound: 69, bsr: 928, bmt: 386, bms: 4269, ddj: 1929}` =
  7 581 fichiers de ressources ; extraction totale ≈ 670 Mo (dont tile2d
  132 Mo), client-public privé ≈ 1,4 Go.
- **Minimaps : 687/687** tuiles extraites (24,7 Mo de DDJ), converties,
  publiées et scellées au catalogue.
- Manifeste monde : **691 digests** ; deux runs complets = octets
  identiques ; `--force --region=0x4939,0x62c9` (reconstruction de zéro de
  deux régions) reproduit les mêmes digests. Manifeste données :
  `ca52b0a8…` (le flip `public` des aires remplace le `72ae3aa9` de la
  partie 1).
- Preload de boot du miroir complet : **8,3 s** (687 bundles, test réel).
- Isolation : le checkout principal ne porte **aucun** fichier étendu
  (vérifié : pas de `assets/world/extended/` ni d'`extended-minimap.json`
  dans son client-public).

## Critères d'acceptation — couverts par cette partie

- « Traversabilité (E6), moitié serveur » : le test réel
  `TestExtendedMirrorWalkableThroughTheChain` charge TOUT le miroir au
  preload et prouve une zone praticable (`SpawnRegionAvailable`) dans
  CHAQUE bande 91→131 ; la chaîne ne change aucune réponse native
  (`TestExtendedChainAbsentKeepsNativeAnswers` : sans chaîne, la région
  2026 reste non couverte ; `TestExtendedChainResolvesLive2026Region` :
  avec chaîne, le natif gagne les conflits).
- « Minimap par zone » : 687 tuiles + overlay client (mêmes chemins et
  même validateur que l'art natif).
- « Flag off » : gates verts sans la chaîne (tous les tests natifs
  passent, la ligne client gate tout, la variable de servage est absente
  par défaut, le principal n'est jamais écrit).
- « Packaging par le pipeline existant » : publication via le ledger
  (`writeIntoPublicTree` / `copyIntoPublicTree`), verrou generated-assets,
  conversion d'images standard.

## Non couvert (reste du jalon M4)

- **Téléports (E7)** : les quatre tables existent dans le Media.pk2 2026
  (`teleportbuilding/data/link`, `siegefortress` — vérifié) et les
  chargeurs natifs sont identifiés (`loadPortalCatalog`,
  `AppendTeleportGates`) ; la greffe derrière le flag (fusion
  natif-d'abord, mondes inconnus du 1.150 écartés avec compteurs) est la
  partie 3.
- **Preuve navigateur** : une traversée réelle d'une zone étendue en flag
  on (browser, capture) — les composants sont testés séparément ; le
  parcours E8 de M5 la couvrira de toute façon bout en bout.
- **Donjons 0x8000** : hors de la liste de zones dérivée (les bandes 91-140
  sont couvertes en extérieur) — reste parqué sauf besoin M5.
- 11 régions sans données de carte dans le client 2026 (aucun téléport ne
  les cible ; documentées au manifeste).

## Tests et gates (sorties de session)

- `go test ./internal/game/world/movement/ ./internal/game/enterworld/
  -run "Extended" -count=1` (projection réelle + extraction native via
  `SRO_GAME_ROOT`) → **10/10 PASS, zéro saut** (3 tests de zones M4p1,
  3 tests de chaîne, 1 test de miroir réel, 3 tests de courbe M2).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 36.7s` (après
  normalisation CRLF d'un fichier réécrit par un script Python de
  découpage — leçon : ne jamais réécrire de source via Python en mode
  texte sur Windows).
- `pnpm task run check:server` (`SRO_CHECK_FORCE=1`, env privé complet) →
  `server gates: PASS (16 package workers, test cache on, 98.4s)` (tests
  98 s, race, govulncheck, release contract).
- `pnpm --filter @sro/client-next check` → `client check: PASS (11 gates,
  93.9s)` (avec `SRO_GAME_ROOT` natif et `SRO_SERVER_GAME_DATA_ROOT`
  pointant la projection 1.150 en fichiers libres matérialisée dans
  l'arbre privé — le test `item-tooltip-magic` lit la projection en
  fichiers libres ; le checkout principal ne porte plus que
  `server.srogz`).

## Écarts connus

- L'env de test Go de la mission §0.4 omettait `SRO_GAME_ROOT` : sans lui,
  les tests licensed sautent silencieusement (« licensed game data is not
  available ») — c'est ainsi que le wipe du miroir par le rebuild données
  a d'abord été manqué. La ligne est ajoutée à la §0.4 par ce commit.
- Le servage union lit l'arbre étendu APRÈS le miss loose+packs du
  principal : un nom identique répond toujours depuis le principal (E1),
  un fichier 2026 nouveau répond depuis le privé.
- Les coordonnées d'entrée des aires restent y=0 (le résolveur de spawn
  en tire la hauteur réelle, comme en M4p1).

