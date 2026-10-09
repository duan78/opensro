/*
===========================================================================

buildExtendedWorldRegionResources.mjs - the live-2026 outdoor world lane

Extended content (isro-live-2026), port-only, not v1.150-native. Builds the
extended outdoor world the M4 zone list seals: one streamable region bundle
per real 2026 region, from the private extraction
(<generated>/extended/game/extracted, written by
scripts/extract_extended_world_data.py) into the PRIVATE client-public
tree under the /assets/world/extended namespace. The native lane never
reads or writes any of it, and no extended file lands in the main
checkout's tree.

What the lane owns, in order:

  - sector discovery from the sealed zones.json (never a hand-written
    list), verified against the extraction's complete MAPM/MAPT/MAPO2/NVM
    plane exactly like the native discovery contract;
  - the shared extended object store (BSR/BMT/BMS mesh files and object
    textures) under the extended namespace - the client resolves meshes
    through the extended object index, so an extended mesh never relies on
    a native tree that does not carry it;
  - one region bundle per sector (same builder, same format v5 as native),
    referencing the NATIVE shared sky/water render resources by public
    path: the live world's sky is the same world, and the file is served
    from the main tree that already publishes it;
  - the client catalog overlay (/assets/world/extended/world-region-catalog.json,
    entries with area "outdoor" and source "extended-outdoor-live-2026") the
    extended client merges over the native catalog;
  - the outdoor minimap tiles (2026 Media.pk2 art, converted in the image
    staging) under the native tile path plus the extended art catalog
    (/assets/data/extended-minimap.json) the extended client overlays onto
    the minimap renderer's set;
  - the server movement mirror under the extended game-data projection's
    movement/ root (the same projections the native server bundle builds,
    reused from buildServerGameDataBundle.mjs), so the movement authority
    chain can resolve surfaces for the 2026 regions;
  - a sealed world manifest with content digests: two clean runs produce
    the same manifest, the audit's reproducibility proof.

Run under SRO_GENERATED_ROOT (the private tree) and SRO_GAME_ROOT pointing
at <generated>/extended/game so every shared world module resolves the
2026 extraction through the native roots contract; SRO_EXTENDED_GAME_DATA_ROOT
names the projection that receives the movement mirror and the manifest.

===========================================================================
*/
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { buildJobs } from "../shared/buildParallelism.mjs";
import { mapWithConcurrency } from "../shared/asyncUtils.mjs";
import { sha256Hex } from "../shared/hash.mjs";
import { copyIntoPublicTree, writeIntoPublicTree } from "../shared/publicWrite.mjs";
import { REGION_SIZE, OUTDOOR_WORLD_SHARED_RENDER_PUBLIC_PATH } from "./constants.mjs";
import { parseJmxMapObjectPlacementO2, readJmxMapObjectInfo, readJmxMapTileCatalog } from "./jmx/index.mjs";
import { buildJmxWorldRegionBundle } from "./maploader/buildMapLoaderRegionBundle.mjs";
import { buildTitleSectorObjectResources } from "./objects/buildTitleSectorObjectResources.mjs";
import { MAIN_CHECKOUT_CLIENT_PUBLIC_ROOT } from "../../lib/generatedRoot.mjs";
import {
	extractedRoot,
	gameRoot,
	imageSourceRoot,
	normalizeRegionId,
	publicRoot,
	resolveExtendedGameDataRoot,
	toHex16
} from "./paths.mjs";
import {
	buildObjectNavProjection,
	projectRegionBundle,
	worldAuthorityPath
} from "../server/buildServerGameDataBundle.mjs";
import { exists } from "./io.mjs";

