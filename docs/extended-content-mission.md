# PROMPT DE MISSION — opensro jouable jusqu'au niveau 140 (fork)

Mission en une phrase : sur le fork `duan78/opensro`, branche
`evolution/extended-content`, rendre **tout jouable jusqu'au niveau 140** —
mobs, équipements, compétences, cartes, téléports, intégré de bout en bout —
à partir du contenu officiel du client live 2026, sans jamais changer le
comportement natif v1.150 cap 90 quand l'option est éteinte.

Ce document est **l'ordre d'exécution**. Le contrat de règles reste
[`docs/extended-content-charter.md`](extended-content-charter.md) (la
« charte ») : mode d'emploi des sessions, règles inviolables, architecture
cible, inventaire des ressources. En cas de conflit, la charte gagne ; si la
mission contredit la charte, corriger la charte d'abord, par commit.

Rédigé le 2026-10-09, après décision owner. Les preuves chiffrées citées
ici viennent de l'audit du 2026-10-09 (annexe A de la charte) et des
mesures du même jour rappelées en §4.

---

## 0. Protocole de session (à lire à chaque reprise)

1. Travailler dans le worktree `C:/Users/Arnaud/opensro-w-evolution`
   (branche `evolution/extended-content`). Un jalon M<N> = une branche
   `evolution/extended-content-m<N>`, mergée dans la lane après audit.
   Un session = un jalon au maximum. Ne jamais travailler dans l'arbre
   principal (`C:/Users/Arnaud/opensro`, WIP concurrent possible).
2. Avant de coder : relire la charte (règles inviolables §2), ce prompt
   (jalon courant), et le dernier `docs/evolution/M<N-1>-audit.md`.
3. Fin de session : rapport `docs/evolution/M<N>-audit.md` (commandes
   réellement exécutées avec leur sortie, chiffres mesurés, écarts), commit,
   push sur le fork (`git push fork`), une ligne dans le journal
   `~/.sro-coordination/append_only.txt`.
4. Environnement — **isolation des données** (décision owner
   2026-10-09) : le fork ne doit RIEN écrire dans le checkout principal
   `C:/Users/Arnaud/opensro`. Toute donnée étendue vit dans l'arbre
   privé `C:/Users/Arnaud/opensro-evolution-data` (extraction, bundle
   natif copié, projection étendue, caches). Le principal n'est lu
   qu'en lecture seule (assets client-public construits par sa propre
   lane).
   - `SRO_GAME_ROOT=C:/Users/Arnaud/Downloads/02 ISRO - Legend 3 (v150)`
     (client natif 1.150) ;
   - `SRO_EXTENDED_GAME_ROOT=C:/Program Files (x86)/Silkroad` (client live
     2026 — **lecture seule, exécutables interdits**) ;
   - Builds (extracteur + builder) :
     `SRO_GENERATED_ROOT=C:/Users/Arnaud/opensro-evolution-data/generated`
     et
     `SRO_EXTENDED_GAME_DATA_ROOT=C:/Users/Arnaud/opensro-evolution-data/server-game-data/extended` ;
   - Tests/gates Go :
     `SRO_EXTENDED_GAME_DATA_ROOT=C:/Users/Arnaud/opensro-evolution-data/server-game-data/extended`
     et
     `SRO_SERVER_GAME_DATA_ROOT=C:/Users/Arnaud/opensro-evolution-data/server-game-data/1.150/server.srogz`
     (copie privée du bundle natif : même les caches de matérialisation
     restent privés ; PAS de `SRO_GENERATED_ROOT` ici, les tests lisent
     le client-public du principal en lecture seule) ;
   - Go doit être sur le PATH AVANT `pnpm check source` (leçon vécue
     2026-10-09 : un build peut « sortir 0 » en ayant échoué — vérifier
     les manifestes produits, pas les exit codes).
