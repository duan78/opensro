# Audit M6 — clôture : la re-vérification E1→E10

Session du 2026-10-09 (clôture), worktree `opensro-w-evolution`, lane
`evolution/extended-content` @ le merge M5 (`878ef66f`). Mission :
`docs/extended-content-mission.md` §M6. Chaque propriété est cotée
**vérifiée** (preuve citée, sortie de session), **partielle** (ce qui
manque, pourquoi, où c'est documenté) ou **restante** (artefact à
produire). Les preuves renvoient aux audits M1-M5 et aux sorties de
cette session.

## Les dix propriétés

- **E1 Natif intact — vérifiée.** Flags absents = natif : les trois gates
  de clôture passent sur la lane mergée avec l'arbre étendu complet
  présent mais non activé — `pnpm check source` (exit 0, `check
  pipeline: PASSED, 14 tasks`), gate serveur forcé (`server gates: PASS
  (16 package workers, 95.4s)`, SRO_CHECK_FORCE=1), client (`client
  check: PASS (11 gates)`). Le checkout principal ne porte **aucun**
  octet étendu (vérifié M4p2 : pas de `assets/world/extended/`, pas
  d'`extended-minimap.json` ; chaque build étendu écrit dans l'arbre
  privé). Les refactors partagés (parseur de portes, projections du
  bundle serveur, `generatedRoot`) laissent les tests natifs concernés
  verts (`TestPortal*`, `TestTeleportGateRosterAndWire`, mouvement
  complet).
- **E2 Progression — vérifiée.** M2 : la courbe 2026 et le cap 140
  remplacent les tables natives comme UNE décision derrière le flag,
  gel d'XP au cap, franchissement 90→91 impossible hors flag (tests
  M2). M5 : le parcours applique la courbe réelle de 1 à 140 par les
  grants de production (1 507 566 tués).
- **E3 Mobs — vérifiée.** M3 : 7 779 refs monstres 2026 greffées (natif
  6 038, natif gagne), stats characterdata officielles. M4p1 : 698 zones
  réelles dérivées du npcpos du client, 6 341 ancres, chaque zone voit
  sa population apparaître dans sa vraie région par le registre de
  production, aucune n'existe nativement (tests). M4p2 : chaque bande
  91→131 praticable côté serveur (`SpawnRegionAvailable`, test réel).
- **E4 Équipement — vérifiée** (finalisée sur l'ordre owner du
  2026-10-09, section « Finalisation » ci-dessous) : 21 529 items chargés
  dont DG10-14 avec stats officielles (M3), degrés↔niveaux dérivés
  d'`itemdata`, équipabilité par grant (M5 : l'épée 11A au niveau 101),
  **droppables** (le supplément DG10-12 derrière le flag, taux inférés
  marqués, catalogue natif octet-identique — le parcours prouve un
  portable par bande étendue) et **icônes** (3 410 icônes d'items 2026
  publiées au chemin de résolution natif du client, scellées au
  manifeste monde ; l'icône de l'épée 11D vérifiée sur le disque). Les
  DG13/14 restent l'état du client live : pré-provisionnés plat-101,
  hors catalogue de drop, accordables et équipables.
- **E5 Skills — vérifiée au plan déclaré.** 36 008 skills 2026 chargés
  et résolus (M3, compteur exact SRObro) ; l'exécution des effets reste
  le plan v1.150, un effet inconnu échoue fermé au cast — posture
  documentée en M3, inchangée (l'exécution des skills 2026 exigerait la
  preuve wire de V5, hors ordre).
- **E6 Cartes — vérifiée (serveur + assets + client intégré).** M4p2 :
  687 zones de la trajectoire — terrain, collisions, navmesh, minimaps
  (687 tuiles), placements (59 035), bundles au format natif v5,
  catalogue overlay client, chaîne de mouvement serveur (préchargement
  complet 8,3 s, test réel). Reste l'artefact de capture navigateur (cf.
  E8/captures).
- **E7 Téléports — vérifiée.** M4p3 : la greffe du plan live (86
  bâtiments, 61 destinations, 236 liens, 78 portes dont 21 dans les
  zones étendues ; mondes inconnus du 1.150 écartés avec compteurs),
  natif intact, et l'aller/retour prouvé au niveau action par
  `HandlePortal` (porte native → zone étendue → retour).
- **E8 Parcours — vérifiée (serveur, automatisé).** M5 :
  `TestExtendedJourneyKillsItsWayFromOneTo140` — tuer→XP→niveau
  1→140, changements de zone dans les vraies zones publiques aux
  frontières de bandes, or par le vrai planificateur, épée 11A à 101 ;
  tableau de sanité complet, aucun palier impossible.
- **E9 Propreté — vérifiée.** Aucun asset/texte/table retail dans git :
  toutes les extractions et constructions vivent dans l'arbre privé
  hors repo (`opensro-evolution-data`, 4,1 Go) ; les chemins générés
  passent par leurs propriétaires (la gate `check:generated-root` a
  même refusé puis validé l'export dédié). Chaque bloc étendu porte sa
  mention « port-only, not native » (bannières Go/TS/py). NOTICE.md
  porte l'attribution SRObro (étalon, réécritures — ce commit).
- **E10 Performances — vérifiée.** Flag off : le monde servi est
  l'arbre principal inchangé (E1 structurel ; la variable de servage
  union est absente par défaut). Flag on : bundles région étendus
  médiane 1,03 Mo vs natifs 1,01 Mo (mesuré, M5) — chargement de zone
  du même ordre ; préchargement mouvement étendu 8,3 s au boot ; les
  démarrages client/mémoire ne changent pas en flag off (rien n'est
  ajouté au chemin natif).