export const EXTENDED_WORLD_AREA = "outdoor";
export const EXTENDED_WORLD_SOURCE_NAME = "extended-outdoor-live-2026";
export const EXTENDED_WORLD_INDEX_PUBLIC_PATH = "/assets/world/extended/world-regions.json";
export const EXTENDED_WORLD_BUNDLE_PATH_TEMPLATE = "/assets/world/extended/regions/region-{id}.json";
export const EXTENDED_WORLD_OBJECT_INDEX_PUBLIC_PATH = "/assets/world/extended/object-resources.json";
export const EXTENDED_WORLD_OBJECT_MESH_ROOT_PUBLIC_PATH = "/assets/world/extended/object-meshes";
export const EXTENDED_WORLD_CATALOG_PUBLIC_PATH = "/assets/world/extended/world-region-catalog.json";
export const EXTENDED_MINIMAP_CATALOG_PUBLIC_PATH = "/assets/data/extended-minimap.json";
const EXTENDED_WORLD_MANIFEST_NAME = "world-manifest.json";

/*
================
extendedRegionBundlePublicPath
================
*/
export function extendedRegionBundlePublicPath( id ) {
	const hex = normalizeRegionId( id ).slice( 2 );
	return EXTENDED_WORLD_BUNDLE_PATH_TEMPLATE.replace( "{id}", hex );
}

/*
================
discoverExtendedWorldSectors

The zones.json regions whose extracted data plane is complete. A region the
client never shipped map data for is reported, never built, and never
silently dropped from the sealed output.
================
*/
export async function discoverExtendedWorldSectors( options = {} ) {
	const sourceExtractedRoot = options.extractedRoot ?? extractedRoot;
	const zonesPath = options.zonesPath ?? defaultZonesPath();
	const zones = JSON.parse( await readFile( zonesPath, "utf8" ) );
	if ( zones.format !== "sro-extended-zone-list" ) {
		throw new Error( `${zonesPath} is not a sealed extended zone list` );
	}

	const sectors = [];
	const skipped = [];
	for ( const regionText of Object.keys( zones.zones ).sort() ) {
		const regionId = Number.parseInt( regionText, 16 );
		const sectorY = (regionId >> 8) & 0xff;
		const sectorX = regionId & 0xff;
		const mapRoot = path.join( sourceExtractedRoot, "Map_extracted" );
		const basePath = path.join( mapRoot, String( sectorY ), String( sectorX ) );
		const nvmPath = path.join(
			sourceExtractedRoot,
			"Data_extracted",
			"navmesh",
			`nv_${regionId.toString( 16 ).padStart( 4, "0" )}.nvm`
		);
		const requiredPaths = [ `${basePath}.m`, `${basePath}.t`, `${basePath}.o2`, nvmPath ];
		const present = await Promise.all( requiredPaths.map( ( filePath ) => exists( filePath ) ) );
		if ( present.some( ( value ) => !value ) ) {
			skipped.push( toHex16( regionId ) );
			continue;
		}
		sectors.push( {
			id: toHex16( regionId ),
			regionId,
			sectorX,
			sectorY,
			mapBasePath: basePath,
			navmeshPath: nvmPath
		} );
	}

	const bandsKept = {};
	for ( const sector of sectors ) {
		for ( const band of zones.zones[`0x${sector.regionId.toString( 16 ).padStart( 4, "0" )}`].bands ?? [] ) {
			bandsKept[band] = (bandsKept[band] ?? 0) + 1;
		}
	}
	const emptyBands = Object.keys( zones.anchorsByBand ?? {} )
		.map( Number )
		.filter( ( band ) => !bandsKept[band] );
	if ( emptyBands.length > 0 ) {
		throw new Error( `Extended world discovery would empty trajectory bands ${emptyBands.join( ", " )}` );
	}
	return { sectors, skipped, zonesPath, bandsKept };
}

/*
================
defaultZonesPath
================
*/
function defaultZonesPath() {
	return path.join( resolveExtendedGameDataRoot( process.env ), "zones.json" );
}

/*
================
buildExtendedWorldRegionIndexDescriptor

Same shape as the native outdoor index (the client loads both through one
loader); delivery is always prebuilt - the extended world has no dev-on-
demand adapter.
================
*/
export function buildExtendedWorldRegionIndexDescriptor( sectors ) {
	const sorted = [ ...sectors ].sort( ( left, right ) =>
		left.sectorY - right.sectorY || left.sectorX - right.sectorX
	);
	return {
		format: "sro-world-region-index",
		version: 2,
		area: EXTENDED_WORLD_AREA,
		regionSize: REGION_SIZE,
		seedRegionId: "0x0000",
		seedSector: { sectorX: 0, sectorY: 0 },
		bundleLayout: "one-region-per-bundle",
		deliveryMode: "prebuilt",
		regions: sorted.map( ( sector ) => ({
			id: sector.id,
			sectorX: sector.sectorX,
			sectorY: sector.sectorY,
			seedRegionId: sector.id,
			bundlePublicPath: extendedRegionBundlePublicPath( sector.id )
		}) )
	};
}

