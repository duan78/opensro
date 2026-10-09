/*
===========================================================================

buildExtendedGameDataBundle.mjs - the extended (live-2026) projection

Extended content (isro-live-2026), port-only, not v1.150-native. Builds
the sibling of the native server projection from the extraction
scripts/extract_extended_client_data.py produced: the progression tables,
the item degree census (requirement column 33, non-rare tiers) and the
character level census (column 57), then derives the extended level cap
by the complete-chain rule (charter 4.5) and REFUSES to build unless the
chain redelivers the owner-sealed 140: XP curve, gold curve and playable
gear must all reach the cap, with mob content covering every band below
it. The manifest never mentions 1.150; the native projection stays the
authority and is not touched.
===========================================================================
*/
import { createHash } from "node:crypto";
import { mkdir, readFile, readdir, rename, rm, stat, writeFile } from "node:fs/promises";
import path from "node:path";
import { pathToFileURL } from "node:url";

import { withGeneratedAssetsLock } from "../../rebuildLock.mjs";
import { generatedRoot, extendedGameDataRoot, retailTextdataRoot } from "../world/paths.mjs";
import { readTextDataRowsSync } from "../shared/textDataIo.mjs";

const MANIFEST_FORMAT = "sro-extended-game-data";
const SCHEMA_VERSION = 1;
const SOURCE_CLIENT = "isro-live-2026";
const SOURCE_INVENTORY = "inventory.json";
// Owner-sealed 2026-10-09 (charter 4.5, mission M1). When the official
// chain completes beyond 140 in a future client, the owner lifts this
// expectation; the rule itself never changes.
const EXPECTED_DERIVED_CAP = 140;
// The columns this build reads. level:0 and expRequired:1 are pinned by
// the native reader (apps/server/internal/game/enterworld/leveldata.go);
// skillPoint:2, monsterExpBasis:5 and jobExp:6 are its documented table.
const LEVELDATA_COLUMNS = { level: 0, expRequired: 1, skillPoint: 2, monsterExpBasis: 5, jobExp: 6 };
const LEVELGOLD_COLUMNS = { level: 0, gold: 1 };
// itemdata: codename is column 2; the requirement level is column 33
// (non-rare tiers carry (degree-1)*10+1 +4/tier; every _RARE row of degree
// 11+ ships a flat 101). characterdata: codename column 2, level column 57.
const ITEM_CODENAME_COLUMN = 2;
const ITEM_REQLEVEL_COLUMN = 33;
const CHAR_CODENAME_COLUMN = 2;
const CHAR_LEVEL_COLUMN = 57;
const SROBRO_ROOT = "C:/Users/Arnaud/Desktop/SRObro";

export const defaultExtendedSourceRoot = path.join( generatedRoot, "extended", "source" );
export const defaultExtendedGameDataRoot = extendedGameDataRoot;

/*
================
pathExists
================
*/
async function pathExists( target ) {
	try {
		await stat( target );
		return true;
	} catch ( error ) {
		if ( error?.code === "ENOENT" ) return false;
		throw error;
	}
}

/*
================
parseLevelTable

Whole raw columns, first cell the level, every row numeric - a slipped
schema fails the build instead of silently misreading the curve.
================
*/
function parseLevelTable( sourceRoot, fileName, columns ) {
	const rows = readTextDataRowsSync( path.join( sourceRoot, fileName ) ).map( ( cells ) =>
		cells.map( ( cell ) => cell.trim() )
	);
	const levels = [];
	let previous = 0;
	for ( const cells of rows ) {
		const level = Number( cells[columns.level] );
		if ( !Number.isInteger( level ) || level !== previous + 1 ) {
			throw Error( `${fileName}: expected consecutive levels from 1, broke at row ${levels.length + 1}` );
		}
		previous = level;
		levels.push( cells );
	}
	if ( levels.length < 2 ) {
		throw Error( `${fileName}: no level rows` );
	}
	return { rows: levels, maxLevel: previous };
}

