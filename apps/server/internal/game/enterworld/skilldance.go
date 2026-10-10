/*
===========================================================================

skilldance.go - the live "Dance with Music" toggle

Extended content (isro-live-2026), port-only, not v1.150-native (M8 s66,
owner order 2026-10-10): the live bard dance tiers author no efr and no
dru - their whole program is the toggle. ycdc (absent from the v1.150
data, arity zero by the walk consistency of every tier), scls and reqc
ride as floor-gated words with no reader in either binary (the DSCR
precedent); onff {period, pulseMP} is the rhythm the shared parser files
as Aura.PulseMs/PulseMP; setv WRITES the caster's parameter slots the
native dances only READ as radius addends (auraRadius: MUER +0x544, DSER
+0x54C) - music area +metres, dance range +metres.

The family's native ancestors (DANCEA_CASTER and friends) admit through
their efr aura; the live tiers pin on their own evidence and run on the
same pulse-owned party-aura engine (585277 retires the toggle when the
pulse MP runs short). The authored area values ride the row carried;
their live fold into combat stats is the recorded boundary.

===========================================================================
*/

package enterworld

/*
================
SkillDance

The toggle's authored facts: the onff rhythm (5000 ms on every tier, the
same word the native dances carry) and the setv area writes.
================
*/
type SkillDance struct {
	Pinned     bool
	RhythmMs   uint32
	MusicArea  uint32
	DanceRange uint32
}

/*
================
pinDanceToggle

Pin the toggle on its own evidence - onff rhythm plus a music-area write,
never targeted, never area-shaped: its lifetime is pulse-owned, not
timed, so the duration-based pin cannot hold it. The executor is the
party-aura dance lane (acceptPartyBuff), whose replacement rules already
know a Bard switching its own dance.
================
*/
func pinDanceToggle(result *SkillTimedEffect, rhythmMs, musicArea, danceRange uint32, targeted bool) {
	if rhythmMs == 0 || musicArea == 0 || targeted || result.Area.Present {
		return
	}
	result.Dance = SkillDance{Pinned: true, RhythmMs: rhythmMs, MusicArea: musicArea, DanceRange: danceRange}
}
