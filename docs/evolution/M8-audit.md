# Audit M8 — l'intégration des skills 91-140 : fondations et carte (sessions 1-2)

Sessions du 2026-10-09 (nuit), worktree `opensro-w-evolution`, branche
`evolution/extended-content-m8` (base : la lane `evolution/extended-content`
@ `5cf175a0`). Ordre owner : « intègre la totalité des skills pour que le
jeu soit jouable jusqu'à lvl 140 ». La mesure d'acceptation est
`enterworld/extended_skills_coverage_test.go` : la part des skills
castables 91-140 que la production admet réellement.

## Session 2 — livré

- **Le contrat setv/E2SA décodé** (la première lane de la carte) : les
  maîtrises d'attaque EU s'épinglent comme PARAMÈTRES PASSIFS — `setv`
  avec la clé `E2SA` = `ParameterTwoHandPower` (la table des clés
  connaissait déjà la clé : c'est la MARCHE PASSIVE qui rejetait le
  pilote live). La marche (`skillpassive_damage.go`
  `encodedPassiveParameters`) tolère maintenant les trois pilotes
  2026 → **+184 rangs** : les 15 familles d'attaque/défense EU
  (TWOHANDP, SHIELDP, DUALP, BOWP, DAGP, HARPE, STAFF, SPIRITP…,
  FRENZYA_DEFENSE, GLORYP…) s'exécutent comme leurs ancêtres natifs
  (kind-4 passif).
- **Le pilote DoT `DMIR` toléré** (`skillperiodic.go`) : les rangs DoT
  2026 portent un troisième `getv` avec la clé `0x52494d44` (« DMIR »,
  l'entrée de scaling du rebalance 2026 — mesuré : rangs 2026
  seulement, jamais natifs). Le moteur v1.150 n'a aucun lecteur pour
  elle ; tolérée comme la marche offense tolère MAAT, inférence
  enregistrée → **+99 rangs** : les quatre familles warlock DoT
  (BURN/POISON/BLEEDING/DISEASE, 99 rangs) passent par le producteur
  périodique exactement comme leurs ancêtres natifs.
- **La mesure corrigée vers l'admission runtime réelle** : la mesure
  M8-s1 ne comptait que les KINDS de plan compilés — mais le runtime
  admet aussi `TimedEffect.Periodic.Pinned` (offense périodique,
  `resolveOffensiveSkill`) et le dispatch recovery (`applySkillRecovery`
  : soins ciblés `Heal.Present && TargetRequired`, cures `Cure`,
  LowestHeal, HoT). La mesure (et la parité) comptent maintenant cette
  union — le prédicat est partagé (`rowRuntimeAdmitted`) entre les deux
  tests. +181 rangs qui étaient DÉJÀ lançables mais sous-comptés.

**Couverture : 58,3 % → 73,7 %** (2 213/3 002 castables 91-140 admis).
Parité joueur : natif 88,9 % vs live **84,9 %** (l'écart natif-live
passe de 12,2 à 4,0 points).

## Session 1 — rappel (fondations)

Les trois instructions live (`psog/repl/srpc`) admises au compilateur
(spawnskillparams.go épinglé resté intact) et tolérées dans les marches
offense/timed/periodic/pulse/item. La carte des 120 familles
(`extended_skills_programs_test.go`). Aucune occurrence native des
pilotes — flag-off identique par construction.

## Session 3 — livré

- **Le décodeur `defp` tolère les pilotes** (`skillpassive_defense.go`) :
  les rangées de défense 2026 (`defp {phys,mag,0} reqi psog`) épinglent
  comme leurs ancêtres natifs → **+37 rangs** (SHIELDP_DEFENSE,
  FRENZYA_DEFENSE, GLORYP et apparentés). Couverture **75,0 %**.
- **La carte des familles re-cartographiée avec la mesure union** : le
  reste réel est **66 familles / 359 rangs** (pas 798 — la carte s2
  comptait au vieux prédicat plan-seul). Plus grosses familles :
  BINGPAN 20, WATER_HARMONY 13, BINGBYEOK 12, FIREA_TRAP 12, les quatre
  danses de barde (40), MANADRY 10, INVISIBLE 10, HWABYEOK 9, JIPJUNG 9.
- **Sondage régén** (foldé) : la forme `efr laeh mwmh` de MANATRANS
  PASSE le prédicat `partyHealProgram` — le refus est à la porte
  lifetime/parse avant la dispatch (lane identifiée, non livrée) ;
  MPHEAL est déjà un HoT admis par l'union.

## Sessions 4-5 — le diagnostic chirurgical, puis la porte ouverte sur preuve du vide