/*
================
censusItems

Degrees and requirement levels from the item shards. Only non-rare rows
define a degree's requirement band (rare rows ship a flat 101 - recorded,
never trusted as a requirement).
================
*/
function censusItems( sourceRoot, shardNames ) {
	const degrees = new Map();
	let rows = 0;
	for ( const shard of shardNames ) {
		for ( const cells of readTextDataRowsSync( path.join( sourceRoot, shard ) ) ) {
			rows += 1;
			if ( cells.length <= ITEM_REQLEVEL_COLUMN ) {
				throw Error( `${shard}: ${cells.length} columns, fewer than the requirement column` );
			}
			const codename = cells[ITEM_CODENAME_COLUMN]?.trim() ?? "";
			const match = /^ITEM_(?:CH|EU)_[A-Z]+_(\d{1,2})_[A-Z]/.exec( codename );
			if ( !match ) {
				continue;
			}
			const degree = Number( match[1] );
			const rare = codename.includes( "RAR" );
			const requirement = Number( cells[ITEM_REQLEVEL_COLUMN].trim() );
			const entry = degrees.get( degree ) ??
				{ items: 0, rare: 0, plain: 0, plainReqLevelMin: null, plainReqLevelMax: null };
			entry.items += 1;
			if ( rare ) {
				entry.rare += 1;
			} else {
				entry.plain += 1;
				if ( Number.isInteger( requirement ) && requirement > 0 ) {
					entry.plainReqLevelMin = Math.min( entry.plainReqLevelMin ?? requirement, requirement );
					entry.plainReqLevelMax = Math.max( entry.plainReqLevelMax ?? requirement, requirement );
				}
			}
			degrees.set( degree, entry );
		}
	}
	return { rows, degrees: Object.fromEntries( [ ...degrees.entries() ].sort( ( a, b ) => a[0] - b[0] ) ) };
}

/*
================
censusCharacters

Mob levels per ten-level band from the character shards; bosses above the
cap are counted, never allowed to lift the chain.
================
*/
function censusCharacters( sourceRoot, shardNames, nativeSkip ) {
	const bands = new Map();
	const bandExamples = new Map();
	const mobsById = new Map();
	const kinds = new Map();
	let rows = 0;
	let maxLevel = 0;
	for ( const shard of shardNames ) {
		for ( const cells of readTextDataRowsSync( path.join( sourceRoot, shard ) ) ) {
			rows += 1;
			if ( cells.length <= CHAR_LEVEL_COLUMN ) {
				throw Error( `${shard}: ${cells.length} columns, fewer than the level column` );
			}
			const codename = cells[CHAR_CODENAME_COLUMN]?.trim() ?? "";
			const kind = codename.startsWith( "MOB_" ) ? "mob" : codename.startsWith( "NPC_" ) ? "npc" : "other";
			const level = Number( cells[CHAR_LEVEL_COLUMN].trim() );
			kinds.set( kind, (kinds.get( kind ) ?? 0) + 1 );
			if ( kind !== "mob" || !Number.isInteger( level ) || level <= 0 ) {
				continue;
			}
			maxLevel = Math.max( maxLevel, level );
			if ( mobsById.size < 65536 ) {
				const id = Number( cells[1]?.trim() );
				if ( Number.isInteger( id ) && id > 0 ) {
					mobsById.set( id, { codename, level } );
				}
			}
			const band = Math.floor( (level - 1) / 10 ) * 10 + 1;
			bands.set( band, (bands.get( band ) ?? 0) + 1 );
			const examples = bandExamples.get( band ) ?? [];
			if ( examples.length < 8 ) {
				// Deterministic NEW-content candidates only: not shipped by
				// the native v1.150 table and not a non-combat family (the
				// leading shards carry whole native-known families - EU_THIEF
				// alone fills ~190 rows per band).
				if (
					nativeSkip?.has( codename ) ||
					NON_COMBAT_MOB_PREFIXES.some( ( prefix ) => codename.startsWith( prefix ) )
				) {
					continue;
				}
				bandExamples.set( band, [ ...examples, { codename, level } ] );
			}
		}
	}
	return {
		rows,
		kinds: Object.fromEntries( [ ...kinds.entries() ].sort() ),
		maxMobLevel: maxLevel,
		mobsPerBand: Object.fromEntries( [ ...bands.entries() ].sort( ( a, b ) => a[0] - b[0] ) ),
		bandExamples: Object.fromEntries( [ ...bandExamples.entries() ].sort( ( a, b ) => a[0] - b[0] ) ),
		mobsById: Object.fromEntries( [ ...mobsById.entries() ].sort( ( a, b ) => a[0] - b[0] ) )
	};
}

