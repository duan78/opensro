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

## Le reste (carte actualisée après s2)

Par impact décroissant, chaque grappe une lane d'admission :
1. **Danses de barde** (~50 rangs, 5 familles) : programme
   `srpc/atfe/2lvo/urd/vteg×3/cks/cqer/iqer/lper/slcs` — le système de
   rythme 2026, probablement wire-nouveau (charte V5, décision owner).
2. **Auras de régénération** (WATER_HARMONY 13, BARD_MANATRANS 15,
   MPHEAL 10 : `dura efr alop cgri` / `efr laeh hmwm`) : étendre
   `skillrecovery.go` (formes mana/OT existantes à recouper).
3. **Froid CH** (BINGPAN 20, BINGBYEOK 12 : `zf bf rfe tnat` /
   `ffno wp`) : débuffs de stats — la mécanique anormale existe
   (`skillabnormal.go`), formes à admettre.
4. Le reste dispersé (INVISIBLE/MANADRY/AGGROLOW/traps…) : une à une.

## Tests et gates (sorties de session 2)

- `go test ./internal/game/enterworld/ -run "TestExtendedSkillsExecutionCoverage"`
  → **73,7 %** (détail par bande : 491/640 en 91, …, 372/514 en 131).
- `go test … -run TestSkillProbeOverallParity` → natif 88,9 % / live
  84,9 % (union runtime, prédicat partagé).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 40.8s`.
- Gate serveur forcé → `server gates: PASS (16 workers, 90.5s)`
  (tous les tests natifs de skills inchangés verts — le pin
  sha256 de spawnskillparams.go respecté).