/*
================
loadNativeSharedRenderResources

The extended bundles reference the native shared sky/water file by public
path; the sky is the same world. Read it from the main checkout's published
tree (read-only) and keep the native validation contract: a missing or
malformed file is an operator error, not an extended-world fallback.
================
*/
async function loadNativeSharedRenderResources( options = {} ) {
	const sharedRenderPath = path.join(
		options.mainPublicRoot ?? MAIN_CHECKOUT_CLIENT_PUBLIC_ROOT,
		OUTDOOR_WORLD_SHARED_RENDER_PUBLIC_PATH.replace( /^\/+/, "" )
	);
	const shared = JSON.parse( await readFile( sharedRenderPath, "utf8" ) );
	if ( shared?.format !== "sro-world-shared-render-resources" || !shared.sky || !shared.water ) {
		throw new Error( `${sharedRenderPath} is not a published shared render resources file` );
	}
	return shared;
}

/*
================
buildExtendedSharedObjectResources

The extended object store: every object ANY extended sector places, its
BSR/BMT/mesh resources parsed once, meshes written content-addressed under
the extended namespace. The store always spans the full sector list - a
partial --region run must never shrink the index the other bundles' mesh
references resolve through - and an existing valid store is reused unless
--force, exactly like the native shared-object lane. Placements whose
object.ifo id is unknown fail the build.
================
*/
async function buildExtendedSharedObjectResources( sectors, options ) {
	const outputPath = publicPathToFile( EXTENDED_WORLD_OBJECT_INDEX_PUBLIC_PATH, publicRoot );
	if ( !options.force && (await exists( outputPath )) ) {
		const index = JSON.parse( await readFile( outputPath, "utf8" ) );
		if ( index?.format !== "sro-world-object-resource-index" || !Array.isArray( index.meshFiles ) ) {
			throw new Error( `${EXTENDED_WORLD_OBJECT_INDEX_PUBLIC_PATH} is not an extended object resource index` );
		}
		return {
			index,
			collisionResources: {
				...index,
				format: "sro-title-sector-object-resources",
				meshes: index.meshFiles.map( ( mesh ) => ({ sourcePath: mesh.sourcePath, bounds: mesh.bounds }) )
			}
		};
	}

	const objectInfoPath = path.join( options.extractedRoot, "Map_extracted", "object.ifo" );
	const objectInfo = await readJmxMapObjectInfo( objectInfoPath );
	const usage = await collectExtendedObjectUsage( sectors, options );

	const missingObjectIds = usage.objectIds.filter( ( objectId ) => !objectInfo.entriesById[String( objectId )] );
	if ( missingObjectIds.length > 0 ) {
		throw new Error( `Extended MAPO2 sectors reference missing object.ifo ids: ${missingObjectIds.join( ", " )}` );
	}

	const resources = await buildTitleSectorObjectResources( {
		area: EXTENDED_WORLD_AREA,
		extractedRoot: options.extractedRoot,
		gameRoot: options.gameRoot,
		objectDefinitions: usage.objectIds.map( ( objectId ) => {
			const entry = objectInfo.entriesById[String( objectId )];
			return { objectId, flags: entry.flags, sourcePath: entry.sourcePath };
		} ),
		placements: [],
		placementCountsByObjectId: usage.placementCountsByObjectId
	} );

	const meshFiles = [];
	const collisionMeshes = [];
	const meshes = resources.meshes;
	for ( let index = 0; index < meshes.length; index += 1 ) {
		const mesh = meshes[index];
		const wrapper = { format: "sro-world-object-mesh-resource", version: 1, mesh };
		const bytes = Buffer.from( `${JSON.stringify( wrapper )}\n`, "utf8" );
		const sha256 = sha256Hex( bytes );
		const publicPath = `${EXTENDED_WORLD_OBJECT_MESH_ROOT_PUBLIC_PATH}/${sha256}.json`;
		const meshOutputPath = publicPathToFile( publicPath, publicRoot );
		if ( options.force || !(await exists( meshOutputPath )) ) {
			await writeIntoPublicTree( meshOutputPath, bytes );
		}
		meshFiles.push( {
			sourcePath: mesh.sourcePath,
			publicPath,
			sha256,
			byteLength: mesh.byteLength,
			vertexCount: mesh.vertexCount,
			triangleCount: mesh.triangleCount,
			bounds: mesh.bounds
		} );
		collisionMeshes.push( { sourcePath: mesh.sourcePath, bounds: mesh.bounds } );
		meshes[index] = undefined;
	}

	const { meshes: _discardedMeshes, ...resourceIndexFields } = resources;
	const index = {
		...resourceIndexFields,
		format: "sro-world-object-resource-index",
		version: 1,
		meshFiles
	};
	await writeIntoPublicTree( outputPath, Buffer.from( `${JSON.stringify( index )}\n`, "utf8" ) );
	return {
		index,
		collisionResources: { ...resources, format: "sro-title-sector-object-resources", meshes: collisionMeshes }
	};
}

