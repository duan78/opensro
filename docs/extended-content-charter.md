# PROMPT MAÎTRE — Évolution « contenu étendu » d'opensro

Cœur natif v1.150 inchangé + contenu du client officiel live 2026 (cap 140),
derrière des options désactivées par défaut. Document de pilotage unique de
l'évolution, sur le fork `duan78/opensro` uniquement. Écrit le 2026-10-09 à
partir de l'audit complet des trois sources locales (client 1.150, client
live 2026, projet SRObro) — les preuves chiffrées sont en annexe A.

---

## 0. Comment utiliser ce prompt

Ce fichier est le contrat de l'évolution. Au début de chaque session de
travail (agent ou humain) :

1. Coller ce prompt en tête de session, en précisant la phase visée
   (« on exécute V1 »). Une session = une phase au maximum, jamais deux.
2. Travailler dans le worktree dédié, sur une branche
   `evolution/extended-content-v<N>` découpée depuis `evolution/extended-content`.
3. Terminer la phase par un rapport d'audit `docs/evolution/V<N>-audit.md`
   (faits mesurés, commandes exécutées, écarts constatés — sur le modèle des
   audits chiffrés de SRObro : on ne déclare rien qu'on ne prouve pas).
4. Merger dans `evolution/extended-content` après relecture, pousser sur le
   fork. **Jamais de push vers `origin` (opensro-dev), jamais de PR amont,
   sans demande explicite de l'owner.**

Une phase n'est « faite » que quand tous ses critères d'acceptation (AC)
sont verts et cités dans le rapport.

## 1. La décision

L'audit du 2026-10-09 a établi :

- opensro est un port de fidélité native : la version `1.150` est une
  constante vérifiée cryptographiquement, le cap 90 une règle de code, et
  tout le plan wire (~1400 opcodes côté client, 16+ packages `wire.go` côté
  serveur) est épinglé au décompile du client v1.150. Un rebase complet sur
  le client live détruirait ce modèle de preuve : ce serait un autre projet.
- Le client officiel installé localement est le client **live** (build du
  2026-09-04, Wemade Max), cap réel **140** ; ses archives `.pk2` sont
  lisibles par `scripts/sro_pk2.py` **sans aucune modification** (conteneur,
  clé, encodage inchangés depuis 2005). Le seul vrai mur est le plan wire
  natif du client 2026 (obfusqué, non documenté ici).
- SRObro (MIT) a déjà extrait et converti tout le contenu du client live :
  tables JSON, GLB, terrain, placements, icônes, sons, effets, plus un
  savoir de reverse des formats `.o2`, `.ban`, DDJ, XMX.

Décision : **architecture en couches** (option D de l'audit). Le cœur natif
1.150 reste l'autorité et le défaut bit-identique ; le contenu du client
live 2026 greffe par-dessus, derrière des options off par défaut, en
respectant la règle non-native d'AGENTS.md. Le wire natif reste v1.150 —
aucun système nouveau n'entre sans preuve binaire.

## 2. Règles inviolables

1. **AGENTS.md prime.** Style id Software, bannières, `rg` avant chaque
   édition, worktrees, conventions d'encodage (UTF-8 sans BOM, LF) : tout
   s'applique à cette évolution comme au reste du repo.
2. **Le natif 1.150 est intouchable par défaut.** Flag éteint = comportement
   1.150 cap 90 identique bit pour bit, y compris tests et manifests. Toute
   régression flag-off est un bug bloquant, quelle que soit la phase.
3. **Règle non-native d'AGENTS.md.** Chaque extension passe par : côté
   serveur, un flag d'environnement `SRO_<FEATURE>` dont la valeur absente
   est le natif (pattern exact : `SRO_BETA_GROWTH` dans
   `apps/server/internal/game/progression/growth.go`, banner « port-only,
   not native ») ; côté client, une ligne booléenne dans la fenêtre
   Experimental (`apps/client-next/src/engine/foundation/ui/experimental-options.ts`,
   « only an explicit true enables »). Les deux réglages sont testés.
4. **Marquage de provenance.** Tout code/ donnée issu du client live porte
   la mention `extended content (isro-live-2026), not v1.150-native` dans sa
   bannière ou son manifest. Jamais de mélange sans étiquette.
5. **Aucun asset ni donnée retail dans git.** NOTICE.md l'interdit : le repo
   ne contient pas de média extrait. Les extractions vivent dans
   `.generated/` (racine gérée par `scripts/lib/generatedRoot.mjs` — jamais
   construire un chemin `.generated/...` à la main) ou hors du repo. Les
   fixtures de test restent minimes et dérivées, comme aujourd'hui.
6. **L'install du client live est un temple.** `C:/Program Files (x86)/Silkroad`
   : lecture seule absolue. Ne jamais exécuter `silkroad.exe`,
   `sro_client.exe`, `replacer.exe`, `sro_ad.exe` — même « pour voir ». Ne
   jamais y écrire. Les extractions écrivent dans `.generated/` ou `%TEMP%`.
   Le launcher updater (`TempPath/replacer.exe` chiffré) rendrait en plus
   l'install non reproductible.
7. **SRObro se lit, ne se copie pas.** Ses faits, ses tables et ses
   conversions servent d'accélérateur et d'étalon de vérification croisée.
   Toute reprise de code est une réécriture dans le style id, avec
   attribution ajoutée dans NOTICE.md (section Origin) — son code est MIT.
8. **Pas de wire nouveau sans preuve.** L'espace d'opcodes reste v1.150.
   Voir phase V5.
9. **Les fichiers SHA256-pinnés ne se modifient pas** (liste dans
   `apps/server/AGENTS.md`). Note machine : `../research` est absent de ce
   poste, les `verify_*.py` ne peuvent pas tourner — n'approche donc pas ces
   fichiers du tout.

## 3. Ressources locales (inventaire exact)

| Ressource | Chemin | Usage |
| --- | --- | --- |
| Client v1.150 (source native) | `C:/Users/Arnaud/Downloads/02 ISRO - Legend 3 (v150)` | `SRO_GAME_ROOT` ; source de la projection native 1.150, inchangée |
| Client live 2026 | `C:/Program Files (x86)/Silkroad` | Source de la couche étendue. Lecture seule. PK2 lus par `scripts/sro_pk2.py` tel quel |
| SRObro (MIT) | `C:/Users/Arnaud/Desktop/SRObro` | Tables JSON de référence, conversions web-ready, savoir RE, docs. Détail ci-dessous |
| Worktree évolution | `C:/Users/Arnaud/opensro-w-evolution` | Branche `evolution/extended-content`. Assets partagés : `SRO_GENERATED_ROOT` vers le `.generated` du checkout principal |

Dans SRObro, les points d'appui (chemins exacts, vérifiés le 2026-10-09) :

- Tables du client live déjà importées en JSON :
  `server/data/game/items.json` (21 529 items), `characters.json`
  (14 642 personnages), `monsters_official.json` (6 483 monstres),
  `skills_official.json` (6 909 skills) — compteurs revérifiés le
  2026-10-09 par mesure directe. Parseur :
  `server/scripts/import-textdata.ts` (colonnes `_RefObj*` vérifiées
  empiriquement).
- Extractions brutes du client live :
  `assets/pk2_media/` (textdata UTF-16LE, icônes, minimaps),
  `assets/pk2_data/` (meshes `prim/`, navmesh `.nvm` + `object.ifo`),
  `assets/pk2_map/` (heightmaps + `.o`/`.o2` par région),
  `assets/pk2_particles/`.
- Monde converti : `client/public/assets/terrain/` — 6 046 régions
  (`.f32` + `.tiles`), `objects.json` (59 115 placements `{bsr,x,y,z,yaw,scale}`),
  `regions.json`, `tile-index.json`.
- Assets web-ready : `glb_blender` (18 648 GLB), `textures` (61 031),
  `icons` (7 931 DDJ→PNG), `audio` (2 932), `effects` (3 252 `.efp` décodés,
  `efp-descriptors.json`), `anims` (3 916 clips), `skeletons`.
- Savoir RE : `scripts/parse-v7-o2.cjs` (grammaire `.o2`),
  `docs/BAN_FORMAT_DOCUMENTATION.md` + `ban-re/` (format `.ban`),
  `docs/XMX_SOLUTION_FOUND.md`, `tools/rust-jmx-converter/` (BMS/BMT/BSK→GLB,
  DDJ→PNG), `docs/GLB_FORMAT_ISSUES.md`.
- Base documentaire : `docs/SRO_KNOWLEDGE_BASE/` (76 fichiers : skills
  chinois/européens, bestiaire, 697 PNJ, 161 téléporteurs, formules).

Dans opensro, les points d'accroche du pipeline (tous v1.150 aujourd'hui) :

- `scripts/sro_pk2.py` — lecteur PK2 (générique ; la clé est la clé Joymax
  multi-versions, vérifiée sur le client 2026).
- `scripts/prepare_client_resources.py` — extraction pk2 → `extracted/`.
- `scripts/build/world/paths.mjs` — toutes les racines (`SRO_GAME_ROOT`,
  `resolveServerGameDataRoot`, `clientV150ResinfoRoot`).
- `scripts/build/server/buildServerGameDataBundle.mjs` — projection serveur
  vérifiée (`GAME_VERSION = "1.150"`).
- `scripts/build/char/buildLevelDataAsset.mjs` — `leveldata.txt` → asset client.
- `apps/server/internal/gamedata/{bundle,resolve,archive}.go` — chargement
  manifest (`SupportedGameVersion = "1.150"`, refus des autres versions).
- `apps/server/internal/game/progression/levelup.go` — `LevelCap = 90`,
  gel d'XP au cap ; injecté via `wiring_authority.go`.
- `apps/server/internal/game/enterworld/leveldata.go` — `LevelDataSource` ;
  le `leveldata.txt` 1.150 contient déjà 140 lignes (50 au-delà du cap).

## 4. Architecture cible

### 4.1 Deux mondes, un binaire

Le serveur et le client compilés restent uniques. Au démarrage, l'état des
flags décide : natif (aucun flag) ou étendu. Le mode étendu **ajoute** des
données ; il ne remplace jamais les tables natives. Si une donnée étendue
entre en conflit avec une table native, la table native gagne et l'écart est
journalisé au build.

### 4.2 Flags

- Serveur : `SRO_EXTENDED_CONTENT` (master, pattern `growth.go` : absent/off
  = natif). Sous-flags par phase si le besoin apparaît
  (`SRO_EXTENDED_LEVEL_CAP`, etc.), jamais de sous-flag qui agit quand le
  master est off.
- Client : ligne `extendedContent` dans `ExperimentalOptions` (booléen
  explicite seulement). Le client étendu ne réclame du contenu étendu que si
  le serveur l'annonce ; un client étendu sur un serveur natif retombe sur
  l'affichage natif.

### 4.3 Projection de données étendue

- Nouvelle racine `.generated/game-data/extended/` (produite par un nouveau
  script `scripts/build/server/buildExtendedGameDataBundle.mjs`), distincte
  de `game-data/1.150/` qui reste l'autorité native inchangée.
- Source : chemin du client live passé par `SRO_EXTENDED_GAME_ROOT` (défaut
  `C:/Program Files (x86)/Silkroad`), lu par `sro_pk2.py`.
- Le manifest étendu porte : `sourceClient: "isro-live-2026"`, le sha256 des
  cinq `.pk2` sources (reproductibilité), ses compteurs, et une version de
  schéma propre. Il ne mentionne jamais `1.150`.
- `SupportedGameVersion` et le manifest natif ne changent pas ; le chargeur
  Go (`gamedata`) lit la projection étendue uniquement quand le flag est on.
  Flag off = aucun fichier étendu ouvert (testé).

### 4.4 Parseurs : ce qui change côté données 2026

Mêmes noms de fichiers, même UTF-16LE, mais : éclatement en centaines de
fragments (`characterdata_*.txt` ×203, `itemdata_*.txt` ×555,
`skilldata_*.txt` ×379, `skilldata_r_*.txt` ×376 « rebirth ») piloté par les
fichiers-index racines ; colonnes ajoutées (`leveldata` 16 colonnes vs 10) ;
nouveaux fichiers (achievements, forteresses, gacha, worldmap…). Les
parseurs étendus tolèrent les colonnes inconnues (lecture par nom quand
l'index le permet, sinon position avec assertion de schéma) et échouent au
build — jamais en silence — sur un schéma qui a glissé.

### 4.5 Progression

- `LevelCap` devient une valeur injectée : 90 en natif (inchangé, gel d'XP
  au cap identique), 140 en étendu.
- **Règle du cap (décision owner du 2026-10-09)** : le cap étendu est le
  plus haut niveau auquel la chaîne officielle est **complète** — courbe XP
  (`leveldata`), courbe d'or (`levelgold`), équipement, contenu de mobs.
  Vérifié le 2026-10-09 sur le client live : la courbe XP va jusqu'à 150,
  mais l'or s'arrête à 140 ; l'équipement jouable complet s'arrête au degré
  14 (les degrés 15-17 existent mais sont pré-provisionnés, même pattern
  que le v1.150 qui provisionnait 140 pour un cap 90) ; 101 mobs au niveau
  140 contre 1 seul au niveau 141 ; le dernier cap officiel annoncé est
  Lv.140. **Cap étendu = 140.** Le build V1 recalcule cette intersection
  et refuse un cap sans chaîne complète ; quand un futur client complétera
  les niveaux 141+, la règle relèvera le cap sans changer de doctrine.
- Courbe XP : le natif continue d'utiliser le `leveldata.txt` 1.150 (qui va
  déjà à 140 — aucune donnée nouvelle nécessaire pour la courbe elle-même) ;
  l'étendu utilise le `leveldata.txt` 2026 (16 colonnes) de la projection
  étendue, recoupé avec `levelgold.txt` 2026.
- Maîtrises, plafonds de groupe, pénalités : natif = valeurs 1.150 actuelles ;
  étendu = valeurs des tables 2026 si elles en contiennent, sinon valeur
  native + écart documenté dans le manifest étendu.

### 4.6 Contenu (items, mobs, skills)

Degrés 11–14, mobs et skills post-90 : importés dans la projection étendue
depuis les textdata 2026, avec vérification croisée obligatoire contre les
JSON SRObro (`items.json`, `monsters_official.json`, `skills_official.json`).
Tout écart de compteur ou de valeur est documenté dans le rapport de build ;
on ne « corrige » pas SRObro ni les textdata, on explique l'écart
(arrondi, filtrage, fragmentation).

### 4.7 Assets et monde

- Le pipeline opensro reste souverain : il extrait lui-même depuis les pk2
  2026 vers `.generated/`. Les conversions SRObro servent d'étalon local et
  d'accélérateur (par exemple pour valider un décodeur, ou comparer un
  compte de placements), pas de source committée.