// Measured 2026-10-09: families the live table carries that are not
// field combat mobs (server notifiers, events, test rows); the seed
// exemplar never picks one.
const NON_COMBAT_MOB_PREFIXES = [ "MOB_BOT_", "MOB_GM_", "MOB_EV_", "MOB_TEST_", "MOB_EVENT_", "MOB_TUTORIAL" ];

/*
================
loadNativeCharacterCodenames

The v1.150 characterdata codenames, so the seed area prefers band
exemplars the native game never shipped. A missing native extraction
leaves the choice to the first candidate.
================
*/
async function loadNativeCharacterCodenames() {
	try {
		const entries = await readdir( retailTextdataRoot );
		const codenames = new Set();
		for ( const entry of entries ) {
			if ( !entry.toLowerCase().startsWith( "characterdata" ) || !entry.toLowerCase().endsWith( ".txt" ) ) {
				continue;
			}
			for ( const cells of readTextDataRowsSync( path.join( retailTextdataRoot, entry ) ) ) {
				if ( cells.length > 2 ) {
					codenames.add( cells[2].trim() );
				}
			}
		}
		return codenames;
	} catch {
		return undefined;
	}
}

/*
================
loadNativeNpcposRegions

The v1.150 npcpos region ids: the regions the native world serves. The
live-2026 npcpos anchors outside this set are the extended zones - real
regions the native game never populated.
================
*/
async function loadNativeNpcposRegions() {
	try {
		const rows = readTextDataRowsSync( path.join( retailTextdataRoot, "npcpos.txt" ) );
		const regions = new Set();
		for ( const cells of rows ) {
			const region = Number( cells[1]?.trim() );
			if ( Number.isInteger( region ) && region > 0 ) {
				regions.add( region );
			}
		}
		return regions;
	} catch {
		return undefined;
	}
}

/*
================
buildExtendedAreas

Real placements, derived from the live-2026 client's own npcpos: every
anchor whose region the native v1.150 npcpos never served and whose mob
is a 91-140 NEW-content field mob becomes one population row, grouped
into one authored area per region (the real zone). The native lab shape
carries over (leash 140, respawn 5s, maxCount 1 per anchor); the zones
are public since the movement lane (M4 part 2) serves the 2026 regions -
this catalog only loads behind the extended flag, so the native world's
access grades never see it. Anchors per region are bounded; the unbounded
zone build belongs to the world lane.
================
*/
const EXTENDED_AREA_MAX_ANCHORS_PER_REGION = 24;
const EXTENDED_AREA_MIN_BAND = 91;
const EXTENDED_AREA_MAX_BAND = 140;

