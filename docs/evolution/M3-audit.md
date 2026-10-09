# Audit M3 — contenu : items DG11-14, mobs 91-140, skills (charter V3)

Session du 2026-10-09 (suite), worktree `opensro-w-evolution`, branche
`evolution/extended-content-m3` (base `evolution/extended-content` @
04b8e19c). Mission : `docs/extended-content-mission.md` §M3. Toutes les
mesures sont des sorties réelles de la session.

## Découverte structurante

Les chargeurs natifs (`TextdataItems`, `TextdataSkills`,
`monster.LoadMonsterRefs`) lisent les tables 2026 **telles quelles**
après deux normalisations à l'extraction :

1. **Reshape characterdata** : le fichier 2026 (104 colonnes) est le
   layout v1.150 (120 colonnes) moins six colonnes supprimées aux index
   71-76 et moins la queue 110-119 (mesuré sur MOB_CH_MANGNYANG :
   colonnes 77+ du natif = colonnes 71+ du live, valeurs identiques).
   L'extracteur réinsère six cellules « 0 » à l'index 71 et padde à 120
   — tous les parseurs Go lisent alors les index natifs sans
   modification.
2. **Noms** : le `textdataname.txt` 2026 est un BOM vide ; les noms
   vivent dans `textdata_object_*` (symbole SN_ col 2, anglais col 9)
   pour les objets et `textdata_equip&skill_*` pour l'équipement.
   L'extracteur synthétise le `textdataname.txt` legacy (symbole col 1,
   anglais col 8) à partir des deux familles (19 708 lignes).
   Les shards sont écrits sous leur nom plié minuscule (les glob Go
   sont sensibles à la casse — c'est la casse du natif).

## Livré

- Extraction élargie : 1 132 fichiers (dg, textdataname synthétisé,
  magicoption, skilldata ×379 + virtual, noms des deux familles) ;
  projection scellée enrichie d'un arbre `textdata/` complet (tout
  l'extrait, digests du manifeste) et d'un `areas/catalog.json`.
  Manifeste final `8d08d81136267db35df53e87d8e27c4edaef218e4cdfdf02b1dbb79ff1920f80`.
- `enterworld.OverlayItems` / `OverlaySkills` (port-only, not native) :
  le natif gagne chaque conflit (charte §4.1), les rangs 2026 ajoutent
  ce que le natif n'a jamais eu. Implémentent toutes les surfaces
  consommées (ItemRefSource, CharacterRefSource, ItemCommandReferences,
  SkillDataSource, caches bornés).
- `monster.GraftRefs` + `worldarea.Merge` : les refs monstres 2026 et
  les aires authorées étendues rejoignent le template natif (natif
  gagne ; collision de région/slug refusée).
- Wiring : une seule décision — flag on + projection vérifiée →
  catalogues 2026 chargés (boot en échec si l'itemdata est vide ou les
  skills ne chargent pas), overlays posés sur DevPaths, refs greffées,
  aires fusionnées, caches bornés des deux côtés, seeders et références
  paires sur la vue fusionnée.
- Aire semée `extended-content-lab` (M3) : 5 mobs — un par bande
  91-140, choisis déterministiquement parmi les mobs **absents du
  natif** (le builder charge les codenames characterdata 1.150 pour
  écarter les familles que le natif connaît déjà, p.ex. EU_THIEF) et
  hors familles non-combat mesurées (BOT/GM/EV/TEST/EVENT).

## Mesures réelles

- Items : **21 529/21 529** rangs parsés par le chargeur natif (DG11
  `ITEM_CH_SWORD_11_A` : id 101, ReqLevel 101, atk 1548-1641, dur 151 ;
  DG14 rare : atk 3591-3790, dur 174). Compteur = SRObro items.json.
- Mobs : **7 779** refs monstres 2026 (natif 6 038) ; mob témoin
  `MOB_TQ_SNAKEWOMAN` niveau 97, HP 28 917, walk 40 / run 75 / scale
  100. Absents du natif par bande : 217/239/172/132/87.
- Skills : **36 008** rangs chargés proprement par le chargeur natif —
  exactement le compteur SRObro (`skills_official.json`).
- Aire semée : Qin-Shi (95), **Haroeris** (~105, Égypte), Jupiter
  (~115), Arabia (~125), Sky Temple (~135) — région 0x62a8 (réelle,
  servie, libre dans le catalogue natif).

## Critères d'acceptation (mission M3) — vérifiés

- « Un mob 100+ apparaît en flag on, absent en flag off » :
  `TestExtendedGraftSpawnsEveryBandExemplar` PASS — greffe + aire semée
  → le registre de production (`simulation.NewMonsterState` +
  `AdvancePopulation`) fait apparaître les 5 mobs en région 0x62a8 ;
  `TestExtendedMobsAreAbsentFromTheNativeTemplate` PASS — aucun des 5
  n'existe dans le template natif.
- « Échantillon de stats par degré » : valeurs citées ci-dessus
  (sondes réelles sur la projection).
- « Chaque tranche E3 couverte » : l'aire semée couvre les 5 bandes ;
  la couverture complète par zone arrive en M4 (placements réels).
- Tests des deux réglages : greffes actives seulement derrière
  `paths.Extended*` (posés par le wiring derrière le flag) ; tous les
  tests natifs des packages touchés restent verts.

## Tests et gates (sorties de session)

- `go test ./internal/game/world/monster/ -run GraftRefs` → 2 PASS ;
  `./internal/game/world/worldarea/ -run Merge` → 2 PASS ;
  `./internal/game/enterworld/ -run "ExtendedMobs|ExtendedGraft"`
  (avec projection réelle, `-count=1`) → 2 PASS.
- `node --test scripts/test/pipeline/extendedGameDataBundle.test.mjs`
  → 4/4 PASS (fixture dg.txt incluse).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 10.5s`.
- `pnpm task run check:server` (SRO_CHECK_FORCE=1) →
  `server gates: PASS (16 package workers, test cache on, 36.3s)`
  (tidy, gofmt, vet, golangci-lint, tests, race, govulncheck, release
  contract — pass complet sur le code M3).

## Écarts connus (documentés, non bloquants)

- **Magic options natives** : l'overlay ne touche pas
  `magicoption.txt` — l'alchimie sur items 11D+ utilisera les groupes
  natifs (les groupes 2026 des hauts degrés manquent). Reprise M4/M5
  si l'owner veut l'alchimie complète des hauts degrés.
- **Noms partiels** : les symboles des items non-rare DG11+ n'existent
  pas dans les familles de noms 2026 (les rare sont nommés, p.ex.
  « Ghost Sword ») ; les codenames restent la référence serveur.
- **Sémantique des skills 2026** : les 36 008 rangs chargent et
  résolvent (IDs/learn-plane), mais l'exécution des effets reste le
  plan v1.150 — un skill 2026 à effet inconnu échoue fermé au cast,
  conformément à la posture fail-closed du moteur. Couverture
  d'exécution : M5+.
- **Placements réels** : l'aire semée démontre la mécanique ; les
  placements dérivés des vraies zones (vSRO188/xSROMap) sont M4.
- `pnpm check source` a échoué une fois transitoirement juste après un
  rebuild (course de stamp) ; re-run immédiat PASSED.

## Suite

M4 — monde : zones de la trajectoire 1→140 (heightmaps, placements
`.o2`, navmesh, minimaps, téléports), placements réels des mobs par
zone, remplacement de l'aire semée.
