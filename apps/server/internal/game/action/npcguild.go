/*
===========================================================================

npcguild.go - the guild manager NPC's guild services

A guild manager holds service 0xF (CGObjNPC_SpawnAndConfigureServices
4C6350), and its talk rows reach the guild through requests that name the
NPC first. v1.188 opens each with CGObjPC_CheckNpcFunctionTargetInRange
and the 0xF test (answer 3 for either); the guild rule follows:

	0x73F0 [u32 npc]   level up   -> 0xB3F0 [1] | [2][code]   (v1.188 0x7113)

The level-up window (CIFGuildLevelUp 5EF9A0) prices the next level; the
master pays its gold, the guild its GP (domain.GuildLevelUpCostAt), and
every member online learns the new level and GP from 0x3B29 subOp 5.

===========================================================================
*/

package action

import (
	"opensro.online/server/internal/domain"
	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/social/guild"
	"opensro.online/server/internal/game/world/simulation"
)

const (
	opGuildLevelUpRequest  uint16 = 0x73f0
	opGuildLevelUpResponse uint16 = 0xb3f0

	// guildNpcRefused is the answer to a request whose NPC is missing,
	// out of range or without the guild service.
	guildNpcRefused uint8 = 0x03
)

/*
================
guildNpcAnswer
================
*/
func guildNpcAnswer(opcode uint16, code uint8) OpResult {
	if code == 0 {
		return OpResult{Frames: []wire.Frame{{Opcode: opcode, Payload: []byte{1}}}}
	}
	return OpResult{Frames: []wire.Frame{{Opcode: opcode, Payload: []byte{2, code}}}}
}

/*
================
guildManagerNpc

CGObjPC_CheckNpcFunctionTargetInRange and the 0xF service test: the
selected NPC, in range, keeps a guild. The caller holds the division lock.
================
*/
func (rt *Runtime) guildManagerNpc(division string, c *enterworld.Character, gid uint32) bool {
	if selected, ok := rt.Selected.Get(division, c.Name); !ok || selected != gid {
		return false
	}
	npc, ok := rt.npcForCurrentViewer(division, c, gid)
	return ok && rt.npcWithinHitRange(division, c, npc) && npc.Services.Has(simulation.NpcServiceGuild)
}

/*
================
guildRefusalCode

The category 0x10 notice of a guild door refusal.
================
*/
func guildRefusalCode(refusal domain.GuildRefusal) uint8 {
	switch refusal {
	case domain.GuildRefusalNotMember:
		return guild.GuildErrNotMember
	case domain.GuildRefusalLeaderRequired, domain.GuildRefusalPermissionDenied:
		return guild.GuildErrPermissionDenied
	case domain.GuildRefusalMaxLevel:
		return guild.GuildErrLevelUpFull
	case domain.GuildRefusalGPDeficit:
		return guild.GuildErrLevelUpGPDeficit
	case domain.GuildRefusalGoldDeficit:
		return guild.GuildErrLevelUpGoldDeficit
	}
	return guildNpcRefused
}

/*
================
HandleGuildLevelUp

0x73F0 from the level-up window. v1.188 5C7330 refuses a player outside a
guild (0x0D) and anyone but the master (0x1E) before its job prices the
level; the door does both under the store lock.
================
*/
func (rt *Runtime) HandleGuildLevelUp(division string, c *enterworld.Character, payload []byte) OpResult {
	r := wire.NewReader(payload)
	gid, err := r.U32()
	if c == nil || err != nil || r.Done() != nil {
		return guildNpcAnswer(opGuildLevelUpResponse, guildNpcRefused)
	}
	store := rt.deps.GuildAuthority()
	if store == nil {
		return guildNpcAnswer(opGuildLevelUpResponse, guildNpcRefused)
	}
	unlock := rt.lockDivision(division)
	defer unlock()
	if !rt.guildManagerNpc(division, c, gid) {
		return guildNpcAnswer(opGuildLevelUpResponse, guildNpcRefused)
	}
	snapshot, refusal := store.LevelUpGuildAs(division, c.ID)
	if refusal.Refused() {
		return guildNpcAnswer(opGuildLevelUpResponse, guildRefusalCode(refusal))
	}
	push := wire.Frame{Opcode: guild.OpGuildUpdatePush, Payload: guild.EncodeGuildLevel3B29(snapshot.Guild.Level, snapshot.Guild.GP)}
	for _, member := range snapshot.Members {
		if member.CharID != c.ID && rt.PushCharacterFrames != nil {
			rt.PushCharacterFrames(division, member.Name, []wire.Frame{push})
		}
	}
	result := guildNpcAnswer(opGuildLevelUpResponse, 0)
	result.Frames = append(result.Frames, push, goldFrame(c))
	return result
}
