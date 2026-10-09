# Audit M8 — l'intégration des skills 91-140 : fondations et carte (session 1)

Session du 2026-10-09 (nuit), worktree `opensro-w-evolution`, branche
`evolution/extended-content-m8` (base : la lane `evolution/extended-content`
@ `5cf175a0`). Ordre owner : « intègre la totalité des skills pour que le
jeu soit jouable jusqu'à lvl 140 » (mission §M7, budget des 120 familles).

## Ce que la session a établi

- **La carte complète des refus** (`enterworld/extended_skills_programs_test.go`,
  gardé comme diagnostic) : chaque famille non supportée a son programme
  encodé décodé par le compilateur de production. Trois formes de refus :
  1. **Trois instructions live inconnues du compilateur** : `psog {1|2}`
     (287 occurrences, sur CHAQUE rangée d'attaque/défense 2026),
     `repl {1}` et `srpc {skillId}` (les danses de barde, ids 9932+).
     Mesure par alignement : la rangée live d'attaque ne diffère de son
     ancêtre natif **que par le pilote** (3 cellules) — tout le reste est
     identique.
  2. Les programmes qui compilent mais qu'aucune admission n'épingle :
     soins ciblés/groupe (`laeh`), cures (`truc/lruc/rucr`), DoT
     (`fubn/fubb/sknl/2skl/arud/slup/tta/cm/<type> + vteg`), furtivité
     (`edih`), transferts de mana, etc.
  3. Les rangées d'attaque EU : leur queue est `setv(E2SA-key, N)` SANS
     tag `att` — la clé `E2SA` n'est pas dans la table des paramètres, et
     l'épinglage timed du natif (kind 4 : ces maîtrises sont des attaques
     PERSISTANTES) ne passe pas pour la rangée live. Le refus est PLUS
     PROFOND que le pilote — le contrat `setv`/`E2SA` reste à décoder.

- **Les fondations posées** (nécessaires à toute la suite, prouvées sans
  régression) :
  - `skillprogram.go` : le compilateur admet les trois pilotes live
    (arité 1, inférence enregistrée en commentaire). Le fichier
    hash-épinglé `spawnskillparams.go` est resté INTACT (la première
    tentative l'a modifié, la gate d'évidence l'a refusé — replié dans
    le compilateur non épinglé, conformément à la charte §2.9).
  - Les marches d'admission tolèrent les pilotes : offense
    (`skilloffense.go`), timed principal (`skilltimedeffect.go`), et les
    producteurs periodic/pulse-area/item-effect.
  - La suite de mesure (M7) re-tourne : **58,3 % inchangé** — les
    pilotes étaient nécessaires mais pas suffisants ; les refus réels
    sont les contrats de contenu (setv/E2SA, laeh ciblé, DoT, cures).

- **Aucune régression** : les trois pilotes n'apparaissent dans AUCUNE
  rangée native (mesuré : 287 occurrences sur 8 173 rangées live-only,
  0 sur le natif) — le comportement flag-off est identique par
  construction. Gates : source PASS (14 tâches), la gate d'évidence
  hash-pinnée a VERROUILLÉ la tentative sur le fichier épinglé (leçon
  citée), les tests natifs passent.

## La carte du reste (la session 2 du jalon)

Par impact joueur décroissant, chaque grappe une lane du style
`skillrecovery.go` :
1. **Le contrat setv/E2SA des maîtrises d'attaque EU** (~15 familles ×
   13 rangs passé 90 + leurs ancêtres 1-90) : décoder ce que la clé
   `E2SA` (= 0x45325341) installe chez le natif (les rangs natifs
   kind-4 s'épinglent — trouver PAR QUEL producteur — et répliquer
   l'admission pour la rangée live à pilote près).
2. **Les soins ciblés/groupe** (`laeh [hhwm] vteg`, `atfe ... laeh`) :
   étendre les quatre formes admises de `skillrecovery.go`.
3. **Les cures d'états** (`truc/lruc/rucr` : bard/cleric innocent).
4. **Les DoT** (le plus gros bloc : ~100 rangs warlock) : le programme
   compile — l'admission périodique existe (`compileSkillPeriodicEffect`)
   — décoder l'écart de forme.
5. Furtivité (`edih`/`iqer`/`cks`), transferts mana (`hmwm`), danses
   (`srpc` : le système de rythme 2026 — probablement V5, wire-nouveau),
   et les P2SKILL (innate/event 2026 — à trier).

La mesure de couverture (`extended_skills_coverage_test.go`) est le
critère d'acceptation du jalon : la cible owner est 100 % des castables
joueur 91-140 ; ce jour : 58,3 % (1 749/3 002), inchangé par la session
qui a posé les fondations et cartographié les refus.

## Tests et gates (sorties de session)

- `go test ./internal/game/enterworld/ -run "TestSkillProbeUnsupportedPrograms"`
  → la carte (exemples cités).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 37.3s`
  (après repli du tolerance hors du fichier épinglé ;
  `TestReplacementMetadataNativeExecution` verrouille
  `spawnskillparams.go` par sha256 — respecté).
