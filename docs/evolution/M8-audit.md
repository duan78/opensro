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

## Session 4 — le diagnostic chirurgical des débuffs (aucun rang basculé)

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

Aucune de ces portes ne s'ouvre gratuitement : chacine exigerait soit
la preuve qu'aucune rangée native ne bascule, soit une nouvelle forme
avec son exécution runtime. C'est le travail de s5, avec les cibles
exactes ci-dessus.

## Le reste (carte actualisée après s4)

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
