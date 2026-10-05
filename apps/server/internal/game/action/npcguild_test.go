/*
===========================================================================

npcguild_test.go - the guild manager's level-up through the store door

===========================================================================
*/

package action

import (
	"bytes"
	"path/filepath"
	"testing"

	"opensro.online/server/internal/game/enterworld"
	"opensro.online/server/internal/game/item/wire"
	"opensro.online/server/internal/game/social/guild"
	"opensro.online/server/internal/game/world/simulation"
)

// guildManagerGid is the fixture guild manager's object id.
const guildManagerGid = 4001

/*
================
guildManagerFixture

A store-backed character who masters a level 1 guild holding 6000 GP,
beside a selected NPC_EU_GUILD.
================
*/
func guildManagerFixture(t *testing.T) *doorRuntime {
	t.Helper()
	d := openDoorRuntime(t, filepath.Join(t.TempDir(), "authority"), testCharacter())
	deps := d.rt.deps.(*enterworld.Deps)
	deps.Guilds = d.authority.Guilds()
	c := d.character
	if _, err := deps.Guilds.CreateGuild(testDivision, enterworld.GuildRecord{Name: "Lanterns", Level: 1, GP: 6000},
		enterworld.GuildMemberRecord{CharID: c.ID, JID: 1, Name: c.Name, Grade: 0, PermMask: 0xffffffff}, c); err != nil {
		t.Fatal(err)
	}
	npc := simulation.NpcDef{ObjectID: guildManagerGid, RefObjID: 7600, Codename: "NPC_EU_GUILD",
		AuthoredSpawn: true, Spawn: simulation.SeedWorldState(c).Spawn}
	npc.Services = simulation.ResolveNpcServices(npc)
	d.rt.NpcRoster = []simulation.NpcDef{npc}
	d.rt.NpcSpawn.Enabled = true
	d.rt.Selected.Set(testDivision, c.Name, guildManagerGid)
	return d
}

/*
================
TestGuildManagerLevelsTheMastersGuild
================
*/
func TestGuildManagerLevelsTheMastersGuild(t *testing.T) {
	d := guildManagerFixture(t)
	c := d.character
	request := wire.NewWriter(4).U32(guildManagerGid).Payload()
	if out := d.rt.HandleGuildLevelUp(testDivision, c, request); !bytes.Equal(out.Frames[0].Payload, []byte{2, guild.GuildErrLevelUpGoldDeficit}) {
		t.Fatalf("a master without gold answered %x", out.Frames[0].Payload)
	}
	gold := int64(3000000)
	if !d.authority.UpdateCharacters([]*enterworld.Character{c}, "fixture-gold", func() bool { c.Gold = &gold; return true }) {
		t.Fatal("fixture gold refused")
	}
	out := d.rt.HandleGuildLevelUp(testDivision, c, request)
	assertOpcodes(t, out.Frames, opGuildLevelUpResponse, guild.OpGuildUpdatePush, wire.OpPointsUpdate)
	if !bytes.Equal(out.Frames[1].Payload, guild.EncodeGuildLevel3B29(2, 600)) || goldOf(c) != 0 {
		t.Fatalf("level up pushed %x, gold %d", out.Frames[1].Payload, goldOf(c))
	}
	d.rt.Selected.Set(testDivision, c.Name, guildManagerGid+1)
	if out := d.rt.HandleGuildLevelUp(testDivision, c, request); !bytes.Equal(out.Frames[0].Payload, []byte{2, guildNpcRefused}) {
		t.Fatalf("an unselected manager answered %x", out.Frames[0].Payload)
	}
}
