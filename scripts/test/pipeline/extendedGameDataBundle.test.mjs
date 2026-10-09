/*
===========================================================================

extendedGameDataBundle.test.mjs - the extended projection and its chain rule

The bundle derives the cap from synthetic shard fixtures shaped like the
live-2026 textdata (loader file + UTF-16 shards, inventory sealed with
digests). The complete-chain rule must redeliver 140 and must REFUSE a
shorter gold curve, a hole in the mob bands, or a drifted extraction
seal - never silently build a different game.

===========================================================================
*/
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdtemp, readFile, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";

import { buildExtendedGameDataBundle } from "../../build/server/buildExtendedGameDataBundle.mjs";

const ITEM_COLUMNS = 40;
const CHAR_COLUMNS = 60;

function utf16( text ) {
	return Buffer.from( "\ufeff" + text, "utf16le" );
}

function levelRows( maxLevel, columns ) {
	const rows = [];
	for ( let level = 1; level <= maxLevel; level += 1 ) {
		const cells = Array( columns ).fill( "0" );
		cells[0] = String( level );
		cells[1] = String( level * 118 );
		rows.push( cells.join( "\t" ) );
	}
	return rows.join( "\r\n" ) + "\r\n";
}

function itemRow( codename, requirement ) {
	const cells = Array( ITEM_COLUMNS ).fill( "0" );
	cells[2] = codename;
	cells[33] = String( requirement );
	return cells.join( "\t" );
}

function characterRow( codename, level, id = 0 ) {
	const cells = Array( CHAR_COLUMNS ).fill( "0" );
	cells[1] = String( id );
	cells[2] = codename;
	cells[57] = String( level );
	return cells.join( "\t" );
}

/*
================
writeSource

A minimal but honest extraction: the tables, one item shard carrying the
degrees the chain reads, one character shard covering every band up to
140, npcpos anchors for a new zone region, and the sealed inventory.
================
*/
async function writeSource( directory, { goldMax = 140, mobLevels = null } = {} ) {
	const itemShard = [
		itemRow( "ITEM_CH_SWORD_11_A", 101 ),
		itemRow( "ITEM_CH_SWORD_11_A_RARE", 101 ),
		itemRow( "ITEM_CH_SWORD_12_A", 111 ),
		itemRow( "ITEM_CH_SWORD_13_A_RARE", 101 ),
		itemRow( "ITEM_CH_SWORD_14_A_RARE", 101 ),
		itemRow( "ITEM_CH_SWORD_17_A_RARE", 101 )
	].join( "\r\n" ) + "\r\n";
	const mobLevelsList = mobLevels ??
		[ 1, 50, 91, 95, 100, 101, 105, 110, 111, 115, 120, 121, 125, 130, 131, 135, 140 ];
	const characterShard =
		mobLevelsList.map( ( level, index ) => characterRow( `MOB_CH_TEST_${index}`, level, 9000 + index ) ).join(
			"\r\n"
		) + "\r\n" +
		characterRow( "NPC_CH_TEST", 1 ) + "\r\n";
	// Two anchors of the level-95 and level-105 fixture mobs in one new
	// zone region (0x7777): the real-zone derivation has material.
	const npcposRows = [
		[ 9003, 0x7777, 659.7, 0.0, 981.1 ],
		[ 9005, 0x7777, 700.5, 0.0, 1000.2 ]
	].map( ( cells ) => cells.join( "\t" ) ).join( "\r\n" ) + "\r\n";
	const files = {
		"leveldata.txt": utf16( levelRows( 150, 14 ) ),
		"levelgold.txt": utf16( levelRows( goldMax, 3 ) ),
		"dg.txt": utf16( levelRows( goldMax, 3 ) ),
		"npcpos.txt": utf16( npcposRows ),
		"itemdata.txt": utf16( "ItemData_1.txt\r\n" ),
		"ItemData_1.txt": utf16( itemShard ),
		"characterdata.txt": utf16( "CharacterData_1.txt\r\n" ),
		"CharacterData_1.txt": utf16( characterShard )
	};
	for ( const [name, bytes] of Object.entries( files ) ) {
		await writeFile( path.join( directory, name ), bytes );
	}
	const inventory = {
		sourceClient: "isro-live-2026",
		archives: { "Media.pk2": { bytes: 1, mtimeNs: 1, sha256: "0".repeat( 64 ) } },
		files: Object.fromEntries(
			Object.entries( files ).map( (
				[name, bytes]
			) => [ name, { bytes: bytes.length, sha256: createHash( "sha256" ).update( bytes ).digest( "hex" ) } ] )
		),
		shardCounts: { itemdata: 1, characterdata: 1 }
	};
	await writeFile( path.join( directory, "inventory.json" ), JSON.stringify( inventory ) );
	return inventory;
}

