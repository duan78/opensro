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