async function buildExtendedAreas( sourceRoot, characters, nativeCodenames, nativeRegions ) {
	const mobsById = characters.mobsById ?? {};
	const rows = readTextDataRowsSync( path.join( sourceRoot, "npcpos.txt" ) );
	const byRegion = new Map();
	let anchors = 0;
	for ( const cells of rows ) {
		if ( cells.length < 5 ) {
			continue;
		}
		const id = Number( cells[0]?.trim() );
		const region = Number( cells[1]?.trim() );
		const x = Number( cells[2]?.trim() );
		const z = Number( cells[4]?.trim() );
		if (
			!Number.isInteger( id ) || !Number.isInteger( region ) || !Number.isFinite( x ) || !Number.isFinite( z )
		) {
			continue;
		}
		// Dungeon regions carry the 0x8000 bit (negative as authored);
		// the outdoor zones are this milestone's scope, the dungeon
		// interiors follow with the indoor world lane.
		if ( (region & 0x8000) !== 0 || region <= 0 || region > 0x7fff ) {
			continue;
		}
		// The authored-area contract bounds region-local coordinates to
		// [0,1920); the live anchors occasionally sit outside (boundary
		// rows), and those cannot compose into a valid area entry.
		if ( x < 0 || x >= 1920 || z < 0 || z >= 1920 ) {
			continue;
		}
		if ( nativeRegions?.has( region ) ) {
			continue;
		}
		const mob = mobsById[id];
		if ( !mob || mob.level < EXTENDED_AREA_MIN_BAND || mob.level > EXTENDED_AREA_MAX_BAND ) {
			continue;
		}
		if (
			nativeCodenames?.has( mob.codename ) ||
			NON_COMBAT_MOB_PREFIXES.some( ( prefix ) => mob.codename.startsWith( prefix ) )
		) {
			continue;
		}
		const zone = byRegion.get( region ) ?? [];
		if ( zone.length >= EXTENDED_AREA_MAX_ANCHORS_PER_REGION ) {
			continue;
		}
		zone.push( { codename: mob.codename, level: mob.level, x, z } );
		byRegion.set( region, zone );
		anchors += 1;
	}
	const areas = [];
	const zones = {};
	for ( const region of [ ...byRegion.keys() ].sort( ( a, b ) => a - b ) ) {
		const population = byRegion.get( region ).map( ( anchor ) => ({
			codename: anchor.codename,
			x: anchor.x,
			y: 0,
			z: anchor.z,
			maxCount: 1,
			respawnDelayMinSec: 5,
			respawnDelayMaxSec: 5,
			respawn: true,
			aggressive: false,
			sightRange: 0,
			leashRadius: 140,
			generateRadius: 0
		}) );
		const first = byRegion.get( region )[0];
		areas.push( {
			slug: `extended-zone-${region.toString( 16 )}`,
			regionId: region,
			access: "public",
			entry: { x: first.x, y: 0, z: first.z, angle: 16384 },
			population
		} );
		zones[`0x${region.toString( 16 )}`] = {
			anchors: population.length,
			bands: [
				...new Set( byRegion.get( region ).map( ( anchor ) => Math.floor( (anchor.level - 1) / 10 ) * 10 + 1 ) )
			].sort( ( a, b ) => a - b )
		};
	}
	if ( areas.length === 0 ) {
		throw Error( "extended areas: no new-zone npcpos anchors between 91 and 140 derived" );
	}
	const bands = {};
	for ( const [regionId, zone] of Object.entries( zones ) ) {
		for ( const band of zone.bands ) {
			bands[band] = (bands[band] ?? 0) + zone.anchors;
		}
	}
	return {
		areasCatalog: { format: "sro-server-world-area-catalog", version: 1, areas },
		zonesDescriptor: {
			format: "sro-extended-zone-list",
			version: 1,
			derivedFrom: "live-2026 npcpos anchors in regions the v1.150 npcpos never served",
			regionCount: areas.length,
			anchorCount: anchors,
			anchorsByBand: bands,
			zones
		}
	};
}

/*
================
deriveCapByCompleteChain

The charter rule: the cap is the highest level whose whole official chain
exists - XP curve, gold curve, playable gear - with mob content covering
every band below it. A chain input below the sealed 140 (or a hole in the
mob bands) fails the build with the chain printed.
================
*/
function deriveCapByCompleteChain( chain ) {
	const cap = Math.min( chain.xpMaxLevel, chain.goldMaxLevel, chain.gearBandTop );
	for ( let band = 91; band <= cap; band += 10 ) {
		if ( !chain.mobBands[String( band )] ) {
			throw Error(
				`extended chain: mob band ${band}-${band + 9} is empty; derived cap cannot exceed ${band - 1}`
			);
		}
	}
	if ( cap !== EXPECTED_DERIVED_CAP ) {
		throw Error(
			`extended chain redelivered cap ${cap}, expected the sealed ${EXPECTED_DERIVED_CAP}: ` +
				`xp=${chain.xpMaxLevel} gold=${chain.goldMaxLevel} gear=${chain.gearBandTop} (degree ${chain.gearDegree})`
		);
	}
	return cap;
}

function digestBytes( buffer ) {
	return createHash( "sha256" ).update( buffer ).digest( "hex" );
}

async function verifySourceSeal( sourceRoot, inventory ) {
	if ( inventory.sourceClient !== SOURCE_CLIENT ) {
		throw Error( `extraction sourceClient ${inventory.sourceClient} is not ${SOURCE_CLIENT}` );
	}
	for ( const [name, record] of Object.entries( inventory.files ) ) {
		const bytes = await readFile( path.join( sourceRoot, name ) );
		if ( bytes.length !== record.bytes || digestBytes( bytes ) !== record.sha256 ) {
			throw Error( `extraction file ${name} drifted from its sealed digest` );
		}
	}
}

function shardNames( sourceRoot, label ) {
	return readTextDataRowsSync( path.join( sourceRoot, label + ".txt" ) ).map( ( cells ) => cells[0].trim() ).filter(
		Boolean
	);
}

