# Audit M5 — intégration bout-en-bout : le parcours 1→140 (E8), l'équilibrage, E10

Session du 2026-10-09 (suite), worktree `opensro-w-evolution`, branche
`evolution/extended-content-m5` (base : la lane `evolution/extended-content`
@ `46766382`, M4 mergé). Mission : `docs/extended-content-mission.md` §M5.

## Livré

- **Le parcours scripté E8**
  (`apps/server/internal/game/action/extended_journey_test.go`,
  `TestExtendedJourneyKillsItsWayFromOneTo140`) : un personnage monte de 1
  à 140 **en tuant**, à travers les vrais chemins — `monsterKillReward`
  (la formule native : mastery-gap, level-gap, dégâts, rareté), la courbe
  2026 réelle, le runtime de progression réel (`ExperienceUpdater`,
  cap 140), la maîtrise entraînée en montée (sinon le taux tombe à 10 %).
  Chaque frontière de bande étendue (91/101/111/121/131) déménage le
  personnage dans une vraie zone de sa bande par la règle d'entrée
  publique (`CanEnterRegion`), et la première mise à mort de la bande
  paie son tas d'or par le vrai planificateur de butin
  (`planMonsterKillLoot`, courbe officielle). Au niveau 101, l'épée 11D
  réelle (`ITEM_CH_SWORD_11_A`) est accordée au voyageur. La chasse
  choisit le mob le plus dense à pleine crédit (±10 niveaux, score
  `ExpToGive × monsterLevelGapRewardScale`) parmi les refs placées, le
  catalogue greffé complet en repli (les placements réels ont des trous
  de granularité par niveau — mesuré au niveau 98).
- **Équilibrage sanitaire** : le tableau des tués par bande ci-dessous,
  produit par le parcours lui-même. Aucun palier impossible (borne
  2 000 000 ; pire palier mesuré : 1 214 358 tués).

## Mesures réelles (sorties du parcours)

- Tués par bande : 1→94, 11→1, 21→5, 31→11, 41→118, 51→52, 61→482,
  71→159, 81→199, **91→4 531, 101→16, 111→41, 121→81,
  131→1 501 776**. Total : 1 507 566 tués ; pire niveau : 139
  (1 214 358). Les bandes 101-121 sont rapides parce que leurs uniques
  (ROC, BONEROC…) paient des centaines de millions d'XP à plein crédit ;
  la bande 131 porte la courbe officielle finale : 139→140 exige
  578 982 029 973 906 XP et la chasse la plus dense paie ~2,1e8/tué
  (chaque grant étant lui-même plafonné par la pince dword native du
  delta EXP, `clampExpDelta` = 0x7fffffff — le contrat client).
  Les accélérateurs pratiques du jeu officiel — champions, géants,
  uniques, bonus de groupe, objets d'XP — sont tous des systèmes natifs
  présents ; le parcours mesure la ligne de base à mob normal.
- Or servi dans les vraies zones (logs du test) : bande 91 → 10 200 or
  (TAHOMET), 101 → 11 250 (ROC), 111 → 13 350 (BONEROC), 121/131 →
  14 400 (ARABIA_HARRISON), chaque tas par `planMonsterKillLoot` sur la
  courbe officielle.
- **E10** : bundles région étendus vs natifs (octets réels des arbres) —
  natifs n=2123, médiane ≈ 1,01 Mo, max ≈ 1,44 Mo ; étendus n=687,
  médiane ≈ 1,03 Mo, max ≈ 1,73 Mo (moyenne +6 %) : le chargement d'une
  zone étendue est du même ordre qu'une zone native. Flag off : le
  checkout principal ne porte aucun octet étendu (E1 structurel, vérifié
  en M4p2) ; le préchargement mouvement étendu coûte 8,3 s au boot
  (mesuré M4p2) pour 687 régions. L'arbre privé total : 4,1 Go.

## Écarts M1-M4 restants — décision

- **Tables de drop des mobs/objets 2026 (E4 « droppables »)** : le
  catalogue de butin embarqué (v1.150, évidences curatorées) ne connaît
  ni les mobs 2026 ni les items 11D+ ; **aucune source locale ne porte
  les taux officiels** (client 2026 : non livrés ; CSV vSRO188/SRObro :
  colonnes stats seulement, vérifié ; `monsters_cap120.csv` : stats
  seulement). La règle de la charte s'applique : jamais compenser par
  invention. L'or (courbe officielle 1→140) et les consommables génériques
  tombent ; l'équipement 11D+ s'acquiert par grant et s'équipe (chemins
  prouvés). **Décision owner requise** pour des taux inférés explicitement
  marqués (précédent : les taux consommables natifs « inferred ») —
  question ajoutée à la charte §8.
- Magic options des hauts degrés, noms partiels non-rare 11D+,
  sémantique d'exécution des skills 2026 (fail-closed) : reportés de
  l'audit M3, inchangés, non bloquants pour E1-E10.

## Tests et gates (sorties de session)

- `go test ./internal/game/action/ -run TestExtendedJourney -count=1` →
  **PASS** (3,8 s, 1 507 566 itérations de grant).
- Les gates complets (source, serveur forcé, client) de la session M4p3
  restent la référence du code partagé ; les gates de clôture de M5
  sont cités dans le M6-audit (les mêmes commandes, rejouées après le
  merge).

## Non couvert / suite

- La capture navigateur du parcours (M6 : une zone par tranche, flag on).
- M6 : re-vérification E1→E10, captures, charte + NOTICE.