La lane status-cast EXISTS (`compileSkillStatusCast` : « damage-free
hostile status programs ») et le programme de BINGPAN
(`fz bf efr{1,1,50,5,0,24} tnat`) correspond presque exactement à son
contrat. Les sondages (foldés) ont isolé les portes exactes :
- **BINGPAN (20 rangs)** : refuse sur `row.ActionDurationMs == 0` — un
  cast de statut INSTANTANÉ dont la durée vit dans les mots de durée
  des blocs fz/bf eux-mêmes, pas dans l'enveloppe d'action. Desserre la
  porte = faire basculer des rangées natives aujourd'hui refusées
  (changement de comportement natif, interdit sans preuve binaire). La
  lane : recenser quelles rangées natives passeraient la porte
  ouverte, et si AUCUNE n'existe, l'ouvrir avec l'inférence enregistrée.
- **BINGBYEOK/HWABYEOK (21 rangs)** : rangées TIMED (handler 3), pas
  instantanées — leur forme `onff {t,pct} wp {…}` n'est pas un
  programme anormal (`abnormal.Present` faux, `onff` n'est pas une
  source anormale). Lane : une admission timed pour la forme
  onff/wp (débuff temporisé), distincte du status-cast.
- **WATER_HARMONY (13)** : `dura efr{3,1,60,0,0,7} pola cgri` —
  `pola` (0x706f6c61) n'a pas de case dans la marche timed ; l'aura de
  régénération de groupe est une nouvelle forme timed (efr select 7).