test("the bundle derives the sealed 140 and writes the projection", async () => {
	const source = await mkdtemp( path.join( tmpdir(), "extended-source-" ) );
	const output = await mkdtemp( path.join( tmpdir(), "extended-output-" ) );
	await writeSource( source );
	await buildExtendedGameDataBundle( { source, output } );
	const manifest = JSON.parse( await readFile( path.join( output, "manifest.json" ), "utf8" ) );
	assert.equal( manifest.format, "sro-extended-game-data" );
	assert.equal( manifest.sourceClient, "isro-live-2026" );
	assert.equal( manifest.derivedCap, 140 );
	assert.deepEqual(
		{ xp: manifest.chain.xpMaxLevel, gold: manifest.chain.goldMaxLevel, gear: manifest.chain.gearDegree },
		{ xp: 150, gold: 140, gear: 14 }
	);
	assert.equal( manifest.counts.levelRows, 150 );
	assert.equal( manifest.counts.characterRows, 18, "seventeen mob rows plus the NPC row" );
	assert.ok( manifest.counts.extendedZones >= 1, "the real-zone derivation produced at least one area" );
	assert.ok( manifest.counts.extendedAnchors >= 2, "both fixture npcpos anchors became placements" );
	assert.ok(
		!JSON.stringify( manifest ).includes( "1.150" ),
		"the extended manifest never mentions the native version"
	);
	const leveldata = JSON.parse( await readFile( path.join( output, "leveldata.json" ), "utf8" ) );
	assert.equal( leveldata.rows.length, 150 );
	assert.equal( leveldata.rows[139][0], "140" );
});

test("a gold curve that stops early refuses to build", async () => {
	const source = await mkdtemp( path.join( tmpdir(), "extended-source-" ) );
	const output = await mkdtemp( path.join( tmpdir(), "extended-out-" ) );
	await writeSource( source, { goldMax: 130 } );
	await assert.rejects( () => buildExtendedGameDataBundle( { source, output } ), /redelivered cap 130/ );
});

test("a hole in the mob bands refuses to build", async () => {
	const source = await mkdtemp( path.join( tmpdir(), "extended-source-" ) );
	const output = await mkdtemp( path.join( tmpdir(), "extended-out-" ) );
	await writeSource( source, { mobLevels: [ 1, 50, 91, 100, 101, 110, 111, 120, 131, 140 ] } );
	await assert.rejects( () => buildExtendedGameDataBundle( { source, output } ), /mob band 121-130 is empty/ );
});

test("a drifted extraction seal refuses to build", async () => {
	const source = await mkdtemp( path.join( tmpdir(), "extended-source-" ) );
	const output = await mkdtemp( path.join( tmpdir(), "extended-out-" ) );
	const inventory = await writeSource( source );
	inventory.files["leveldata.txt"].sha256 = "f".repeat( 64 );
	await writeFile( path.join( source, "inventory.json" ), JSON.stringify( inventory ) );
	await assert.rejects( () => buildExtendedGameDataBundle( { source, output } ), /drifted from its sealed digest/ );
});