5. Fork uniquement. Pas de push `origin`, pas de PR amont, pas de merge
   depuis `origin/main` sans demande owner (sync fork déjà faite au
   2026-10-09).

## 1. Décisions actées par l'owner (2026-10-09)

1. **Cap étendu = 140** (règle de la chaîne complète, charte §4.5 : plus
   haut niveau où XP + or + équipement + mobs existent tous ; mesuré :
   `levelgold` s'arrête à 140, DG14 = dernier degré jouable, 101 mobs au
   niveau 140 contre 1 au 141).
2. **Degrés 11–14 d'un bloc** : tout l'équipement 11D→14D entre en M3,
   pas progressivement.
3. **Toutes les zones de la trajectoire 1→140** : le monde étendu couvre
   chaque zone nécessaire pour qu'un joueur monte 1→140 sans trou de
   contenu, pas seulement une zone témoin.
4. Hors scope (parkés, ne pas y toucher sans nouvel ordre) : skills
   « rebirth » (`skilldata_r_*`), pets/fellowship 2ᵉ génération, battle
   arena, forteresses modernes, gacha, tout nouveau wire/opcode (charte
   §4.8), quêtes post-90 (voir M5 : uniquement si la charge le permet).

## 2. Définition de « jouable jusqu'à 140 » — le théorème d'acceptation

La mission est finie quand ces dix propriétés sont **toutes** vérifiées et
citées dans `M6-audit.md` :

- **E1 Natif intact** : sans aucun flag, comportement 1.150 cap 90
  identique bit pour bit (tests existants verts, monde et assets identiques
  par hash).
- **E2 Progression** : flag on, un personnage gagne ses niveaux 91→140 sur
  la courbe XP du `leveldata.txt` 2026 (16 colonnes), l'or suit
  `levelgold.txt` 2026 (1→140), le gel d'XP au cap fonctionne à 140 comme
  il fonctionne à 90 en natif.
- **E3 Mobs** : chaque tranche de 10 niveaux (91-100, 101-110, 111-120,
  121-130, 131-140) a ses mobs de terrain spawnés, avec stats officielles
  (characterdata 2026 : HP, atk, def, XP), variantes champion/giant et
  aggro native. Mesure d'appui : 535/577/412/416/415 mobs par tranche dans
  les données (2026-10-09) — le compteur spawné doit couvrir chaque
  tranche, à exporter dans l'audit.
- **E4 Équipement** : items 11D→14D droppables/équipables avec stats
  officielles, icônes DDJ 2026 affichées, degrés↔niveaux dérivés
  d'`itemdata` (pas d'un tableau codé en dur).
- **E5 Skills** : les skills requis pour jouer 91→140 (jusqu'à
  `reqMasteryLv` 120) fonctionnent : apprentissage, cast, dégâts,
  animations quand disponibles.
- **E6 Cartes** : chaque zone de la trajectoire est rendue et traversable :
  terrain (heightmaps), collisions, navmesh, minimap, placements d'objets
  `.o2`. La liste exacte des zones est **dérivée des données** (mobs par
  région), pas d'une liste écrite à la main.
- **E7 Téléports** : les PNJ téléporteurs et portes vers les zones
  étendues existent (données de référence : 161 téléporteurs documentés,
  xSROMap), voyage aller/retour testé.
- **E8 Intégration** : un parcours scripté serveur simule tuer→XP→loot→
  équipement→niveau suivant, de 1 à 140, sans blocage (test automatisé).
- **E9 Propreté** : zéro asset/donnée retail dans git ; tout vit dans
  `.generated/` ; chaque bloc étendu marqué `extended, not v1.150-native` ;
  attributions SRObro dans NOTICE.md.