**Issue s5** : la preuve du vide a montré que les 38 bascules natives
potentielles sont TOUTES les propres paliers 1-90 de BINGPAN (jamais
admis par ce port, livrés par le client, exécutés par l'original) —
porte ouverte avec l'inférence enregistrée, couverture 75,2 %, aucun
rang admis ne change de genre.

Aucune de ces portes ne s'ouvre gratuitement : chacine exigerait soit
la preuve qu'aucune rangée native ne bascule, soit une nouvelle forme
avec son exécution runtime. C'est le travail de s5, avec les cibles
exactes ci-dessus.

## Session 6 — la seconde porte BINGPAN nommée

Les 11 BINGPAN 91+ restantes passent TOUTES les portes d'enveloppe du
contrat status-cast — le refus est dans la case `efr` de la marche du
programme : ces paliers authorisent `efr{1, 4, rayon, cap, 0, 18}` —
forme d'aire **4** (cône directionnel centré sur la cible primaire) et
**sélection 18**, là où le contrat n'admet que forme 2 (cible) /
forme 1 (lanceur) et sélection 24. La sélection 18 n'est pas non plus
dans l'ensemble hostile admis de la marche offense (10/24/26). Ouvrir
exigerait la sémantique de « sélection 18 » prouvée (que fait
TargetSelection_DispatchByShape pour 18 ?) et la preuve du vide natif
associée — c'est de l'RE dirigé, pas une tolérance. Lane s7, cible
exacte : la case efr de `compileSkillStatusCast`.

## Session 7 — les cônes BINGPAN admis (et une correction honnête)

La « sélection 18 » de s6 était une erreur de lecture de mes propres
sondages : les args étaient affichés en HEXADÉCIMAL — `select = 0x18`
= **24 décimal**, la sélection hostile standard. La seule différence
réelle était la **forme d'aire 4** (cône directionnel) sur les rangées
ciblées. La porte est ouverte pour la forme 4 (le sélecteur partagé
`areaVictims` exécute les formes 1-4 et 6 génériquement via
`directionalVictims`), avec l'inférence enregistrée et la même preuve
du vide (seuls les paliers BINGPAN basculent, jamais admis avant).
**Couverture 75,2 % → 75,6 %** (+11 cônes 91+). Leçon citée :
vérifier la base d'affichage d'un sondage avant d'en dériver une
sémantique.

## Session 8 — les portes timed localisées au switch exact

Sondage (foldé) : BINGBYEOK, HWABYEOK et WATER_HARMONY passent TOUTES
les portes d'enveloppe de la marche timed principale (actif, handler 3,
chaînes, consommation, cast, durée, timing, remplacement, HP/HP%,
colonnes). Le refus est exactement l'absence de case pour
`onff {t, pct}` / `wp {…}` et `pola {…}` / `cgri {…}` dans le switch
d'instructions de la marche — et la doctrine du fichier l'exige :
« reconnaître une instruction seule n'active jamais une route »,
« aucune ne peut être effacée pour fabriquer un programme self-only ».
Chaque case exige son consommateur runtime (l'exécution du débuff
temporisé onff/wp ; l'aura de régénération pola/cgri) AVANT
l'admission. C'est le travail de s9 : deux formes avec exécution, pas
des tolérances. Aucun rang basculé cette session (correctement).

## Session 9 — BINGBYEOK/HWABYEOK re-identifiées : des AURAS, pas des timed

Découverte de portée : `onff {périodeMs, coûtMP}` est la MÉTADONNÉE
d'aura pulsée (`SkillAura.PulseMs/PulseMP`, documentée dans
`skilloffense.go` — « efr kind 2, onff +0x290/+0x284 : persistent
aura »). BINGBYEOK/HWABYEOK ne sont donc pas des débuffs timed mais
des **togglés d'aura persistante** — et leur bloc d'effet est `wp {4
mots}` : la marche offense parse `wp` en… un CastGate (`CastGate.Pw =
true`), en IGNORANT ses quatre mots. Deux conséquences pour la lane :
1. l'aura n'épingle jamais (l'admission `Aura.Present` exige un `efr
   kind 2` que ces rangées n'authorisent pas — aura auto-centrée sans
   efr, rayon par défaut ?) ;
2. les 4 mots de `wp` — les valeurs du débuff — n'ont NI parse NI
   consommateur : leur sémantique est le vrai travail de s10 (RE des
   mots + admission d'aura auto-centrée + exécution).
WATER_HARMONY (pola/cgri) inchangée : aura de régén de groupe.
Aucun rang basculé (correctement) — cette session a corrigé la NATURE
de la cible avant d'y investir.

## Session 10 — la découverte décisive : la lane MUR existe déjà, complète

`skillwall.go` (enterworld) + `skillwall.go` (action) implémentent
DÉJÀ exactement cette forme — la bannière du fichier nomme les
familles : « the Chinese Force walls (Crystal Wall, Fire Wall) ». Le
parseur épinglent `onff{période,MP} + pw{mask, pool, defense, parry}`
(les 4 mots : masque de voies normalisé 587A19, pool d'absorption
593684, défense 40EBE0, parade) ; le runtime caste, pulse le coût MP
(585262), absorbe et retire (5851F7), et `playervictim.go` l'applique.
Les BINGBYEOK/HWABYEOK 2026 sont **les paliers supérieurs des mêmes
familles** que le moteur exécute au cap 90 natif — il ne s'agit PAS
d'une nouvelle forme mais d'un écart d'enveloppe dans les portes de
`parseSkillWall` (colonnes 21-33 à zéro, champs 50/51 = « 255 »,
etc.). La lane s11 est donc probablement une porte unique comme s5/s7 :
sonder les portes du parseur sur une rangée 2026, prouver le vide,
ouvrir. Aucun rang basculé cette session — mais la cible est passée
de « RE de 4 mots + exécution » à « une porte d'enveloppe ».

## Session 11 — les murs 2026 s'exécutaient déjà : la mesure les comptait pas

La sonde (foldée) a clos la lane en une question : les rangées
BINGBYEOK/HWABYEOK 2026 passent TOUTES les portes de `parseSkillWall`
— **`Wall.Pinned = true`**. Elles étaient admises par le runtime
depuis le début (le dispatch de cast lit `skill.Wall.Pinned`
directement, `action/skillwall.go:59`) ; c'est le PRÉDICAT DE MESURE
qui ne comptait pas la lane mur. Corrigé dans le prédicat partagé
(`rowRuntimeAdmitted` + la liste de kinds du test de couverture).
**Couverture 75,6 % → 76,3 %** (+21 rangs — la famille mur
complète : 6/5/3/3/4 par bande). Quatrième correction de mesure du
jalon, même nature que s2 : le moteur exécutait, la mesure ne le
disait pas.

## Session 12 — HARMONY décodée : pola = la garde préventive, cgri = le taux de récupération

Le décodage des tags referme la fausse piste « aura de régén » : `pola`
(0x706f6c61) est `tagTimedPreemptive` — la GARDE PRÉVENTIVE de la
marche timed (« A self-only guard: the protection is the owner's, so
it never rides a target, an area or a link ») ; `cgri`
(0x69726763) est `itemEffectRecovery` — le taux de récupération
{flat, pct} (consommateur : `skillitemeffect.go`). WATER_HARMONY est
donc une **garde préventive de groupe avec taux de récupération**
(`dura efr{kind 3} pola{masque, niveau} cgri{flat, pct}`) — et le
refus est la clause self-only de la case preemptive : la marche refuse
pola sur une rangée à aire. La lane s13 : relâcher la clause pour
l'aire de groupe (efr kind 3) avec la preuve du vide natif + le
consommateur runtime de la garde étendue au groupe. De la vraie
implémentation, une seule grappe (13 rangs). MANATRANS non touché
cette session (contexte). Aucun rang basculé.

## Le reste (carte actualisée après s12)

1. BINGPAN : la porte duration (preuve du vide natif d'abord).
2. BINGBYEOK/HWABYEOK : forme timed onff/wp.
3. WATER_HARMONY : forme timed pola (aura régén).
4. MANATRANS : la porte lifetime du parse recovery.
5. **Danses de barde** (40 rangs) : le rythme 2026 — décision owner V5.
6. INVISIBLE/MANADRY/traps/JIPJUNG et la traîne : une à une.

## Tests et gates (sorties de session 2)

- `go test ./internal/game/enterworld/ -run "TestExtendedSkillsExecutionCoverage"`
  → **73,7 %** (détail par bande : 491/640 en 91, …, 372/514 en 131).
- `go test … -run TestSkillProbeOverallParity` → natif 88,9 % / live
  84,9 % (union runtime, prédicat partagé).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 40.8s`.
- Gate serveur forcé → `server gates: PASS (16 workers, 90.5s)`
  (tous les tests natifs de skills inchangés verts — le pin
  sha256 de spawnskillparams.go respecté).
