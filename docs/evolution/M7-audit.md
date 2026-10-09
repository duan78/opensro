# Audit M7 — jouabilité passé 90 : mesure des skills, villes vivantes, boutiques

Session du 2026-10-09 (clôture M7), worktree `opensro-w-evolution`, branche
`evolution/extended-content-m7` (base : la lane `evolution/extended-content`
@ `c7ad4045`). Ordre owner : « comble et règle ce que tu peux pour que le
jeu soit le plus jouable et logique possible passé lvl 90 » (mission §M7).
Tout est derrière le flag, natif-d'abord, marqué port-only.

## Livré

- **La mesure d'exécution des skills 91-140** (la première chiffraison
  honnête du plus grand écart M6) :
  `enterworld/extended_skills_coverage_test.go` classifie par le
  compilateur de PRODUCTION : **1 749/3 002 skills castables 91-140
  s'exécutent (58,3 %)** — offense 1 205, temporisés 308, passifs 113,
  instantanés 80, soins 27, position 4, menace 10 (les ~1 050 rangées
  « chain-stage » sont les étages internes des chaînes, couverts par le
  plan de leur racine — hors dénominateur).
  `enterworld/extended_skills_measurements_test.go` donne la parité :
  **le catalogue live exécute 68,2 % des skills joueur contre 80,4 % au
  natif** — et le sondage par paires (un rang live vs son ancêtre natif :
  8 cellules de différence) montre que les familles non supportées ne le
  sont **pas non plus au natif** (le HEALA_TARGET natif est kind 0
  aussi) : la frontière est celle du MOTEUR, pas une casse étendue. Le
  budget restant : **798 rangs / 120 familles joueur** (DoT 24,
  convocation de froid 12, soins de groupe, résurrections…), chaque
  famille une lane du style `skillrecovery.go` — multi-sessions, ordre
  owner.
- **Les PNJ de service des villes 2026**
  (`simulation/npcworlddata_extended.go` →
  `AppendExtendedNpcWorldRoster`) : le roster live (son npcpos, ses
  characterdata, ses refmappings) greffé natif-d'abord sur le codename ;
  les additions re-numérotées dans une bande dédiée (NPCGIDBase +10000),
  jamais de collision avec les portes (250000) ni le sol (300000).
  Mesuré : **340 PNJ greffés** (290 dans des régions jamais natives —
  les villes Arabie/Jupiter/etc., 50 boutiquiers avec onglets, 58
  directement dans les zones de chasse de la trajectoire).
- **Les boutiques** : les neuf tables commerce extraites du Media.pk2
  live. Découvertes de données : la face marchandises est **fragmentée
  en loaders** (racines de 7-12 lignes listant des shards) —
  `extract_sharded` les concatène en forme monolithique (5 185
  marchandises, 8 124 prix, 4 990 colis) ; et `refpricepolicyofitem`
  live **insère une colonne zéro à l'index 4** (invariant mesuré :
  8 124/8 124 rangées à 14 cellules) — déposée à l'extraction, le
  précédent du reshape characterdata. `action/commerce_extended.go` →
  `MergeExtendedCommerce` fusionne les onglets natif-d'abord.
  Mesuré : **50 onglets fusionnés** ; 13 marchands échantillonnés → 16
  onglets résolus (22 légalement non-admis : arena/silk, devises hors
  contrat natif), **160 offres dont 70 exigent le niveau 91+** — le
  ravitaillement du joueur 91-140 est réel (les magasins live sont
  bandés par niveau : armes/armures/potions par palier).
- Câblage : le roster greffe dans `devdeps` derrière le flag (avant les
  portes), la fusion commerce dans `wiring_gameplay` — un échec de
  fusion est un échec de boot, jamais une boutique à moitié servie.
  Manifeste données : `3cd5dd86…` (le miroir movement a survécu tous
  les rebuilds de la session).

## Non couvert (les murs mesurés)

- Les 798 rangs de skills 91-140 non exécutés : frontière du moteur,
  évidence vSRO 1.188 pour les tranches couvertes, inférence enregistrée
  au-delà (charte §scope) — l'ordre du propriétaire décidera la
  priorité des 120 familles.
- L'alchimie 11D+ : `magicoption.txt` live mesuré vide (4 lignes,
  172 octets) — aucune évidence locale, jamais d'invention.
- La capture navigateur : recette M6-audit inchangée.

## Tests et gates (sorties de session)

- `go test ./internal/game/enterworld/ -run "TestExtendedSkills"` →
  couverture + parité + familles PASS (chiffres cités ci-dessus).
- `go test ./internal/game/world/simulation/ -run TestAppendExtendedNpcWorldRoster`
  → PASS (340 greffés, préfixe natif intact, bande dédiée, 50
  boutiquiers).
- `go test ./internal/game/action/ -run TestExtendedCommerceShopsResolve`
  → PASS (50 onglets, 160 offres, 70 à niveau 91+).
- `pnpm check source` → `check pipeline: PASSED, 14 tasks in 18.2s`
  (le comparateur de tri cassé des mesures a été attrapé par
  golangci-lint et corrigé vers `sort.Slice` — le classement des
  familles était en ordre de carte avant ça, les chiffres d'audit
  citent la version corrigée).
- Gate serveur forcé → `server gates: PASS (16 package workers, 102.6s)`.
- Client → `client check: PASS (11 gates)`.