- **E10 Performances** : démarrage client et mémoire dans le même ordre de
  grandeur qu'aujourd'hui en flag off ; en flag on, chargement de zone
  étendue raisonnable (mesuré dans l'audit, comparé à la baseline 1.150).

## 3. Sources de vérité (ordre d'autorité) et provenance

1. **Client live 2026** (`SRO_EXTENDED_GAME_ROOT`) via
   `scripts/sro_pk2.py` (lecture mmap, rien n'écrit dans l'install) :
   autorité pour les *définitions* — `leveldata.txt` (1→150),
   `levelgold.txt` (1→140), `characterdata_*.txt` (×203 fragments),
   `itemdata_*.txt` (×555), `skilldata_*.txt` (×379), `textzonename_*`,
   icônes, minimaps, heightmaps/placements par Map.pk2.
2. **DB serveur vSRO 1.188 (évidence, pas import)** : le client retail ne
   livre pas les tables de spawn serveur. SRObro en a extrait des CSV
   (`C:/Users/Arnaud/Desktop/SRObro/docs/SRO_KNOWLEDGE_BASE/ML_RESEARCH/data/`) :
   `monsters_vsro188.csv` (7 157 monstres, stats serveur),
   `uniques_vsro188.csv` (830), `zones_vsro188.csv` (agrégats par zone).
   Usage opensro : comme le repo traite déjà `Tab_RefTactics` (caravanes,
   `apps/server/internal/game/world/monster/caravan.go`) — données dérivées
   committées en TSV d'évidence sous `apps/server/internal/game/world/monster/data/`,
   avec sha256 de la source dans l'en-tête, jamais un import direct.
3. **Placement géographique 91→140** : `MONSTERS_SPAWN_LOCATIONS.md` de
   SRObro (coordonnées xSROMap/rev6 + sections « zones de spawn KSRO
   106-140 » et « leveling zones »), recoupées avec les minimaps 2026
   (`minimap/` + `minimap_d/` du Media). Chaque zone livrée documente ses
   sources de placement dans son entrée de manifest étendu.
4. **SRObro tables JSON** (`server/data/game/`) : étalon de vérification
   croisée (items 21 529 / characters 14 642 / monsters 6 483 /
   skills 6 909 — compteurs mesurés 2026-10-09).
5. **Web officiel** en dernier recours pour arbitrage de règle (jamais
   pour des valeurs de données).

Règle de provenance : chaque bloc de l'extended porte sa source dans le
manifest (pk2 sha256, CSV d'évidence, doc de placement). Un fait sans
source citée = un bug.

## 4. Rappels de mesures déjà faites (ne pas re-mesurer, sauf divergence)

- `leveldata.txt` 2026 : 150 lignes, niveaux 1→150, 14 colonnes tabulées
  (mesuré ; le v1.150 en a 10) ; valeur XP (col. 2) au niveau 140 =
  578 982 029 973 906.
- `levelgold.txt` 2026 : 140 lignes (1→140), s'arrête là — c'est le plafond
  de la chaîne complète.
- Items : degrés présents 1→17 (+21) ; DG14/15/16/17 = 66 items chacun ;
  14D = dernier jouable, 15-17 = pré-provisionnés (hors scope).
- Mobs : 101 au niveau 140, 1 au 141 ; 64 entrées >140 = bosses de raid
  (jusqu'à 190) — intégrables comme contenu de zone à la discrétion de M3,
  jamais comme justification de cap.
- Skills : `reqMasteryLv` max = 120 ; le champ `level` (max 30) est le
  niveau du skill, pas celui du personnage requis.

## 5. Plan d'exécution

Chaque jalon finit par ses AC citées dans son audit. Les jalons reprennent
les phases V0→V4 de la charte et les complètent (V0→M0, V1→M1, V2→M2,
V3→M3, V4→M4) ; M5 et M6 sont les ajouts de cette mission (intégration
bout-en-bout et clôture).

### M0 — Squelette de flags (charte V0)

Tâches : `SRO_EXTENDED_CONTENT` serveur (pattern `growth.go`), ligne
`extendedContent` client (pattern `experimental-options.ts`), tests des
deux réglages, zéro effet mesurable.

AC : gates vertes (`pnpm check source`, Go, client) ; test prouvant que
flag off n'ouvre aucun chemin étendu ; la fenêtre Experimental affiche la
ligne off.

### M1 — Bundle de données étendues (charte V1)

Tâches : `scripts/build/server/buildExtendedGameDataBundle.mjs` ; lecture
pk2 2026 ; parseurs fragments + colonnes étendues (14 mesurées sur
`leveldata` 2026 vs 10 en 1.150 ; tolérance aux colonnes inconnues, échec
au build sur schéma glissé, jamais silencieux) ;
projection `.generated/game-data/extended/` + manifest scellé (sha256 des
5 pk2, `sourceClient: "isro-live-2026"`, compteurs) ; chargeur Go derrière
le flag ; cross-check compteurs/valeurs contre les JSON SRObro ; **calcul
automatique du cap par la règle de la chaîne complète** (charte §4.5) —
le build échoue si l'intersection ne redonne pas 140.

AC : build reproductible (deux runs = même manifest) ; rapport de
cross-check (écarts expliqués, on ne « corrige » pas les sources) ;
flag off = chargement identique au natif (test) ; le cap dérivé = 140.

### M2 — Progression 140 (charte V2)

Tâches : `LevelCap` injecté (90 natif / 140 étendu) ; courbes XP/or 2026
en étendu ; gel d'XP au cap identique ; franchissement 90→91 impossible
hors flag ; plafonds de maîtrise étendus selon les tables 2026 (sinon
valeur native + écart documenté au manifest).

AC : tests Go des deux réglages ; tables 89-91 et 139-140 des deux sources
dans l'audit ; E2 vérifié en simulation serveur.

### M3 — Contenu : mobs, équipement, skills (charte V3, scope « tout »)

Tâches :

- Items 11D→14D : import, stats dérivées, loot (tables de drop 2026 quand
  présentes, sinon évidence vSRO188 documentée), icônes DDJ 2026 rendues,
  équilibrage = valeurs officielles, aucune invention.
- Mobs 91→140 : stats characterdata 2026 ; placements par zone (§3,
  sources 2+3) ; champions/giants (ratio vSRO documenté) ; aggro, leash,
  respawn natifs du serveur actuel.
- Skills 91→140 (jusqu'à reqMastery 120) : données 2026 + timings natifs
  existants ; les timings manquants sont dérivés des tables et marqués.
- Uniques/bosses des zones étendues : spawnés, avec leurs mécaniques
  natives quand le moteur 1.150 les supporte.

AC : échantillon de stats par degré dans l'audit (comparaison aux JSON
SRObro) ; chaque tranche E3 couverte (compteur spawné par tranche) ; un
mob 100+ combat et droppe un item 11D en flag on, rien n'existe en flag
off ; E3, E4, E5 vérifiés en tests.

### M4 — Monde étendu : toutes les zones de la trajectoire (charte V4, scope « tout »)

Tâches :

- Dériver des données la liste des zones requises (mobs par région +
  leveling path 1→140 sans trou) ; la publier dans le manifest étendu.
- Décodeurs réécrits style id (attributions NOTICE) : `.o2` placements
  (étalon : grammaire SRObro `scripts/parse-v7-o2.cjs` + 59 115 placements
  décodés à recouper), heightmaps, navmesh `.nvm`, minimaps 2026.
- Par zone : terrain, collisions, navmesh, minimap, placements, PNJ
  (incl. téléporteurs), puis activation des spawns de M3 dessus.
- Packager via le pipeline d'assets existant (packs incrémentaux) ; ne
  construire que les zones de la trajectoire (le monde 2026 complet fait
  6 046 régions — hors scope).

AC : liste de zones dérivée citée avec ses compteurs ; chaque zone
traversable en flag on (collision + navmesh + minimap vérifiés par test
automatisé) ; hash des assets servis en flag off = hash natif (E1) ;
téléports aller/retour testés (E7) ; captures dans l'audit.

### M5 — Intégration bout-en-bout

Tâches : test automatisé du parcours complet E8 (tuer→XP→loot→équiper→
niveau suivant, 1→140, y compris changement de zone au bon moment) ;
équilibrage sanitaire (pas de palier impossible : XP requis vs XP par mob
de la tranche) ; nettoyage des écarts M1-M4 restants ; performance E10
(mesures démarrage/mémoire/chargement de zone, comparées à la baseline).

AC : E8 vert en continu ; tableau XP-par-tranche vs mobs-disponibles dans
l'audit ; mesures E10 avec baseline citée.

### M6 — Clôture

Tâches : re-vérification des dix propriétés E1→E10 une à une ; audit final
`M6-audit.md` avec preuves ; capture vidéo ou série de captures du
parcours (une zone par tranche, flag on) ; mise à jour de la charte
(questions restantes, leçons) et de NOTICE.md (attributions).

AC : les dix E cochés avec preuve ; branche `evolution/extended-content`
à jour, poussée sur le fork ; journal coordination à jour.

## 6. Risques connus et replis

| Risque | Repli |
| --- | --- |
| Placements de mobs : le client ne livre pas les tables serveur | Ordre §3 : vSRO188 (évidence) → xSROMap/SRObro KB → minimaps ; chaque zone documente ses sources ; en dernier recours, placement dérivé des minimaps marqué « approximatif » |
| Formats d'assets 2026 différents par endroits (`.o2`, `.nvm`, nouveaux champs) | Étalon SRObro (conversions réussies massives) ; tout décodeur réécrit, jamais importé ; échec au build > silence |
| Volume : 3,2 Go de Data.pk2, 6 046 régions | Ne construire que la trajectoire ; packs incrémentaux ; mesurer E10 à chaque jalon M4+ |
| Équilibre : valeurs 2026 conçues pour des systèmes absents ici (forteresse, jobs modernes) | Valeurs officielles quand jouables ; si un système manque, noter l'écart au manifest — jamais compenser par invention |
| WIP concurrent dans l'arbre principal | Travailler uniquement dans le worktree ; stash ciblé si un fichier partagé bouge (protocole mémoire du 2026-10-08) |
| `../research` absent : fichiers SHA256-pinnés invérifiables | Ne pas les modifier du tout (charte §2.9) |
| Build « exit 0 » ayant échoué (leçon 2026-10-09) | Vérifier les manifestes produits (compteurs, tailles), jamais les seuls exit codes |

## 7. Hygiène et vérification (rappel exécutoire)

- Aucun asset/texte/table retail dans git ; `.generated/` et
  `scripts/lib/generatedRoot.mjs` uniquement ; fixtures minimes dérivées.
- Style id, bannières, taille <1000 lignes/fichier, dprint pour TS/JS
  (`pnpm exec dprint fmt <path>` + `node scripts/checks/check_formatting.mjs --update`
  si fichier du ledger).
- Chaque jalon : `pnpm check source` + gates Go/client selon le périmètre
  touché ; les deux réglages de chaque flag testés.
- Rapports d'audit honnêtes : citer la sortie réelle, jamais « tout vert »
  sans exit 0 dans la session.

## 8. Reprise de mission (pour l'agent qui ouvre ce fichier)

1. `git -C C:/Users/Arnaud/opensro-w-evolution log --oneline -5` et le
   dernier `docs/evolution/M*-audit.md` disent où on en est.
2. Prendre le premier jalon sans audit. Ses tâches et AC sont en §5.
3. En cas de doute de règle : charte d'abord, ce prompt ensuite, question
   owner en dernier recours (jamais d'invention silencieuse).
4. Fin de session = audit + commit + push fork + ligne au journal
   coordination. Ne jamais laisser un jalon « presque fini » sans rapport.