/*
================
collectExtendedObjectUsage
================
*/
async function collectExtendedObjectUsage( sectors, options ) {
	const objectIds = new Set();
	const placementCountsByObjectId = new Map();
	await mapWithConcurrency( sectors, options.jobs, async ( sector ) => {
		const placements = parseJmxMapObjectPlacementO2(
			await readFile( `${sector.mapBasePath}.o2` ),
			`${sector.mapBasePath}.o2`
		);
		for ( const placement of placements.placements ) {
			objectIds.add( placement.objectId );
			placementCountsByObjectId.set(
				placement.objectId,
				(placementCountsByObjectId.get( placement.objectId ) ?? 0) + 1
			);
		}
	} );
	return {
		objectIds: [ ...objectIds ].sort( ( left, right ) => left - right ),
		placementCountsByObjectId
	};
}

/*
================
buildExtendedWorldRegionResources
================
*/
export async function buildExtendedWorldRegionResources( options = {} ) {
	const sourceExtractedRoot = options.extractedRoot ?? extractedRoot;
	const sourceGameRoot = options.gameRoot ?? gameRoot;
	const jobs = options.jobs ?? buildJobs();
	const discovery = options.sectors ??
		(await discoverExtendedWorldSectors( { extractedRoot: sourceExtractedRoot, zonesPath: options.zonesPath } ));
	const { sectors, skipped } = discovery;
	const descriptor = buildExtendedWorldRegionIndexDescriptor( sectors );
	const selectedSectors = selectRequestedSectors( sectors, options.regionIds );

	if ( options.planOnly ) {
		return {
			planOnly: true,
			sectorCount: sectors.length,
			skippedRegionCount: skipped.length,
			selectedSectorCount: selectedSectors.length,
			descriptor
		};
	}

	const mapRoot = path.join( sourceExtractedRoot, "Map_extracted" );
	const [objectInfo, tileCatalog, sharedRender] = await Promise.all( [
		readJmxMapObjectInfo( path.join( mapRoot, "object.ifo" ) ),
		readJmxMapTileCatalog( path.join( mapRoot, "tile2d.ifo" ) ),
		loadNativeSharedRenderResources( options )
	] );
	const sharedObjects = await buildExtendedSharedObjectResources( sectors, {
		extractedRoot: sourceExtractedRoot,
		gameRoot: sourceGameRoot,
		jobs,
		force: Boolean( options.force )
	} );

	let built = 0;
	let reused = 0;
	await mapWithConcurrency( selectedSectors, jobs, async ( sector ) => {
		const outputPath = publicPathToFile( extendedRegionBundlePublicPath( sector.id ), publicRoot );
		if ( !options.force && (await exists( outputPath )) ) {
			reused += 1;
			reportProgress( options, {
				completed: built + reused,
				total: selectedSectors.length,
				regionId: sector.id,
				status: "reused"
			} );
			return;
		}
		const bundle = await buildJmxWorldRegionBundle( {
			area: EXTENDED_WORLD_AREA,
			sectorId: sector.regionId,
			sectorX: sector.sectorX,
			sectorY: sector.sectorY,
			terrainSectorMargin: 0,
			extractedRoot: sourceExtractedRoot,
			gameRoot: sourceGameRoot,
			objectInfo,
			tileCatalog,
			objectResources: sharedObjects.collisionResources,
			objectResourceIndexPublicPath: EXTENDED_WORLD_OBJECT_INDEX_PUBLIC_PATH,
			sharedRenderResourcesPublicPath: OUTDOOR_WORLD_SHARED_RENDER_PUBLIC_PATH,
			skyResources: sharedRender.sky,
			waterResources: sharedRender.water
		} );
		if ( normalizeRegionId( bundle.source.sectorId ) !== sector.id || bundle.terrain.sectorCount !== 1 ) {
			throw new Error( `Extended ${sector.id} did not emit an independent one-sector bundle` );
		}
		await writeIntoPublicTree( outputPath, Buffer.from( `${JSON.stringify( bundle )}\n`, "utf8" ) );
		built += 1;
		reportProgress( options, {
			completed: built + reused,
			total: selectedSectors.length,
			regionId: sector.id,
			status: "built"
		} );
	} );

	// Every bundle must exist before the routing planes publish.
	const missingBundlePaths = [];
	await mapWithConcurrency( descriptor.regions, jobs, async ( region ) => {
		if ( !(await exists( publicPathToFile( region.bundlePublicPath, publicRoot ) )) ) {
			missingBundlePaths.push( region.bundlePublicPath );
		}
	} );
	if ( missingBundlePaths.length > 0 ) {
		throw new Error( `Extended world routing would publish with ${missingBundlePaths.length} missing bundle(s)` );
	}

	const catalog = await publishExtendedRouting( descriptor );
	const minimap = await publishExtendedMinimapTiles( sectors );
	const movement = await buildExtendedMovementProjection( descriptor );
	const manifest = await sealExtendedWorldManifest( discovery, descriptor, sharedObjects.index );

	return {
		planOnly: false,
		sectorCount: sectors.length,
		skippedRegions: skipped,
		selectedSectorCount: selectedSectors.length,
		built,
		reused,
		jobs,
		descriptor,
		catalogPublicPath: EXTENDED_WORLD_CATALOG_PUBLIC_PATH,
		movementRoot: movement.movementRoot,
		manifestPath: manifest.manifestPath,
		sharedObjectIndex: {
			publicPath: EXTENDED_WORLD_OBJECT_INDEX_PUBLIC_PATH,
			bsrCount: sharedObjects.index.bsrCount,
			meshCount: sharedObjects.index.meshFiles.length
		},
		minimap,
		worldManifest: manifest
	};
}