- Zone par zone : heightmaps, placements `.o2`, navmesh `.nvm` du client
  2026. La grammaire `.o2` est déjà documentée côté SRObro
  (`scripts/parse-v7-o2.cjs`, `docs/SRO_KNOWLEDGE_BASE/opensro/o2-placement.md`)
  — le décodeur opensro est réécrit dans le style du repo, avec attribution.
- Flag off : le monde rendu est exactement le monde 1.150 d'aujourd'hui.

### 4.8 Wire

Inchangé. L'espace d'opcodes, les layouts de paquets et le handshake
restent le plan v1.150. Le contenu étendu voyage sur les opcodes existants
(item/mob/skill des données étendues = mêmes formats). Voir V5 pour tout
système réellement nouveau.

## 5. Phases

### V0 — Charte, flags, squelette

- Cette charte relue et poussée ; flags `SRO_EXTENDED_CONTENT` (serveur) et
  `extendedContent` (client) créés **off**, sans aucun effet mesurable ;
  tests des deux réglages ; fenêtre Experimental affiche la ligne.
- AC : `pnpm check source` vert ; gate serveur Go verte ; gate client
  verte ; un test prouve que flag off n'ouvre aucun chemin étendu.

### V1 — Pipeline de données étendues

- `buildExtendedGameDataBundle.mjs` : lit les pk2 2026, produit la
  projection `.generated/game-data/extended/` + manifest scellé (sha256 des
  pk2 sources, compteurs). Parseurs fragments + 16 colonnes.
