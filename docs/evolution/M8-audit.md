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

## Session 13 — le consommateur de la garde localisé : noise.go, installation self

Le consommateur runtime de la garde préventive est trouvé :
`action/noise.go` installe `TimedEffect.Preemptive` sur le LANCEUR
(lane self-effect : « enterworld admits it as
SkillTimedEffect.Preemptive », mask/level lus et appliqués au
propriétaire). Étendre la garde au groupe = brancher l'installation
preemptive dans la marche de sélection de groupe de la lane timed (là
où le HoT de groupe installe ses soins sur chaque membre — le chemin
existe), avec la preuve du vide natif (quelles rangées natives portent
pola sur une aire — attendu : les ancêtres HARMONY seulement ou
aucune). C'est une lane d'implémentation complète (parse + runtime +
tests deux réglages), pas une ouverture de porte — le contexte de la
session ne permet pas de la faire proprement. Reste la lane la mieux
spécifiée du jalon : case preemptive de `skilltimedeffect.go`
(relâcher `targeted || result.Area.Present` pour l'efr kind 3),
installation dans la marche groupe, preuve du vide.

## Session 14 — la preuve du vide HARMONY faite : 21/21 rangées natives = la famille elle-même

Mesuré sur le catalogue natif complet : **21 rangées natives portent
`pola` ET un `efr` — toutes sont WATER_HARMONY** (les paliers A/B/D
1-90, jamais admis par ce port). Le relâchement de la clause self-only
de la case preemptive (pour l'efr kind 3) basculerait donc exactement
les ancêtres jamais-admis de la même famille + les 13 rangs 2026 — le
même profil « complétion de port » que BINGPAN s5 (aucune rangée admise
ne change de genre). La preuve est faite et archivée ici ; il reste
l'implémentation : la relaxation de la case + l'installation de la
garde dans la marche groupe de la lane timed (le chemin du HoT
parti) + les tests deux réglages. C'est une session complète de code
— pas tentée à bout de contexte. Aucun rang basculé.

## Session 15 — correction de cadrage : l'efr kind 3 d'HARMONY est la PORTE DE CAST Efr3, pas une aire de groupe

La relecture croisée avant d'écrire le code referme la lecture
« aire de groupe » de s12-s14 : la marche offense documente déjà l'efr
kind 3 — `skilloffense.go:200` : kind 1 = aire d'action, kind 2 =
aire d'aura persistante, **kind 3 = `CastGate.Efr3Present/Efr3Radius`
— une porte de cast** (une exigence de proximité au moment du cast),
PAS une sélection de victimes. WATER_HARMONY est donc une **garde
préventive self avec taux de récupération et une porte de cast Efr3**
— pas une aura de groupe. Conséquence : la marche timed refuse la
rangée AVANT même d'évaluer pola (sa case efr n'admet que kind 1,
shape 1) ; la lane s16 est d'admettre l'efr kind 3 dans la marche
timed COMME PORTE (pas comme aire), ce qui laisse la case preemptive
self-only INTACTE (la garde reste au lanceur — conforme à son
contrat et à son consommateur noise.go). Plus simple et plus sûr que
la relaxation s14 : aucune clause native ne bouge, une seule case efr
s'élargit d'un kind documenté. La preuve du vide (21/21 = HARMONY)
reste valable pour ce périmètre. Aucun rang basculé.

## Session 16 — WATER_HARMONY s'exécute : l'efr kind 3 admis comme porte, la case cgri ajoutée

Deux edits dans la marche timed (`skilltimedeffect.go`), tous deux
documentés par l'évidence existante :
1. la case efr tolère le **kind 3** — la PORTE DE CAST Efr3 que
   `skilloffense.go:200` documente (jamais une sélection de victimes ;
   le mot passe sans rien sélectionner, les effets self restent au
   lanceur). Inférence enregistrée, preuve du vide : 21/21 rangées
   natives pola+efr = HARMONY elle-même.
2. une case **cgri** ({flat, pct}) remplit `result.Recovery`
   (`SkillRecoveryRates` existant), avec la sémantique exacte de
   `itemEffectRecovery` (595A33..595A93 : écritures percent-sum
   indépendantes aux paramètres de récupération HP/MP).
**Couverture 76,3 % → 76,7 %** (+13 rangs 91-140 ; les 21 ancêtres
natifs 1-90 s'exécutent aussi). La clause self-only de la case
preemptive est INTACTE — la garde reste au lanceur, conforme au
contrat et au consommateur noise.go. Package enterworld complet
vert, gates source + serveur vertes, aucun test natif bougé.

## Session 17 — la « porte MANATRANS » n'existe plus : déjà admise, carte re-coupée

Le sondage (foldé) a vérifié la rangée témoin : `admitted=true`,
`PartyHealPinned=true` — MANATRANS est ADMISE et comptée depuis les
corrections de mesure et de tolérances précédentes ; la « porte
lifetime » du sondage s3 était une trace périmée. **Carte re-coupée
avec le prédicat union courant : 58 familles / 268 rangs joueur
restants** (plus grosses : FIREA_TRAP 12, les quatre danses 40,
MANADRY 10, INVISIBLE 10, WATER_HEAL 9, JIPJUNG 9, AGGROLOW 9,
RECOVERYA_GROUP 8, SWORD_SHIELDPD 8, FORGETA_AGGRO 8, WATER_CANCEL
7…). La traîne continue grappe par grappe sur le même patron ; les
danses (40) restent l'arbitrage owner V5.

## Session 18 — la traîne décodée : chaque famille son programme et son drapeau

Le sondage (foldé) a décodé les dix plus grosses familles restantes —
programmes exacts + drapeaux parsés :
- **WATER_HEAL (9)** : `efr{1,6,300,2,50,7} eshp heal` — la forme
  lowestHealProgram d'origine, à UN mot de sélection : le contrat
  exige 4|5, la rangée authorise **7** (= 4|2|1, parti+personnages+
  lanceur ?). La lane : la sémantique du bit 1 (SelectCaster) sur un
  heal eshp + vide natif. La plus proche d'ouvrir.
- **FIREA_TRAP (12)** : `dura lnks part efr att efr burn hide getv×3`
  — un piège périodique lié ; la lane trap existe
  (`skilltrap.go`, questTrapTag) mais cette forme parti+burn+hide est
  distincte.
- **RECOVERYA_GROUP (8)** : `atfe(?) dura puls efr eshp heal hhwm getv×2`
  — HoT parti avec eshp + terme d'arme (heal-over-time parti au
  plus bas) — extension du contrat HealOverTime existant.
- **JIPJUNG (9)** : `dura re{…}` — un buff mono-mot ; sémantique de
  `re` (0x00000065 ?) à identifier.
- **AGGROLOW (9) / FORGETA (8)** : `efr tntd tdwm [getv]` — baisses
  d'aggro de zone ; la lane dtnt existe (discordwave) mais pas
  l'admission hostile de ce patron.
- **WATER_CANCEL (7)** : `bbuf dura drht tnat` — purge timed (drht ?).
- **SWORD_SHIELDPD (8)** : `dura adps iqer` — débuff temporaire de
  parade (adps).
- **MANADRY (10)** : `pmsc tnat getv×2` — drain de mana ciblé
  (abnormal=true déjà).
- **INVISIBLE (10)** : `dura hide getv reqi sk?` — furtivité timed ;
  la lane concealment existe (`skillconcealment.go`), forme à admettre.
Chaque famille a maintenant sa lane nommée. Aucun rang basculé cette
session (contexte) — les cibles s19+ sont prêtes.

## Session 19 — WATER_HEAL : le mot de sélection admis, la seconde porte nommée

Le vide natif mesuré : **9/9 rangées natives** avec `efr{forme 6,
sélection 7}` sont WATER_HEAL elle-même (paliers D 1-90, jamais
admis). Le mot de sélection 7 (0x04|0x02|0x01 = parti|personnages|
lanceur) est admis dans `lowestHealProgram` avec l'inférence
enregistrée. Mais le sondage suivant a nommé la porte suivante :
`lowestHealProgram` PASSE maintenant — le refus est la porte
d'enveloppe de `parseSkillRecovery` elle-même (`row.TargetRequired`
→ return) : WATER_HEAL est un soin eshp **ciblé** (l'efr kind 1
centré sur la cible primaire), une forme que le contrat recovery
refuse par construction (ses spreads eshp sont non ciblés). La lane
s20 : admettre le soin eshp ciblé-centre (efr kind 1 comme les
aires offense) dans le parse + le dispatch runtime. Aucun rang
basculé cette session (le mot admis est nécessaire mais la porte
d'enveloppe reste).

## Session 20 — WATER_HEAL s'exécute : le soin eshp ciblé admis

Le soin au plus bas ratio CIBLÉ (WATER HEAL : `efr{kind 1, forme 6,
sélection 7} eshp heal`, l'aire centrée sur la cible primaire) est
admis dans `parseSkillRecovery` : sa forme possède son ciblage dans
le mot efr lui-même, il contourne donc l'enveloppe « non ciblé
seulement » du parse — le prédicat `lowestHealProgram` vérifie chaque
mot. Inférence enregistrée ; vide natif : 9/9 = WATER_HEAL ; et la
VÉRIFICATION qui a autorisé l'ouverture : le chemin runtime
`lowestHeal` **centre déjà la sélection sur la primaire**
(`secondaryHealTargets` lit `primaryAt`) — le consommateur existait.
**Couverture 76,7 % → 77,1 %** (+12 rangs 91-140 : les 9 WATER_HEAL
et 3 apparentées ; les ancêtres natifs D exécutent aussi). Package
enterworld complet vert, gates source + serveur vertes, aucun test
natif bougé.

## Session 21 — INVISIBLE et toute la lane concealment : déjà exécutées, comptées maintenant

Cinquième correction de mesure : le sondage a montré
`Concealment.Pinned=true, Hide=true` sur la rangée INVISIBLE — le
parse `skillconcealment.go` l'épinglait DEJA, et le runtime la
dispatche (`action/concealment.go` lit `Concealment.Pinned`/
`.Hide`). C'était le prédicat de mesure qui ne comptait pas la lane.
`Concealment.Pinned` rejoint l'union (`rowRuntimeAdmitted` + liste de
kinds). **Couverture 77,1 % → 78,5 %** (+40 rangs — INVISIBLE, la
furtivité CH/EU et leurs apparentées). Cinquième gap de mesure du
jalon : à chaque fois le moteur exécutait déjà, la mesure ne le
disait pas.

## Session 22 — RECOVERYA_GROUP : le mot d'aire kind 2 admis, la dernière porte nommée

Le mot d'aire de la récupération de groupe est élargi : le prédicat
`partyRecoveryArea` admet l'**efr kind 2** (l'aire persistante que
`skilloffense.go` documente) à côté du kind 1 natif — même forme
centrée lanceur, même sélection parti, inférence enregistrée. Mais la
rangée REFUSE encore : son programme MÈNE avec le mot sans-argument
`atfe` (0x65667461, le drapeau d'état d'aire — groupe arity-0 de la
table) AVANT `dura puls efr eshp heal mwhh getv×2`, et les contrats
HoT parti exigent l'efr en tête. La dernière porte est donc
POSITIONNELLE : la position du drapeau atfe dans
`parseHealOverTime`/`partyHealProgram` — le sondage l'a isolée (le
prédicat d'aire passe maintenant, le parse de tête non). Lane s23.
Aucun rang basculé (le mot admis est nécessaire, la position reste).