/*
================
publishExtendedRouting

The client catalog overlay the extended client merges over the native
catalog, plus the extended region index it routes through.
================
*/
async function publishExtendedRouting( descriptor ) {
	const regionsById = {};
	for ( const region of descriptor.regions ) {
		regionsById[region.id] = [ {
			id: region.id,
			area: EXTENDED_WORLD_AREA,
			seedRegionId: region.seedRegionId,
			seedSector: { sectorX: region.sectorX, sectorY: region.sectorY },
			sectorX: region.sectorX,
			sectorY: region.sectorY,
			worldRegionsPublicPath: EXTENDED_WORLD_INDEX_PUBLIC_PATH,
			bundlePublicPath: region.bundlePublicPath,
			source: EXTENDED_WORLD_SOURCE_NAME
		} ];
	}
	await writePublicJson( EXTENDED_WORLD_INDEX_PUBLIC_PATH, descriptor );
	const catalog = {
		format: "sro-world-region-catalog",
		version: 1,
		regionsById
	};
	await writePublicJson( EXTENDED_WORLD_CATALOG_PUBLIC_PATH, catalog );
	return catalog;
}

/*
================
publishExtendedMinimapTiles

The 2026 outdoor minimap tiles (Media.pk2 minimap/<x>x<z>.ddj, converted
to PNG in the image staging) published under the native tile path - the
minimap renderer resolves one image per region from its art set - plus
the extended art catalog the extended client overlays. A region the
client ships no tile for draws no minimap, exactly like a native edge
region; it is counted, never fatal.
================
*/
async function publishExtendedMinimapTiles( sectors ) {
	const tilePaths = [];
	let missing = 0;
	for ( const sector of sectors ) {
		const relative = `Media_extracted/minimap/${sector.sectorX}x${sector.sectorY}.png`;
		const source = path.join( imageSourceRoot, ...relative.split( "/" ) );
		if ( !(await exists( source )) ) {
			missing += 1;
			continue;
		}
		const publicPath = `/assets/images/${relative}`;
		await copyIntoPublicTree( source, publicPathToFile( publicPath, publicRoot ) );
		tilePaths.push( publicPath );
	}
	if ( tilePaths.length === 0 ) {
		throw new Error( "Extended world has no minimap tiles: run the image conversion over the extended extraction" );
	}
	await writePublicJson( EXTENDED_MINIMAP_CATALOG_PUBLIC_PATH, {
		format: "sro-mission-dungeon-minimap-manifest",
		version: 1,
		dungeons: [],
		tilePaths: tilePaths.sort()
	} );
	return { published: tilePaths.length, missing };
}