- Vérification croisée SRObro : compteurs (items 21 529, characters
  14 642, monsters 6 483, skills 6 909) et sondages de valeurs ; écarts
  documentés.
- AC : build reproductible (deux runs = même manifest) ; rapport de
  cross-check dans `docs/evolution/V1-audit.md` ; flag off = chargement
  serveur identique au natif (test) ; `pnpm check source` vert.

### V2 — Progression étendue

- Cap injecté, ext = 140 ; courbe 2026 en mode étendu ; gel d'XP au cap
  inchangé en natif ; franchissement 90→91 impossible hors flag.
- AC : tests Go des deux réglages (courbe, gel, franchissement) ; audit V2
  avec la table des niveaux 89–91 et 139–140 des deux sources ; gates vertes.

### V3 — Contenu jouable étendu

- Items/mobs/skills degrés 11–14 en flag on : stats, loot, spawns,
  affichage HUD/icônes (DDJ du client 2026 via pipeline).
- AC : test de stats dérivées sur un échantillon par degré ; un mob > 90
  apparaît et combat en flag on, absent en flag off ; une icône degré 11+
  s'affiche ; audit V3.

### V4 — Monde étendu

- Une première zone post-1.150 (proposée : Alexandria — données terrain
  complètes, contenu lié au cap) : heightmaps, placements, navmesh.