async function crossCheckSrobro( sourceRoot, report ) {
	try {
		const textdata = path.join( SROBRO_ROOT, "assets", "pk2_media", "server_dep", "silkroad", "textdata" );
		for ( const table of [ "leveldata.txt", "levelgold.txt" ] ) {
			const ours = digestBytes( await readFile( path.join( sourceRoot, table ) ) );
			const theirs = digestBytes( await readFile( path.join( textdata, table ) ) );
			report.push(
				`${table}: ${ours === theirs ? "identical" : "DIFFERENT"} to SRObro's extraction (${
					ours.slice( 0, 12 )
				} vs ${theirs.slice( 0, 12 )})`
			);
		}
	} catch ( error ) {
		report.push( `SRObro cross-check skipped: ${error.code ?? error.message}` );
	}
}

/*
================
buildExtendedGameDataBundle
================
*/
export async function buildExtendedGameDataBundle( options ) {
	const sourceRoot = path.resolve( options.source ?? defaultExtendedSourceRoot );
	const outputRoot = path.resolve( options.output ?? defaultExtendedGameDataRoot );
	const inventory = JSON.parse( await readFile( path.join( sourceRoot, SOURCE_INVENTORY ), "utf8" ) );
	await verifySourceSeal( sourceRoot, inventory );
	const itemShards = shardNames( sourceRoot, "itemdata" );
	const characterShards = shardNames( sourceRoot, "characterdata" );

	const leveldata = parseLevelTable( sourceRoot, "leveldata.txt", LEVELDATA_COLUMNS );
	const levelgold = parseLevelTable( sourceRoot, "levelgold.txt", LEVELGOLD_COLUMNS );
	// dg.txt is the gold-walk basis the native server reads (v1.150
	// CDropGoldData); the live-2026 file carries the same role to 140.
	const goldcurve = parseLevelTable( sourceRoot, "dg.txt", LEVELGOLD_COLUMNS );
	const items = censusItems( sourceRoot, itemShards );
	const nativeCodenames = await loadNativeCharacterCodenames();
	const characters = censusCharacters( sourceRoot, characterShards, nativeCodenames );
	const nativeRegions = await loadNativeNpcposRegions();
	const extendedWorld = await buildExtendedAreas( sourceRoot, characters, nativeCodenames, nativeRegions );

	// Gear: the highest shipped degree at or under the sealed cap. A
	// rare-only shipment is that band's playable gear (degrees 13+ ship
	// seal rows whose requirement cell stays a flat 101); degrees above
	// 14 are pre-provisioned and excluded by the owner decision.
	const playableDegrees = Object.entries( items.degrees ).map( ( [degree, entry] ) => ({
		degree: Number( degree ),
		entry
	}) );
	const gear = playableDegrees.filter( ( { degree } ) => degree <= 14 ).reduce( (
		best,
		current
	) => (current.degree > best.degree ? current : best) );
	const chain = {
		xpMaxLevel: leveldata.maxLevel,
		goldMaxLevel: levelgold.maxLevel,
		gearDegree: gear.degree,
		gearBandTop: gear.degree * 10,
		mobBands: characters.mobsPerBand
	};
	const derivedCap = deriveCapByCompleteChain( chain );

	const temporaryRoot = `${outputRoot}.tmp-${process.pid}`;
	await rm( temporaryRoot, { recursive: true, force: true } );
	try {
		await mkdir( temporaryRoot, { recursive: true } );
		const write = async ( name, value ) => {
			const bytes = Buffer.from( JSON.stringify( value, null, 2 ) + "\n" );
			await mkdir( path.dirname( path.join( temporaryRoot, name ) ), { recursive: true } );
			await writeFile( path.join( temporaryRoot, name ), bytes );
			return { path: name, bytes: bytes.length, sha256: digestBytes( bytes ) };
		};
		const files = [
			await write( "leveldata.json", {
				sourceFile: "leveldata.txt",
				columns: LEVELDATA_COLUMNS,
				rows: leveldata.rows
			} ),
			await write( "levelgold.json", {
				sourceFile: "levelgold.txt",
				columns: LEVELGOLD_COLUMNS,
				rows: levelgold.rows
			} ),
			await write( "goldcurve.json", {
				sourceFile: "dg.txt",
				columns: LEVELGOLD_COLUMNS,
				rows: goldcurve.rows
			} ),
			await write( "census.json", { items, characters } ),
			await write( "areas/catalog.json", extendedWorld.areasCatalog ),
			await write( "zones.json", extendedWorld.zonesDescriptor )
		];
		// The whole extraction ships inside the projection under textdata/,
		// sealed by the same digests: the extended catalogs (items, mobs,
		// skills, names) then read one verified tree - the characterdata
		// shards carry the reshape back to the legacy column layout, the
		// name table the synthesis from the object/equip&skill shards.
		await mkdir( path.join( temporaryRoot, "textdata" ), { recursive: true } );
		for ( const [name, record] of Object.entries( inventory.files ) ) {
			const bytes = await readFile( path.join( sourceRoot, name ) );
			await writeFile( path.join( temporaryRoot, "textdata", name ), bytes );
			files.push( { path: `textdata/${name}`, bytes: record.bytes, sha256: record.sha256 } );
		}
		const manifest = {
			format: MANIFEST_FORMAT,
			schemaVersion: SCHEMA_VERSION,
			sourceClient: SOURCE_CLIENT,
			archives: Object.fromEntries(
				Object.entries( inventory.archives ).map( (
					[name, record]
				) => [ name, { bytes: record.bytes, sha256: record.sha256 } ] )
			),
			derivedCap,
			chain: {
				xpMaxLevel: chain.xpMaxLevel,
				goldMaxLevel: chain.goldMaxLevel,
				gearDegree: chain.gearDegree,
				mobBandCount: Object.keys( chain.mobBands ).length
			},
			counts: {
				levelRows: leveldata.rows.length,
				goldRows: levelgold.rows.length,
				goldCurveRows: goldcurve.rows.length,
				itemRows: items.rows,
				characterRows: characters.rows,
				extendedZones: extendedWorld.zonesDescriptor.regionCount,
				extendedAnchors: extendedWorld.zonesDescriptor.anchorCount
			},
			contentDigest: digestBytes( Buffer.from( files.map( ( file ) => file.path + file.sha256 ).join( "\n" ) ) ),
			files
		};
		const manifestBytes = Buffer.from( `${JSON.stringify( manifest, null, 2 )}\n` );
		await writeFile( path.join( temporaryRoot, "manifest.json" ), manifestBytes );
		// The world lane (buildExtendedWorldRegionResources.mjs) owns the
		// movement mirror and the sealed world manifest beside this data
		// projection; move them into the swap so the tree replacement below
		// never erases them.
		for ( const carried of [ "movement", "world-manifest.json" ] ) {
			const from = path.join( outputRoot, carried );
			const to = path.join( temporaryRoot, carried );
			if ( !(await pathExists( from )) || (await pathExists( to )) ) continue;
			await rename( from, to );
		}
		await rm( outputRoot, { recursive: true, force: true } );
		await rename( temporaryRoot, outputRoot );
		const report = [];
		await crossCheckSrobro( sourceRoot, report );
		console.log( `Built extended game-data (${SOURCE_CLIENT})` );
		console.log( `  root: ${outputRoot}` );
		console.log(
			`  derivedCap: ${derivedCap} (xp ${chain.xpMaxLevel}, gold ${chain.goldMaxLevel}, gear DG${chain.gearDegree}, bands ${
				Object.keys( chain.mobBands ).length
			})`
		);
		console.log(
			`  counts: ${items.rows} item rows, ${characters.rows} character rows, ${leveldata.rows.length} levels`
		);
		console.log(
			`  zones: ${extendedWorld.zonesDescriptor.regionCount} real regions, ${extendedWorld.zonesDescriptor.anchorCount} npcpos anchors (bands ${
				Object.keys( extendedWorld.zonesDescriptor.anchorsByBand ).join( "," )
			})`
		);
		console.log( `  manifest: ${digestBytes( manifestBytes )}` );
		for ( const line of report ) {
			console.log( `  cross-check: ${line}` );
		}
	} catch ( error ) {
		await rm( temporaryRoot, { recursive: true, force: true } );
		throw error;
	}
}

/*
================
main
================
*/
if ( process.argv[1] && import.meta.url === pathToFileURL( path.resolve( process.argv[1] ) ).href ) {
	const args = process.argv.slice( 2 );
	const option = ( name ) => {
		const index = args.indexOf( name );
		return index >= 0 ? args[index + 1] : undefined;
	};
	await withGeneratedAssetsLock(
		"extended game-data projection",
		() => buildExtendedGameDataBundle( { source: option( "--source" ), output: option( "--output" ) } )
	);
}