/*
================
buildExtendedMovementProjection

The server movement mirror under the extended projection's movement/ root,
reusing the native projections (projectRegionBundle, object-nav, authority
paths) over the extended bundles. The movement authority chain resolves
native regions from the native root first, so the extended mirror only
ever needs the 2026 regions.
================
*/
async function buildExtendedMovementProjection( descriptor, options = {} ) {
	const projectionRoot = options.projectionRoot ?? resolveExtendedGameDataRoot( process.env );
	const authorityRoot = path.join( projectionRoot, "movement" );
	const referencedObjectNav = new Map();

	const regionsById = {};
	for ( const region of descriptor.regions ) {
		regionsById[region.id] = [ {
			id: region.id,
			area: EXTENDED_WORLD_AREA,
			seedRegionId: region.seedRegionId,
			seedSector: { sectorX: region.sectorX, sectorY: region.sectorY },
			sectorX: region.sectorX,
			sectorY: region.sectorY,
			worldRegionsPath: worldAuthorityPath( EXTENDED_WORLD_INDEX_PUBLIC_PATH ),
			bundlePath: worldAuthorityPath( region.bundlePublicPath ),
			source: EXTENDED_WORLD_SOURCE_NAME
		} ];
	}

	const indexPath = worldAuthorityPath( EXTENDED_WORLD_INDEX_PUBLIC_PATH );
	await writePlainJson(
		path.join( authorityRoot, indexPath ),
		{
			format: "sro-server-world-region-index",
			version: 1,
			regionSize: descriptor.regionSize,
			seedRegionId: descriptor.seedRegionId,
			regions: descriptor.regions.map( ( region ) => ({
				id: region.id,
				bundlePath: worldAuthorityPath( region.bundlePublicPath )
			}) )
		}
	);
	// The object-nav projections read referenced indexes and meshes from
	// the published world plane (buildServerGameDataBundle's worldRoot
	// contract: the client-public tree's assets/world directory).
	const worldPublicRoot = path.join( publicRoot, "assets", "world" );
	for ( const region of descriptor.regions ) {
		const relativePath = worldAuthorityPath( region.bundlePublicPath );
		const source = JSON.parse(
			await readFile( publicPathToFile( region.bundlePublicPath, publicRoot ), "utf8" )
		);
		const objectProjection = await buildObjectNavProjection(
			authorityRoot,
			relativePath,
			source,
			worldPublicRoot,
			referencedObjectNav
		);
		await writePlainJson(
			path.join( authorityRoot, relativePath ),
			projectRegionBundle( source, objectProjection )
		);
	}
	await writePlainJson( path.join( authorityRoot, "catalog.json" ), {
		format: "sro-server-movement-catalog",
		version: 1,
		regionsById
	} );
	return { authorityRoot, movementRoot: authorityRoot };
}