- AC : build assets local réussi ; zone traversable en flag on (collision,
  navmesh, minimap) ; monde 1.150 strictement identique en flag off (hash
  des assets servis) ; audit V4 avec captures.

### V5 — Systèmes nouveaux (optionnel, sur décision owner uniquement)

- Tout système qui n'existe pas en v1.150 (battle arena, forteresse moderne,
  skills rebirth, gacha…) exige d'abord une étude de preuve wire dédiée
  (revue du client 2026 ou evidence communautaire recoupée). Sans preuve :
  pas d'implémentation. Cette phase ne démarre que sur demande explicite.

## 6. Vérification et preuves

| Périmètre | Commande |
| --- | --- |
| Politique source | `pnpm check source` |
| Serveur Go | voir `apps/server/AGENTS.md` |
| Client | `pnpm --filter @sro/client-next check` |

- Chaque phase produit `docs/evolution/V<N>-audit.md` : commandes réellement
  exécutées, sortie citée, chiffres mesurés. On n'écrit « vert » que pour ce
  qui a exité 0 dans la session.
- Journal de coordination : une ligne par session dans `~/.sro-coordination`
  (pas de run lock ; prévenir avant de réécrire un fichier partagé).
- Les deux réglages de chaque flag sont testés (règle AGENTS.md).