## Session 23 — le drapeau atfe admis en tête ; la porte suivante est l'eshp dans la queue HoT

La position du drapeau `atfe` est admise : `parseHealOverTime` saute
un `atfe` sans argument en tête du programme parti (inférence
enregistrée, M8 s23) avant le même contrat. La rangée REFUSE
encore — la porte suivante est isolée au mot près : la queue du HoT
(`healProgramTail` : `[mwhh][mwmh][getv HLRU][getv HLMD]`) ne
contient pas **eshp**, que RECOVERYA_GROUP authorise APRÈS le bloc
heal — un HoT parti « au plus bas ratio » par pulse. L'admettre
exige son consommateur dans le chemin partyOverTime du runtime
(`lowestChainHealTarget` existe) — une vraie extension sémantique,
pas une tolérance. Aucun rang basculé (le drapeau admis est
nécessaire, l'eshp reste).

## Session 24 — l'eshp admis dans la queue HoT AVEC son consommateur runtime ; la rangée refuse encore ailleurs

Les deux moitiés de la lane sont livrées ensemble : `healProgramTail`
admet `eshp` (sans argument, au plus une fois, inférence enregistrée)
ET le chemin `partyOverTime` du runtime sélectionne **le membre au
plus bas ratio par pulse** quand `Aura.Eshp` est posé
(`lowestChainHealTarget` — jamais tout le groupe d'un coup). La
sémantique est donc exécutée, pas juste admise. La rangée REFUSE
encore malgré tout : une porte subsiste en aval des mots — le suspect
le plus probable est l'égalité `dura == EffectDurationMs` (la rangée
porte 300 384 ms de durée, l'enveloppe peut la porter différemment) —
à sonder en tête de session s25 avant FIREA_TRAP. Aucun rang basculé
; les deux edits sont corrects et nécessaires (la runtime-gate ne
peut pas régresser : elle ne s'active QUE si Aura.Eshp && HoT parti,
une forme que seules les rangées 2026 portent).

## Session 25 — l'écart de durée 2026 mesuré et toléré ; la rangée refuse encore, la sonde ligne-à-ligne reste

Mesuré : le client live authorise le mot `dura` **384 ms au-dessus de
la colonne durée d'enveloppe** (programme 300 384 vs envelopne 300
000 sur `RECOVERYA_GROUP_B_05`) — les paliers natifs authorisent
l'égalité exacte (8×300000 vérifiés). La tolérance (≤4096 ms) est
admise UNIQUEMENT sur la forme atfe-en-tête (aucune rangée native ne
peut entrer dans la branche — le vide est structurel), inférence
enregistrée. La rangée REFUSE toujours : une porte subsiste après
dura (les candidats restants : Consumption.Pinned/TimingPinned vus
vrais au sondage s3 mais jamais re-vérifiés depuis les edits, ou une
colonne hors liste) — la sonde s26 est un pas-à-pas ligne à ligne
du parse sur la rangée, pas une hypothèse de plus. Aucun rang
basculé ; golangci a attrapé un delta signé impossible sur uint32
(corrigé en int64 — cité en leçon).

## Session 26 — la porte fautive TROUVÉE : la position de l'efr ; restructuration tentée puis ANNULÉE pour préserver le contrat natif

La sonde ligne-à-ligne a nommé la porte : le programme live est
`atfe, dura, puls, EFR, eshp, heal, mwhh, getv×2` — l'efr arrive
**en quatrième position**, pas en tête ; le s23 ne sautait que atfe,
le prédicat d'aire lisait donc DURA et refusait. La vraie forme est
une **interpénétration** de la tête (les mots dans un autre ordre),
pas un simple drapeau de plus. Une restructuration en marche flexible
a été écrite : elle a fait basculer +8 rangs (78,7 %) mais a **cassé
quatre tests natifs** (TestHealOverTimeRowsAreAdmittedByCompleteProgram,
TestHealOverTimeProgramRefusesAlteredShapes, deux RecoveryDivision) —
elle a été ANNULÉE : le contrat natif prime sur la rangée, toujours.
La lane s27 : la marche flexible doit admettre les deux ordres SANS
changer une seule rangée native (répliquer les quatre tests contre la
nouvelle tête AVANT de rouvrir), ou un prédicat dédié à la forme
atfe-interleavée à côté du contrat natif intact.

## Session 27 — le prédicat dédié écrit À CÔTÉ du contrat natif ; natif 100 % vert, la rangée ne épingle pas encore

Le prédicat `interleavedPartyHoT` (la forme atfe, dura, puls, efr,
eshp, heal + queue ordinaire, tolérance de durée s25) est posé AVANT
le contrat natif dans `parseHealOverTime` — qui reste INTACT mot pour
mot en dessous. Les quatre tests natifs qui avaient cassé la
restructuration s26 sont **tous verts** sur cet arbre (TestHealOverTime×2,
RecoveryDivision×2 + package complet). La rangée n'épingle pas
encore : une porte DANS le prédicat lui-même refuse (les candidats :
l'ordre de mes cases — la queue après heal passe par healProgramTail
depuis l'intérieur du switch, ou partyRecoveryArea kind-2 — à
instrumenter en tête de s28 avec un print par case). La voie sûre est
prouvée : prédicat dédié + contrat natif intact = zéro test natif
cassé, contrairement à la restructuration s26.

## Session 28 — la porte fautive ÉTAIT la re-visite de la queue ; la vraie leçon : le test licencié tournait en SKIP depuis des sessions

Le probe par case (probe jetable, effacé avant commit) a nommé la porte
en dix minutes : chaque rangée RECOVERYA_GROUP étendue est
`atfe, dura(300000), puls(5000), efr[2,1,rayon,1,0,5], eshp, heal(n),
mwhh(105), getv, getv`. Les cinq mots de tête marchent tous jusqu'à
op[5], `healProgramTail(6)` dit **true** — puis la marche RE-VISITE
op[6] (`mwhh`, le terme d'arme du heal) et meurt dans `default`. La
queue appartient au vérificateur de queue, exactement comme le contrat
natif consomme `first+3` sans re-marcher : le `break` étiqueté sur le
cas heal est le correctif entier. Couverture : **+8 rangs, 78,7 %**
(2364/3002).

**Mais** — le test licencié natif
`TestHealOverTimeRowsAreAdmittedByCompleteProgram` a FAILI à ce
précis moment, et pour la bonne raison : il mesurait que le catalogue
v1.150 porte les RECOVERYA_GROUP cap-90 AVEC le même programme
interleavé (atfe en tête, efr kind 2 en position 3), et que le contrat
natif reconstruit les refuse — mon prédicat les admettait donc AUSSI
en natif. Deux découvertes en une :

1. **Les preuves du vide s23/s25/s27 (« les natifs ne mènent jamais un
   HoT par atfe ») étaient fausses** — mesurées sur le seul catalogue
   étendu. Le vide-réel de s22–s25 : l'ORDRE interleavé fait refuser
   les rangées natives à la demande d'aire (partyRecoveryArea voit
   dura, pas efr). Les trois commentaires corrigés dans
   skillrecovery.go : la sûreté vient de l'ordre, jamais de
   l'absence du mot.
2. **Ce test licencié tournait en SKIP silencieux depuis au moins s22**
   — `SRO_GENERATED_ROOT` pointé sur l'arbre evolution-data (dont
   client-public/assets est vide) cache le manifest packs et le test
   saute au lieu d'échouer ; un SKIP imprime quand même « ok ». La
   bonne porte : PAS de SRO_GENERATED_ROOT pour les tests — le
   résolveur worktree suit le .git vers le .generated du checkout
   principal. Toutes les re-vérifications natifs de s28 ont tourné
   AVEC le test licencié réel (1,8 s de vrai travail, 28 rangs
   comptés).