/*
================
sealExtendedWorldManifest

Content digests over everything the lane published; two clean runs produce
identical bytes. Sits beside the data-lane manifest (manifest.json) in the
extended projection root and never touches it.
================
*/
async function sealExtendedWorldManifest( discovery, descriptor, objectIndex ) {
	const projectionRoot = resolveExtendedGameDataRoot( process.env );
	const digests = {};
	for (
		const publicPath of [
			EXTENDED_WORLD_INDEX_PUBLIC_PATH,
			EXTENDED_WORLD_CATALOG_PUBLIC_PATH,
			EXTENDED_WORLD_OBJECT_INDEX_PUBLIC_PATH,
			EXTENDED_MINIMAP_CATALOG_PUBLIC_PATH,
			...descriptor.regions.map( ( region ) => region.bundlePublicPath )
		]
	) {
		digests[publicPath] = sha256Hex( await readFile( publicPathToFile( publicPath, publicRoot ) ) );
	}
	const movementCatalogPath = path.join( projectionRoot, "movement", "catalog.json" );
	const manifest = {
		format: "sro-extended-world-manifest",
		version: 1,
		sourceClient: "isro-live-2026",
		zonesPath: path.relative( projectionRoot, discovery.zonesPath ).replaceAll( "\\", "/" ),
		regionCount: descriptor.regions.length,
		skippedRegions: discovery.skipped,
		bands: discovery.bandsKept,
		objectBsrCount: objectIndex.bsrCount,
		objectMeshCount: objectIndex.meshFiles.length,
		movementCatalogSha256: sha256Hex( await readFile( movementCatalogPath ) ),
		contentSha256: digests
	};
	const manifestPath = path.join( projectionRoot, EXTENDED_WORLD_MANIFEST_NAME );
	await writeFile( manifestPath, `${JSON.stringify( manifest, null, "\t" )}\n`, "utf8" );
	return { manifestPath, manifest };
}

/*
================
selectRequestedSectors
================
*/
function selectRequestedSectors( sectors, requestedIds ) {
	if ( !requestedIds || requestedIds.length === 0 ) {
		return sectors;
	}
	const requested = new Set( requestedIds.map( normalizeRegionId ) );
	const selected = sectors.filter( ( sector ) => requested.has( normalizeRegionId( sector.id ) ) );
	const found = new Set( selected.map( ( sector ) => normalizeRegionId( sector.id ) ) );
	const missing = [ ...requested ].filter( ( id ) => !found.has( id ) );
	if ( missing.length > 0 ) {
		throw new Error( `Requested extended region(s) are not complete extracted sectors: ${missing.join( ", " )}` );
	}
	return selected;
}

/*
================
publicPathToFile
================
*/
function publicPathToFile( publicPath, root ) {
	return path.join( root, publicPath.replace( /^\/+/, "" ) );
}

/*
================
writePublicJson / writePlainJson

Public-tree writes go through the publication ledger's writer (every file
players can download is claimed); projection-root writes are plain.
================
*/
async function writePublicJson( publicPath, value ) {
	await writeIntoPublicTree(
		publicPathToFile( publicPath, publicRoot ),
		Buffer.from( `${JSON.stringify( value )}\n`, "utf8" )
	);
}

async function writePlainJson( outputPath, value ) {
	await mkdir( path.dirname( outputPath ), { recursive: true } );
	await writeFile( outputPath, `${JSON.stringify( value )}\n`, "utf8" );
}

/*
================
reportProgress
================
*/
function reportProgress( options, progress ) {
	if ( typeof options.onProgress === "function" ) {
		options.onProgress( progress );
		return;
	}
	if ( progress.completed === progress.total || progress.completed === 1 || progress.completed % 25 === 0 ) {
		console.log(
			`[world:extended] ${progress.completed}/${progress.total} ${progress.status} ${progress.regionId}`
		);
	}
}