## 7. Interdits (rappel exécutoire)

- Ne pas exécuter, patcher, ni écrire dans `C:/Program Files (x86)/Silkroad`.
- Ne pas committer assets, textdata ou tables retail (NOTICE.md) ; `.generated/` only.
- Ne pas pousser sur `origin`, ne pas ouvrir de PR amont, sans demande.
- Ne pas modifier les fichiers SHA256-pinnés ni le manifest natif 1.150.
- Ne pas introduire d'opcode/wire non prouvé.
- Ne pas copier-coller de code SRObro : réécrire + attribuer.

## 8. Décisions actées et questions ouvertes pour l'owner

Acté le 2026-10-09 par l'owner : **cap étendu = 140**, le plus haut niveau
à chaîne officielle complète (règle et preuves en §4.5).

Questions restantes :

1. Zone prioritaire V4 : **Alexandria** (proposée), Constantinople, autre ?
2. Degrés 11–14 d'un bloc en V3, ou progressif (11 puis 12…) ?
3. Skills « rebirth » (`skilldata_r_*`) dans le scope V3 ou reportés ?
4. Pets seconde génération / fellowship : reportés à V5 ?

## Annexe A — Faits établis par l'audit du 2026-10-09

| Fait | Preuve |
| --- | --- |
| L'install Program Files est le client live iSRO Global | `sro_client.exe` signé Wemade Max, build PE 2026-09-04, certificat émis 2026-08-13 ; installé 2026-10-01 |
| Cap réel 140 (max exploitable) | Mesuré 2026-10-09 sur les données : `leveldata.txt` 1→150 mais `levelgold.txt` 1→140 ; équipement DG14 ×66 complet, DG15-17 ×66 chacun pré-provisionnés ; 101 mobs au niveau 140, 1 au niveau 141 ; dernier cap officiel annoncé = Lv.140 (15 sept.) ; pattern historique v1.150 (tables à 140 pour un cap 90) |
| Conteneur pk2 inchangé | `scripts/sro_pk2.py` lit l'index du Media.pk2 2026 sans modification (29 292 fichiers listés) ; même clé, même UTF-16LE |
| Volume 2026 vs 1.150 | Data 3,24 Go / 64 861 fichiers (×2,4) ; Media 927 Mo / 29 292 (×3,2) ; Map 1,30 Go / 19 587 |
| Schémas étendus | `characterdata_*` ×203, `itemdata_*` ×555, `skilldata_*` ×379, `skilldata_r_*` ×376 ; leveldata 16 colonnes |
| Couplage opensro 1.150 | `bundle.go` `SupportedGameVersion = "1.150"` (manifest crypto-vérifié) ; ~1400 opcodes épinglés client ; 16+ packages wire serveur |
| Cap 90 = code, pas données | `levelup.go:36` `LevelCap = 90` ; le `leveldata.txt` 1.150 (déjà extrait, déjà chargé) couvre 1→140 |
| SRObro : MIT, complet côté assets | Tables JSON (21 529 items / 14 642 personnages / 6 483 monstres / 6 909 skills, comptés 2026-10-09), 18 648 GLB, 59 115 placements `.o2`, 76 fichiers de docs ; dernier commit 2026-10-04 |
| SRObro : pas de wire natif | Transport Socket.io maison ; aucune connaissance opcode du client 2026 à reprendre |
| `../research` absent de ce poste | `apps/server/AGENTS.md` attend `../research/tools/re/verify_*.py` ; le dossier n'existe pas — les fichiers pinnés ne sont pas modifiables ici |