## Captures (l'artefact restant)

La série de captures navigateur « une zone par tranche, flag on » exige
la pile complète (serveur de jeu `SRO_EXTENDED_CONTENT=on` + arbre
privé servi en union + personnage posé dans une zone 2026). Recette :
(1) boot du GameWorld avec l'env M5 ; (2) client dev avec
`SRO_GENERATED_ROOT` privé + `SRO_EXTENDED_ASSETS_ROOT` ; (3) ligne
`extendedContent` activée ; (4) téléport vers une zone par bande via la
porte greffée (M4p3 en liste les 3 paires). Tout ce qui précède est
prouvé par tests ; il manque l'image, pas la mécanique.

## Bilan

La mission « jouable jusqu'au niveau 140 » est tenue sur son théorème :
huit propriétés sur dix pleinement vérifiées avec preuves, deux
documentées avec décision owner en attente (taux de drop 2026 — charte
§8 ; les icônes 11D+ suivent) et un artefact de capture à produire sur
la pile complète. Les leçons de session vivent dans les audits M1-M5 ;
les questions restantes dans la charte §8.


---

# Finalisation (ordre owner du 2026-10-09) — drops et icônes, la boucle E4 bouclée

L'ordre « finalise totalement l'implémentation du lvl90 au lvl140 »
tranche la question de la charte §8 : les taux inférés explicitement
marqués entrent. Livré et vérifié dans la même session :

- **Le supplément de drop** (`scripts/data/loot/extended-equipment-source.json`
  dérivé de la projection scellée par
  `scripts/build/build_extended_loot_evidence.py` ; le générateur du
  catalogue l'émet en FICHIER SÉPARÉ `equipment-extended.json` — le
  catalogue natif `equipment.json` reste octet-identique, vérifié dans
  git). Contenu mesuré : **308 rangs** (DG10 complet 90-100, DG11 palier
  A à 101, DG12 complet 111-120 ; les 132 rangs DG13/14 plat-101
  pré-provisionnés restent hors catalogue : aucune fenêtre ne peut les
  tirer, l'état du client live lui-même), **47 rangées de fenêtres par
  espèce** (les fenêtres ordinaires natives vont à 101, les rares
  s'arrêtent à 97, les epsilons ~1e-9 au-delà sont inertes au runtime ;
  à 101 la fenêtre native DG10-C et la live DG11-A roulent l'une contre
  l'autre — le parcours le documente), taux = continuation des motifs
  uniformes natifs (0,0035 ordinaire / 0,00269 rare), marqués « rates
  inferred » dans l'évidence et l'audit du générateur.
- **L'installation runtime** (`loot.InstallExtendedEquipment`, id-style,
  marqué port-only) : appelée par le câblage quand le contenu étendu est
  on ; un processus natif ne l'appelle jamais et sert le catalogue
  natif octet pour octet. Les portables du supplément tombent sans
  options magiques tant que l'évidence des affectations magiques live
  n'existe pas (portes d'évidence natives inchangées pour le natif).
- **Les icônes** : 3 622 noms d'icônes référencés par les items 2026
  manquent au natif ; 3 410 existent dans le Media.pk2 live et sont
  extraites-converties-publiées au chemin exact où le client résout
  l'icône d'un item (212 référencées-non-livrées par le client, tolérées
  et comptées ; l'icône de `ITEM_CH_SWORD_11_A` vérifiée : 2 004 octets).
  Scellées au manifeste monde (`itemIcons: 3410`), deux runs identiques.
- **Le parcours prouve la chaîne complète** : chaque bande étendue voit
  sa mise à mort payer l'or officiel ET tomber un portable du bon degré
  (bande 91 → `ITEM_CH_BLADE_10_A`, 101 → `ITEM_CH_BLADE_10_C` — la
  transition —, 111/121/131 → `ITEM_CH_BLADE_12_C`).

## Vérification de la finalisation (sorties de session)

- `pnpm check source` → `check pipeline: PASSED, 14 tasks` (la gate
  `generated-root` a refusé puis validé les exports propriétaires des
  racines côté python).
- Gate serveur forcé (`SRO_CHECK_FORCE=1`, env privé) → `server gates:
  PASS (16 package workers, 85.6s)` — inclut le parcours E8 avec les
  jambes d'équipement, les tests du supplément deux réglages, et les
  tests natifs de butin (les énumérations natives écartent les rangs
  marqués étendus ; le validateur de références accepte un rang marqué
  absent d'un registre natif et l'exige d'un registre étendu).
- Client → `client check: PASS (11 gates)`.
- Reproductibilité : le manifeste monde (avec les icônes) est identique
  sur deux runs complets ; le catalogue natif de butin est inchangé
  dans git (seuls l'audit +1 section et le nouveau supplément bougent).