Le règlement : l'ensemble d'admission natif reste figé (aucune preuve
binaire pour l'élargir — décision owner, comme les danses) ; le
prédicat dédié reçoit un PLANCHER DE MAÎTRISE ≥ 91 lu dans les
cellules brutes (colonnes 36/37 — `row.Masteries` est encore zéro à
cet instant du parse, la première version du plancher lisait le champ
et refusait tout, la couverture retombait à 78,5 %) ; le catalogue
natif n'a AUCUNE rangée ≥ 91 (« native past-90 total: 0 rows ») donc
le plancher est une preuve du vide hermétique. Ouvert pour l'owner :
les tiers cap-90 de RECOVERYA_GROUP restent non-admis en natif —
élargir exige la preuve du binaire (58D8F0/5830B0) que le serveur
original acceptait les programmes interleavés.

## Session 29 — FIREA_TRAP était DÉJÀ jouable : le lane trap existe, la mesure ne le comptait pas

Le probe (jetable, effacé) a retourné la surprise en une ligne : chaque
rangée FIREA_TRAP étendue est `pinned=true` par `compileCombatTrap`
(dura 120000, lnks[14,300,1,1], trap, efr kind-3 déclencheur rayon 50,
att, efr kind-1 explosion 70/5/35, hide, getv×3) — le compilateur
existe depuis le natif (« la forme Fire Trap du Wizard »), la chaîne
d'exécution existe (projectilecast → p.trap → acceptCombatTrap,
skillcombtrap.go avec ses tests), le planneur seulement ne nomme jamais
ce lane. Comme les murs (s17) et la dissimulation (s25) : un écart de
MESURE, pas un écart moteur. L'union des deux tests compte désormais
`CombatTrap.Pinned` (bucket « trap »). **+12 rangs, 78,7 → 79,1 %**
(2376/3002). Zéro code de production touché, zéro rangée native
modifiée — les tiers cap-90 du Wizard pinnaient déjà aussi (la parité
native monte d'autant).

Les échantillons de la re-mesure dessinent la carte suivante :
HEALA_CYCLE_B_03/05/07 (la famille du test licencié ! — ses tiers
past-90 refusent quelque part, à sonder en premier), RECOVERYA_QUICK,
BINGBYEOK D/E, JIPJUNG, AGGROLOW, SWORD_SHIELDPD, STEALTHA
HIDING/DETECT/POINT, SAINTA_INNOCENT.

## Session 30 — HEALA_CYCLE_B était une fausse piste : le filtre d'échantillon ne mesurait pas l'union

Le probe par ID rangée a retourné la vérité en une passe : les 7 tiers
HEALA_CYCLE_B étendus (maîtrises 80–116, forme efr-en-tête classique
efr[1,1,300,8,0,5] dura 16000 puls 2000 heal mwhh getv×2) sont TOUS
`pinned=true` par le contrat natif — comptés « recovery » dans les
bandes. Ils n'apparaissaient dans les échantillons « unpinned » que
parce que le filtre D'ÉCHANTILLON (ligne dédiée du test de couverture)
ne regardait que le kind du planneur, sans l'union runtime que les
COMPTEURS mesurent. Le filtre reflète désormais `rowRuntimeAdmitted`
(le même prédicat que les comptes). Zéro compteur changé (79,1 %),
les échantillons maintenant honnêtes dessinent la vraie carte :
MANADRY, WATER_CANCEL, JIPJUNG, SWORD_SHIELDPD, SOULA_STUNLINK,
RAZEA INT/STR/PHYSICAL (certains tiers), STEALTHA_CHANGE,
BATTLAA_GUARD, GUARDA_PHYSICAL, REBIRTHA_SPECIAL — plus les P2SKILL/
INNATE (rangées d'arène/innées) et les danses (owner V5).

## Session 31 — la carte corrigée par les preuves client ; AGGROLOW livré (le lane mirage) ; deux écarts de mesure de plus

Les tooltips du client live-2026 (`textdataname.txt`, UTF-16) ont
corrigé la carte AVANT l'implémentation :

1. **MANADRY est PvP-seulement** — les colonnes disent `EnemyP=1,
   EnemyM=0`, le tooltip dit mot pour mot « This skill is not effective
   on monsters », et la bannière du lane status-cast
   (skillstatuscast.go) le nommait déjà comme la rangée que la porte
   Enemy_M écarte À DESSEIN. L'inférence de l'amorce (« restitution au
   lanceur ») était FAUSSE : le vrai effet est le gel de mana (coupe
   de capacité MP + état de panique, délai de potion +4 s) sur des
   CIBLES JOUEURS. Un runtime hostile-joueur est du PvP — hors du
   périmètre « jouable jusqu'à 140 » (PvE). **Reclassé parked-owner**
   avec les trois preuves (colonnes, tooltip, bannière), comme les
   danses. Aucun code écrit.
2. **FORGETA_AGGRO (barde) était déjà épinglé** — tous les tiers
   jusqu'à A_1_03 (m=138) passent le contrat ciblé ; la carte M7
   comptait 8 rangs fantômes. QUATRIÈME écart de mesure du lane
   (murs, dissimulation, pièges, coupes d'hostilité) : l'union ne
   comptait pas `Threat.Decrease`. Corrigé dans les deux tests.
3. **AGGROLOW (warlock, 9 rangs B_03–B_11, m=92–140) livré** — le
   miroir non-ciblé de Discord Wave : `efr[1,1,300,8,0,16] dtnt mwdt`,
   cast préparé 834 ms, tooltip « the monsters that are around you…
   will reduce their hostility toward the caster ». Le prédicat dédié
   `compileUntargetedThreatDecrease` (skillthreatdecrease.go) porte le
   plancher maîtrise ≥ 91 (les tiers cap-90 Mirage/Phantasma de la
   même famille authors la même forme — le contrat ciblé les refuse,
   le plancher les garde dehors : test natif dédié
   TestUntargetedThreatDecreaseLeavesNativeRowsUnpinned sur TOUT le
   catalogue livré). L'exécuteur `acceptUntargetedThreatDecrease`
   (discordwave.go) suit le pattern trap : beginUntargetedCast +
   drapeau p.mirage, et à la libération coupe l'hostilité des monstres
   autour du LANCEUR envers le lanceur (l'événement haineux à source
   lanceur, les mots dtnt plat + terme mwdt), via le même registre que
   Discord Wave. Test runtime dédié (prepare 834 ms → release → deux
   monstres proches coupés, le lointain intact).

**Couverture : 79,1 → 79,7 %** (2393/3002 ; +9 admission AGGROLOW,
+8 comptage FORGETA_AGGRO). Gates : enterworld+action verts (avec le
test licencié en vrai travail), check source PASSED, gate serveur
forcée PASS (89,0 s).

## Amorce s31 (corrigée par la session) — MANADRY : PARKED-OWNER, PAS un lane

Sondé en fin de nuit 30 (probe effacé) : **les 10 tiers natifs
(A_01–A_08, B_01–B_02, maîtrises 16–86) sont AUSSI non-admis** —
famille frontière-moteur, pas un écart étendu. L'admission devra donc
porter le plancher maîtrise ≥ 91 (motif s28) : les rangées natives ne
doivent pas basculer. Formes mesurées : A_10 (maîtrise 106) =
`pmsc[30000,80,11,50,50,5] tnat[4519] getv(1464421700)
getv(1464422997)`, f68=0 (handler instant), effectDur=0, ciblé ;
B_07 (136) mène par `efr[1,2,80,3,0,8]` (aire, 3 cibles, select 8)
avant le même corps. Les deux getv sont les clés que le trap
compileCombatPtrap connaît déjà. pmsc[0]=30000 porte la durée (30 s)
dans le programme, pas dans l'enveloppe ; tnat porte le montant par
palier. C'est un VRAI lane runtime (drain MP périodique sur la cible
+ restitution au lanceur) : parser + exécuteur + preuves du vide —
budget d'une session entière, pas une queue de nuit.

## Amorces s32 (mesurées en fin de nuit 31, implémentation à faire)

Sondées après la livraison s31 (probe effacé) ; les deux sont de VRAIS
lanes runtime, pas des ouvertures de porte :

1. **WATER_CANCEL** (12 rangées étendues dont ~7-9 past-90 ; 6 natives
   cap-90 refusées pareil → plancher maîtrise requis). Forme :
   `bbuf(count 0) dura[10000] drht[86] tant[3185]`, handler 3, ciblé
   EnemyM+EnemyP, tooltips « Decreases the enemy's dodge ability » (A)
   / « hitting ratios » (B). `drht` = 0x74687264 (mot du métadonnées
   name-attack 7F85A0, PAS un status abnormal). Le hook existe dans le
   moteur : formula.go:382 `ratio = attacker.HitRate /
   defender.EvasionRate`, et monsterabnormal.go:132 expose
   EvasionRate/HitRate comme paramètres. À construire : installation
   temporisée d'un modificateur de ratio SUR UN MONSTRE + expiration +
   intégration aux deux jets. Aucun lane existant ne fait de debuff de
   stat temporisé hostile.
2. **JIPJUNG** : LIVRÉ en s32 (voir la section session 32) — le
   `er` des amorces ci-dessus, admis dans le walk timed skill avec
   plancher ≥ 91, installé sur le paramètre 9 comme le lane item.

## Session 32 — JIPJUNG livré (le buff er) ; WATER_CANCEL réduit à ses lignes timed

Le sondage par ligne a encore corrigé la carte avant d'écrire :

1. **Les lignes instantanées C-G de WATER_CANCEL étaient DÉJÀ épinglées**
   — leurs mots sont des STATUSES abnormal connus (`dpsc`=Decay,
   `dmsc`=Weaken, `rssc`=Impotent, `tisc`=Division, `st`=Stun) et le
   lane status-cast (s5/s7) les admet : `statusCastPinned=true` sur
   chaque ligne C/D/E/F/G et CANCEL2_B/C past-90. Le « WATER_CANCEL 7
   rangs » de la carte M7 ne concernait que les lignes TIMED.
2. **WATER_CANCEL timed (A/B/CANCEL2_A) reste un vrai lane** :
   `bbuf(0) dura[10000] terd[90] tant` (A, esquive) / `drht[96]`
   (B, précision) / les deux (CANCEL2_A). terd/drht ne sont PAS des
   statuses (mots du métadonnées name-attack 7F85A0). Le précédent
   natif du mécanisme existe : callbacks.go:97 — ElectricShock écrit
   le PARAMÈTRE 9 (EvasionRate) en canal percent-product avec
   expiration par slot ; MonsterInstanceStats applique les
   Modifiers du block (ids 9/11) aux deux jets (formula.go:382).
   Mais un debuff SANS bit d'état exigerait des modifiers sans slot
   + expiration hors Update + block gardé vivant sans slot actif
   (finish() le pose à nil) — un changement de cycle de vie du block
   abnormal, cœur du combat. Trop gros pour une queue de session :
   **amorce affinée, lane s33, budget session entière.**
3. **JIPJUNG livré (9 rangs, D_04..D_09 + E_01..E_03)** : le mot `er`
   {flat, percent} — la bannière de timeditemmodifier.go documente la
   règle native « 594AC0 installs hpi/mpi on parameters 3/4, er/hr on
   9/11 » — entre dans le walk timed des SKILLS (le lane item
   l'admettait déjà pour les scrolls/pierres, colonne 8 = 1) avec le
   plancher maîtrise ≥ 91 (les tiers cap-90 natifs authorisent er) ;
   la liste d'épinglage gagne Evasion.Present (première mesure : tout
   passait mais rien n'épinglait) ; l'installation écrit le paramètre
   9 en percent-sum puis flat, exactement comme le lane item, dans
   commitCharacterEffect. Le tooltip EN « parry ratio » est
   l'inversion dodge/parry connue de la localisation — les preuves de
   nommage (SOCKET_STONE_ER, EVATION_SCROLL) et la bannière native
   disent EVASION. Tests : plancher étendu (9 rangs, valeurs par
   palier), vide natif (aucune rangée skill apprise — les rangées
   item exclues explicitement), installation (écriture param-9 flat
   35 via le registry).

**Couverture : 79,7 → 80,0 %** (2402/3002). Gates : enterworld+action
verts, check source PASSED, gate serveur forcée PASS (97,3 s).

## Session 33 (étape 1, moteur) — le debuff de ratio SANS slot livré côté moteur ; la leçon residentSkill

Le lane est livré en deux étapes ; celle-ci est le moteur, autonome et
inerte (les rangées épinglent, l'union ne les compte pas — pas de faux
comptage avant l'exécuteur s34).

1. **Block abnormal** : `Modifier.Until` (0 = possédé par un slot,
   comme avant ; horloge propre sinon), `ApplyRatioDebuff` (facteur
   restant, canal factor-product, source 5 — l'arithmétique
   ElectricShock callbacks.go:97 ; une seconde application RAFRAÎCHIT
   l'entrée, jamais de pile — inférence enregistrée),
   `HasTimedModifiers`, expiration dans `Update` (même Mask==0), et le
   block reste vivant sans slot actif (`finish()` de
   monsterstate_abnormal.go). Test unitaire : vie, refresh, expiration,
   et les écritures natives (Until 0) jamais touchées.
2. **MonsterState.ApplyRatioDebuff** : installation sous verrou,
   trackAbnormal garde le tick vivant.
3. **Prédicat** `compileSkillRatioDebuff` : bbuf + dura(==enveloppe) +
   terd/drht (≤100) + tant, ciblé hostile, **plancher lu sur
   row.Masteries** — l'appel vit à la FIN du parse de rangée (après la
   ligne Masteries), plus propre que les cellules brutes.
   **8 rangées épinglées** : A_10..A_12, B_07..B_10, CANCEL2_A_01
   (m=94..122). Test de plancher + vide natif dédiés.

**La leçon (coûteuse, à retenir)** : `skill_storage.go` projette
SkillRow champ-par-champ dans `residentSkill` — **un nouveau champ non
listé aux trois sites (struct, set, projection de relecture) est
silencieusement jeté au stockage**. Symptôme : le prédicat épingle en
appel direct, le test synthétique épingle, la rangée chargée relit
false — trois heures de sondage (doublons d'ID, index de shards,
cache partagé, ordre de parse) avant l'instrumentation store/readback
qui a nommé la porte. Tout futur champ de lane DOIT toucher les trois
sites de residentSkill.

Couverture inchangée : **80,0 %** (2402/3002) — voulu, l'union compte
RatioDebuff quand l'exécuteur existera (s34 : cast ciblé-monstre par
intent walk-then-cast, tant hostility à l'installation, expiration
déjà câblée). Gates : abnormal/enterworld/action/simulation verts,
check source PASSED, gate serveur forcée PASS (87,5 s).

## Session 34 — l'exécuteur du debuff de ratio : le lane WATER_CANCEL timed COMPLET (+8, 80,3 %)

La seconde étape du lane s33 : le cast. `acceptRatioDebuff`→
`advanceRatioDebuffIntent`→`executeRatioDebuff` (action/ratiodebuff.go)
suivent le pattern du capture (monstercapture.go) — le SEUL lane
d'intent existant dont la cible est un monstre : un nouveau drapeau
`DebuffCast` sur basicAttackIntent + une branche dans le dispatch du
tick (basicattack.go), résolution par characterMonster + Mover,
espacement playerToMonsterCombatSpacing, marche par
approachIntentTarget puis exécution à portée. L'exécution :
admission (admitTarget avec le motion du monstre), coût/commit comme
le capture, **installation par MonsterState.ApplyRatioDebuff** (les
écritures sans slot du moteur s33 : paramètres 9/11, facteur restant,
until = now+dura), **tant** déposé comme événement d'hostilité positif
vers le lanceur, trames cast + vitals. Aucun changement de formule :
les deux jets lisent déjà les Modifiers du block via
MonsterInstanceStats.

Test runtime complet (discordwave_test.go) : cast ciblé refusé sans
HasTarget, MP débité, block installé, **MonsterInstanceStats.Monstre
.HitRate = 4 % de la base** (drht 96), tant = 5831 ajouté, et le tick
abnormal 4A4390 expire l'écriture à l'instant exact et restaure la
stat (le test conduit advanceMonsterAbnormals directement). Deux
leçons de test mineures : l'horloge factice a une base non-nulle (les
bornes absolues sortent immédiatement), et le wire exige HasTarget
explicite.

L'union des deux tests de mesure compte `RatioDebuff.Pinned` (bucket
« debuff »). **Couverture : 80,0 → 80,3 %** (2410/3002, +8 exact).

## Session 36 — SWORD_SHIELDPD livré : la posture bouclier-épée (+13, 80,7 %)

Le mapping a été tranché par INFÉRENCE ENREGISTRÉE (aucune preuve
binaire disponible — pas de chaîne de format dans textdataname, pas de
slot documenté) : **a = hausse d'attaque physique en pourcent, b =
coupe de défense encodée au-delà de cent (coupe = b−100, 23..76 %)**.
La lecture gagnante est la seule dont le trade-off se dégrade
monotoni- quement et garde les deux mots vivants à chaque tier :
l'ordre inverse clampe sa coupe à rien (mots morts aux tiers hauts) et
une paire hr {flat, percent} laisse le flat comme bruit face à une
stat d'attaque de niveau 100. Les alternatives sont nommées dans le
commentaire du code pour le prochain lecteur.

Livré sur le motif JIPJUNG : le cas `tagTimedSpda` dans le walk timed
(plancher ≥ 91 en cellules brutes — 22 rangées natives cap-90
authorisent le même mot), le champ `SkillShieldStance` (l'épinglage le
compte), et l'installation dans commitCharacterEffect — la hausse
d'attaque sur la paire 13/14 en percent-sum, la coupe de défense sur
le paramètre 5 en percent-sum NÉGATIF, exactement le canal que la
saignée native utilise pour sa propre coupe de défense (callbacks.go).
`reqi{4,1}` (le bouclier requis) est appliqué par 58D480 avant
dispatch — zéro code. Première mesure : +9 au lieu de +13 — ma borne
a ≤ 100 rejetait les quatre tiers F (a = 102..110 : une hausse
d'attaque de +110 % est légitime) ; borne élargie à 200.

Tests : plancher (13 rangs, valeurs par palier 76..110/23..76), vide
natif (aucune rangée épingle la posture), installation (les trois
écritures : 13/14 à +90 %, 5 à −45 %). **Couverture : 80,3 → 80,7 %**
(2423/3002, +13 exact).

## Amorce s35 (résolue par la session 36) — le texte original reste pour la traçabilité du raisonnement

Sondé en ouverture de nuit 35 (probe effacé) ; la carte disait
« debuff de parade ciblé, 8 rangs » — **les deux étaient faux** :

- **13 rangs past-90** (D_05 m=92 .. F_05 m=132), 35 au total dans la
  famille.
- **Non ciblé, SELF** : toutes colonnes cibles à zéro, cast PRÉPARÉ
  1300 ms, durée **120 000 ms** (2 min), handler 3.
- Programme : `dura[120000] adps{a, b} reqi{4, 1}`. **Le troisième mot
  n'est PAS « iqer » : c'est reqi** (0x72657169, le word
  d'exigence d'équipement) — {4, 1} = bouclier requis, DÉJÀ appliqué
  par 58D480/noteParameterIndex avant dispatch. Aucun travail requis.
- Tooltip : « Decreases the physical defense power of the shield and
  increases the physical attack power for a period of time » — un
  trade-off self : défense du bouclier ↓, attaque physique ↑.
- **adps = 0x73706461** (constante LE corrigée — la première sonde
  cherchait l'inversée et trouvait zéro), authorisé par **22 rangs
  natives cap-90** → plancher maîtrise ≥ 91 obligatoire, motif s28.
- Les couples mesurés {76,123} {83,133} {90,145} {102,164} {110,176} :
  **b ≈ 1,6·a** exactement — deux pourcents couplés.

**La question de preuve avant d'écrire** : quel mot va à quel
paramètre. Lecture naturelle : a → attaque physique +a % (params
13/14, percent-sum), la défense ↓ par (b−100) % (23..76) — mais le
graphe de paramètres n'exprime pas « la défense DU bouCLIER »
spécifiquement (param 5 = toute la défense physique) et la lecture
inverse (a = défense) contredit le tooltip. Sans preuve binaire du
mapping, la règle du lane s'applique : inférence à enregistrer ou
preuve à chercher dans 594AC0 (l'installateur des blocs de buff) —
pas de supposition sur un trade-off de combat. Chemin de livraison :
cas adps dans le walk timed (plancher), SkillFlatRate-paire, écritures
dans commitCharacterEffect, le reqi existant fait le reste.

## Session 37 — la tolérance DMIR : +28 rangs d'un coup (80,7 → 81,6 %) ; la carte s38 mesurée

Le sondage des six familles nommées a trouvé le gros lot : **RAZEA
(33 past-90)** — des status-casts `dmsc/tisc/dpsc/rssc` (Weaken /
Division / Decay / Impotent — les mots du lane livré en s5/s7) +
`tant` + `getv`. La porte refusante, nommée porte-par-porte : la clé
getv **0x52494d44 = DMIR** — le rider du rebalance 2026 que le
producteur périodique tolère déjà depuis s2. La tolérance dans le lane
status-cast (le même motif, trois lignes) admet **28 rangs past-90**
(60 au total, tiers ≤90 inclus) : RAZEA et les autres familles
porteuses de la même combinaison.

**La preuve du vide, corrigée en la mesurant** : le premier probe
mélangeait les cellules ÉTENDUES avec les rangées natives (les
codenames se recoupent) et voyait « 134 porteurs natifs refusés
ailleurs » ; le test natif dédié a cassé sur cette prémisse — et la
vérité est PLUS FORTE : **aucune rangée native ne porte DMIR** (la
clé n'existe pas dans les données v1.150, exactement ce que le
commentaire s2 disait : « measured on the 2026 rows only, never on a
native row »). TestDMIRToleranceAdmitsNoNativeStatusCast gèle ce
fait : si la donnée v1.150 portait jamais la clé, il casse.

Tests : bornes d'admission status-cast vertes, test DMIR dédié,
enterworld+action complets verts. Gates : check source PASSED, gate
serveur forcée PASS (39,9 s).

### Carte s38 (mesurée par le sondage, probe effacé)

- **BATTLAA_GUARD (6)** : aura de groupe barde — atfe scls onff
  efr{2,1,700,8,0,5} ovl2 odar{8,44} getv×3 cks reqi{6,14} —
  handler 3, non ciblé. Proche du lane acceptPartyBuff (efr kind 2,
  odar) : à sonder contre lui.
- **GUARDA_PHYSICAL (4)** : lien guerrier 30 min — lnks{1,1500,2,1}
  dura[1800000] **lkdr{4,72,0}** (mot inconnu) reqi×3, ciblé allié.
- **REBIRTHA_SPECIAL (2)** : B_02 = résurrection-ish `tmur
  heal{0,100} resu{120,30}` ciblé allié+DeadBody (tmur inconnu, sinon
  la forme du lane resu) ; B_BUFF_02 = grosse aura self (hpi odar
  chrch crmc msch pmdg cks, cast 2834 ms).
- **STEALTHA_CHANGE (2)** : `nbuf bbuf dura hitm efr{1,2,100,3,0,8}
  reqi×2 lnks` ciblé **EnemyP** — forced-target sur joueur, saveur
  PvP : candidat parked-owner.
- **SOULA_STUNLINK (5)** : `lnks{0,1500,2,1} dura[120000] abnb{150}
  st{5000,35,10} reqi{6,10}` ciblé allié — un lien avec STUN sur
  l'attaquant (st = le status Stun !) : à sonder contre le lane link.
- La dispersion (~110) : une à une après.

## Session 38 — le lane aura compté : +49 rangs (81,6 → 83,3 %) ; STEALTHA_CHANGE parked

BATTLAA_GUARD s'est avéré **déjà épinglé et exécutable** : le métadonnées
(noteParameterIndex) pose `Aura.Present` pour tout efr kind 2 et
`BuffModifiers.Odar` dans la même passe, et le dispatch
`Aura.Present && (BuffModifiers.Present() || Aura.Eshp)` lance
acceptPartyBuff — la famille entière (tiers natifs ≤90 compris) était
jouable depuis l'origine, plan kind 0, jamais comptée. **CINQUIÈME
écart de mesure** (murs, dissimulation, pièges, coupes d'hostilité,
auras) : l'union des deux tests compte désormais la condition exacte
du dispatch. **+49 rangs past-90** (81,6 → 83,3 %, 2500/3002) —
presque autant que le DMIR, caché au même endroit : les auras de
groupe du barde et leurs cousines.

**STEALTHA_CHANGE reclassé parked-owner (2 rangs)** : le tooltip dit
mot pour mot « You cannot use the skill against monsters » — moquer
un ennemi JOUEUR pour le forcer à vous cibler (hitm). Même classe de
preuve que MANADRY : un runtime hostile-joueur est du PvP, hors du
périmètre PvE.

### Carte s39 (restante)

- **GUARDA_PHYSICAL (4)** : lien guerrier 30 min, mot `lkdr` inconnu.
- **SOULA_STUNLINK (5)** : lien portant un bloc Stun (`st{5000,35,10}`).
- **REBIRTHA_SPECIAL (2)** : forme resu + mot `tmur` inconnu.
- La dispersion (~110) : une à une.
- Parked-owner : danses (V5), MANADRY (PvP), STEALTHA_CHANGE (PvP).

## Amorces s39 — les trois mots DÉCODÉS par les tooltips ; la vraie carte de la dispersion

Les trois tooltips ont nommé les effets (probe effacé) :

1. **GUARDA_PHYSICAL/MAGIC / lkdr — DÉCODÉ en s40 (probe effacé)** :
   le mot 0 est le MASQUE DE VOIE (4 = physique, 8 = magique — les
   bits 0x04/0x08 d'att, confirmé par les deux lignes), le mot 1 la
   PART DÉTOURNÉE EN POURCENT (33 → 75 % selon le palier, m=30→118),
   le mot 2 toujours 0. `lkdr = 0x6c6b6472` (piège de constante :
   chercher l'inversée ne trouve rien). « You divert part of physical
   damage from one member to yourself » : le guerrier encaisse la
   part des dégâts masqués du protégé, 30 min, hors-portée = fin
   (le lnks distance word). Le côté CAST est gratuit : le lane lien
   (acceptLinkedTargetEffect) installe déjà les paires sans écritures
   (précédent Protect/lkag) ; il manque le MOT (cas lkdr dans le walk
   timed, plancher ≥91 — les tiers natifs m=30..88 authorisent la
   même forme, 9 rangées past-90) et le CROCHET DÉGÂTS SUBIS : quand
   le personnage lié encaisse un coup de la voie masquée, débiter la
   part au guerrier — le miroir défensif de commitLinkedMana
   (linkedmana.go, alimenté côté attaquant). Taille s34 : une session
   fraîche.
2. **SOULA_STUNLINK / abnb{150} + st{5000,35,10}** : « covers one
   member with a mask of horror... enemies who are weak lose
   consciousness » — une GARDE RÉACTIVE sur un allié : les attaquants
   subissent l'horreur, les faibles le STUN ({durée 5000, chance 35,
   niveau 10}). Lane : crochet dans le chemin monstre-attaque-joueur —
   taille s34.
3. **REBIRTHA_SPECIAL_B / tmur{10276} + heal{0,100} + resu{120,30}** :
   « the dead is raised to life with a crippled soul... reduced attack
   power and greatly increased defense... when the state ends, HP and
   MP drastically decrease » — une résurrection à état mutilé : le
   tmur porte l'état (attaque ↓ défense ↑, à l'expiration HP/MP
   chutent — le clamp d'expiration du lane item est le précédent).
   Lane : variante resu + effet temporisé avec chute à l'expiration —
   taille s34. Pas d'implémentation partielle (ressusciter sans
   l'état serait un mensonge).

### La dispersion mesurée (l'histogramme, top-20)

**451 rangées non admises sur 230 familles** — mais ~180 sont
**P2SKILL_INNATE_*** (les innates des monstres d'arène/événement :
ASS/PRO/ENC ×020..140 par paliers) — des rangées de MONSTRES
d'événement, pas des skills joueurs : le dénominateur de la
jouabilité joueur les exclut naturellement (l'arène est du domaine
V5/owner, comme la forteresse). La dispersion JOUEUR réelle est
~270 rangées : les 2-5-rang familles du top-20 (FRENZYA_TOUNT_SPRINT,
GUARDA_MAGIC, POISONA_FIELD, TRANSFORMA_MASK, CONFUSIONA_ILLUSION,
FIRE_SHIELD…) plus la traîne. Parked-owner : danses (V5), MANADRY
(PvP), STEALTHA_CHANGE (PvP) — 52 rangées.

## Amorce s41 complétée — le chemin de frisse cartographié pour le crochet GUARDA

Ouverture de s41 (10/10) : les trois derniers faits mesurés avant la
livraison (lectures, pas de probe) :

1. **Le registre** : `statuseffect.Link` (link.go) porte déjà les mots
   lkdh (ManaHPPercent/ManaPercent/ManaCap) et `ManaLinks(division,
   target, now)` est la requête exacte à mirroir — un
   `RedirectLinks(division, victim, now)` filtrant les liens
   redirect, même structure (byOwner du RECIPIENT, phase 2, non
   stoppé, non expiré).
2. **Le point d'atterrissage** : `monsterStrikePlayer`
   (monsterstrike.go:120) — le débit HP de la victime vit DANS
   `strikePlayerInDoor` (sous le verrou du personnage, ligne 149) ;
   `formula.MagicalDamage != 0` marque la voie magique (précédent :
   la boucle COS ligne 247) ; `monsterStrikeInput.percent` est un
   scalaire de dégâts EXISTANT (ligne 129) — la part victime peut
   chevaucher via (100−share)/100, le débit guerrier se fait dans la
   porte du source (précédent commitLinkedMana ligne 87).
3. **Preuve du vide par construction** : AUCUNE rangée native
   n'épingle lkdr (plancher ≥91) → aucun lien redirect ne peut
   exister en natif → le crochet dans le chemin de frappe est
   inopérant pour chaque frappe native, par construction. Le risque
   du changement est confiné aux rangées étendues.

Reste à écrire (une session fraîche, le chemin de frappe natif exige
du soin) : le cas `tagTimedLinkedRedirect = 0x6c6b6472` dans le walk
timed (Link.Present requis, masque ∈ {4,8}, part 1..100, mot 2 = 0,
plancher ≥91, épinglage `Link.Present && (... || Redirect)`), les
trois champs sur SkillEffectLink ET sur statuseffect.Link (l'install
les recopie), `RedirectLinks`, et le crochet dans
monsterStrikePlayer : diversion avant le commit, débit guerrier dans
sa porte, trames vitals des deux côtés.

## Session 42 — GUARDA livré : le lien de redirection de dégâts (+9, 83,6 %)

Écrit selon l'amorce s41 complète, pièce par pièce :

1. **Parseur** : `tagTimedLinkedRedirect = 0x6c6b6472` dans le walk
   timed — Link.Present requis (lnks lu en premier), masque ∈ {4, 8},
   part 1..100, mot 2 = 0, plancher maîtrise ≥ 91 en cellules brutes ;
   l'épinglage compte `Link.Redirect`. Neuf rangées past-90 épinglées
   (PHYSICAL A_12..A_15, MAGIC A_10..A_14), les tiers natifs m=30..88
   épargnés (tests dédiés). Découverte en passant : la variante
   GUARDA_MAGIC_BLOCK est une AUTRE famille (la traîne).
2. **Les deux structs de lien** : Redirect/RedirectMask/RedirectPercent
   sur SkillEffectLink et statuseffect.Link ; l'install les recopie ;
   `RedirectLinks(division, victime, now)` mirror de ManaLinks (byOwner
   du RECIPIENT, phase 2, non stoppé, non expiré).
3. **Le crochet** (action/linkedredirect.go, intégré à
   monsterStrikePlayer) : `scaleLinkedRedirect` réduit CHAQUE formule
   résolue de la voie masquée (MagicalDamage != 0 = voie magique) de
   sa part AVANT le commit — une frappe qui n'atterrit pas détourne
   rien ; `commitLinkedRedirect` débite le guerrier dans SA porte une
   fois la frappe commise, avec trames vitales privées. **Borne v1
   enregistrée** : la diversion blesse mais ne tue jamais le
   protecteur (débit tenu à 1 PV) — la mort du protecteur appartient
   à la machinerie complète du chemin de frappe, qu'un débit latéral
   ne doit pas singer. Preuve du vide PAR CONSTRUCTION : aucun natif
   n'épingle lkdr → aucun lien redirect n'existe en natif → le crochet
   est inopérant pour chaque frappe native.

Test runtime : le lien installé par le ApplyLink EXPORTÉ du registre
(pas de cast nécessaire), trois formules (physique masquée 1000,
magique non masquée 500, physique bloquée 100) → le guerrier prend
440, la victime 600/500/60 ; la borne 1 PV tenue. Tests natifs :
plancher (9 rangées, masques 4|8, parts 33..75) et vide (aucune
rangée épingle le lien).

**Couverture : 83,3 → 83,6 %** (2509/3002, +9 exact). Gates :
enterworld+action+statuseffect verts, check source PASSED, gate
serveur forcée PASS (le gofmt d'un fichier de test a cassé la première
passe — corrigé).

## Session 43 — SOULA_STUNLINK livré : la garde réactive à stun (+5, 83,7 %)

Les trois inférences tranchées et enregistrées :

1. **Le lien anonyme (groupe 0)** : toléré dans le cas lnks UNIQUEMENT
   sous le plancher ≥ 91 — le contrat natif (groupe ≠ 0) reste intact
   pour chaque rangée native.
2. **abnb{150} = le plafond de niveau de la cible du stun** :
   « enemies who are WEAK » — l'attaquant au niveau ≤ 150 (tout le
   game) roll le stun ; le mot est appliqué comme authorisé (la garde
   sans son plafond refuse — jamais de stun non borné). Capture
   indépendante de l'ordre (abnb précède st dans le programme :
   capture en locale, jointe après le walk — première version
   dépendante de l'ordre corrigée par la mesure).
3. **Le roll réactif** : `st{durée, chance, niveau}` devient le
   Link.StunGuard ; une fois la frappe SUR LE MEMBRE COUVERT commise
   (jamais avant — une conséquence, pas un bouclier), chaque garde
   vivante vérifie le plafond puis roll sur le flux natif
   (effectOutcome, clé 0x10000000) et installe le Stun par la
   transaction de block ordinaire (ApplyStunGuard — publication de la
   copie détachée + tick vivant, corrigé : la première version
   mutait la copie sans la rendre).

Les lignes A et B portent des mots différents (A : 5000/35, B :
3000/50) — le test de plancher les admet comme paires mesurées. Le
hook : rollStunGuards dans monsterStrikePlayer après le commit, à
côté du redirect s42 — vide PAR CONSTRUCTION (aucun natif n'épingle
la garde). Tests : plancher (5 rangées, les deux paires de mots),
vide natif, roll sous plafond (stun installé), skip au-dessus (151).
Fixture : nearbyCharacter pour le warlock (le vrai monstre du fixture
de combat — un template vide crée un monstre à 0 PV, découvert en
diagnostiquant).

**Couverture : 83,6 → 83,7 %** (2514/3002, +5 exact). Gates : les
quatre paquets verts, check source PASSED, gate serveur forcée PASS
(deux corrections en route : un staticcheck S1002 et l'ordre
abnb/st).

## Le reste (carte actualisée après s43)

1. **REBIRTHA_SPECIAL (2)** : résurrection mutilée tmur (l'état avec
   chute HP/MP à l'expiration — le clamp d'expiration du lane item est
   le précédent). Puis la traîne joueur (~245).
2. **La traîne joueur (~270−15 rangées)** : les familles 2-5 de
   l'histogramme, une à une après les trois lanes.
3. **Décision owner de périmètre (nouveau s39)** : ~180 rangées
   P2SKILL_INNATE (monstres d'arène/événement) gonflent le
   dénominateur — la jouabilité joueur les exclut-elle du compte ?
   (L'arène est V5.) Sans décision, elles restent dans le dénominateur
   et le plafond de couverture joueur est ~88 %.
4. **Danses de barde** (40 rangs) : le rythme 2026 — décision owner V5,
   jamais reçue ; sans preuve : pas d'implémentation.
5. **MANADRY (10, parked-owner s31)** : PvP-seul — EnemyP sans EnemyM,
   tooltip « not effective on monsters », bannière status-cast le
   refusant à dessein. Un runtime hostile-joueur = du PvP, hors
   périmètre PvE. Sur décision owner uniquement.
6. **Ouvert pour l'owner (s28)** : élargir l'admission HoT native aux
   tiers cap-90 RECOVERYA_GROUP (interleavé) exige la preuve du
   binaire (58D8F0/5830B0).

## Tests et gates (sorties de session 2)

- `go test ./internal/game/enterworld/ -run "TestExtendedSkillsExecutionCoverage"`
  → **73,7 %** (détail par bande : 491/640 en 91, …, 372/514 en 131).
- `go test … -run TestSkillProbeOverallParity` → natif 88,9 % / live
  84,9 % (union runtime, prédicat partagé).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 40.8s`.
- Gate serveur forcé → `server gates: PASS (16 workers, 90.5s)`
  (tous les tests natifs de skills inchangés verts — le pin
  sha256 de spawnskillparams.go respecté).
