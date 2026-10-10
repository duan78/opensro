/*
===========================================================================

skilldamagereturn.go - dmgr, the damage-return block

The Warlock's Soul Return lane: the type and the one block's parser,
extracted from the timed walk (M8 s66 keeps skilltimedeffect.go under
the size gate on a real boundary - this file is the dmgr owner).

===========================================================================
*/

package enterworld

/*
================
SkillDamageReturn

dmgr {chance, physical %, magical %, range} (+0x204 on a passive, +0x208
on a buff). CSkillManager_ProcessDamageEffects (5A0C2D) runs on the
defender of every hit SkillCombat_CalculateHitOutcome resolves: an
attacker within range (range > distance) takes back, with chance percent,
trunc(physical% of the hit's physical lane) + trunc(magical% of its
magical lane). The defender's own damage is not reduced.
================
*/
type SkillDamageReturn struct {
	Present                          bool
	Chance, Physical, Magical, Range uint32
}

/*
================
parseDamageReturn

One dmgr block: four words, a chance, a range and at least one lane.
================
*/
func parseDamageReturn(op SkillInstruction) (SkillDamageReturn, bool) {
	if op.Count != 4 || op.Arguments[0] == 0 || op.Arguments[3] == 0 || op.Arguments[1] == 0 && op.Arguments[2] == 0 {
		return SkillDamageReturn{}, false
	}
	return SkillDamageReturn{Present: true, Chance: op.Arguments[0], Physical: op.Arguments[1],
		Magical: op.Arguments[2], Range: op.Arguments[3]}, true
}
