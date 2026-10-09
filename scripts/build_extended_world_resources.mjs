/*
===========================================================================

build_extended_world_resources.mjs - the extended outdoor world build entry

Extended content (isro-live-2026), port-only, not v1.150-native. Builds the
live-2026 outdoor regions into the PRIVATE tree:

  SRO_GENERATED_ROOT=<private tree>            output root (client-public)
  SRO_GAME_ROOT=<private>/extended/game        the 2026 extraction
  SRO_EXTENDED_GAME_DATA_ROOT=<projection>     movement mirror + manifest

Run scripts/extract_extended_world_data.py first (it writes the extraction
and the zone-verified inventory). Nothing here touches the main checkout
except reading its published shared sky/water render resources.

===========================================================================
*/
// First: it sizes libuv's thread pool before anything starts it.
import "./build/shared/buildParallelism.mjs";
import { buildExtendedWorldRegionResources } from "./build/world/buildExtendedWorldRegionResources.mjs";
import { runConvertImages } from "./build/shared/convertImagesRunner.mjs";
import { beginPublication, commitPublication } from "./build/shared/publicationLedger.mjs";
import { withGeneratedAssetsLock } from "./rebuildLock.mjs";

const options = parseArguments( process.argv.slice( 2 ) );

if ( options.help ) {
	console.log( `Build the extended (live-2026) outdoor world as streamable regions.

Usage:
  node scripts/build_extended_world_resources.mjs [options]

Options:
  --plan                 Discover and report without writing files.
  --force                Rebuild selected region files and shared resources.
  --jobs=N               Concurrent region builders (default SRO_BUILD_JOBS, else cores - 1).
  --region=HEX[,HEX...]  Incrementally build specific zone regions (full routing still publishes).
  --help                 Show this help.

Environment (all three are required for a real build):
  SRO_GENERATED_ROOT          the private tree (output)
  SRO_GAME_ROOT               <private>/extended/game (the 2026 extraction)
  SRO_EXTENDED_GAME_DATA_ROOT the extended projection (movement mirror + manifest)` );
	process.exit( 0 );
}

/*
================
runBuild

Convert the source images the extended writers read (2026 DDJ staging),
then build the regions and the routing planes.
================
*/
const runBuild = async () => {
	if ( !options.planOnly ) {
		const sourceImages = await runConvertImages( [] );
		if ( sourceImages.status !== 0 ) {
			throw new Error( `Source image conversion failed with exit status ${sourceImages.status}.` );
		}
		beginPublication( "extended-world", { complete: true } );
	}
	try {
		const result = await buildExtendedWorldRegionResources( options );
		if ( result.planOnly ) {
			console.log(
				`Extended world plan: ${result.sectorCount} complete sectors ` +
					`(${result.skippedRegionCount} skipped by the client), ${result.selectedSectorCount} selected.`
			);
			return;
		}
		console.log(
			`Extended regions: ${result.built} built, ${result.reused} reused of ${result.sectorCount}; ` +
				`${result.skippedRegions.length} region(s) absent from the client's map data; ` +
				`shared objects ${result.sharedObjectIndex.bsrCount} BSR / ${result.sharedObjectIndex.meshCount} BMS.`
		);
		console.log( `Client overlay ${result.catalogPublicPath}; movement mirror ${result.movementRoot}.` );
		console.log(
			`World manifest ${result.manifestPath} (regionCount ${result.worldManifest.manifest.regionCount}).`
		);
	} finally {
		if ( !options.planOnly ) await commitPublication();
	}
};

if ( options.planOnly ) {
	await runBuild();
} else {
	await withGeneratedAssetsLock( "extended world resource build", runBuild );
}

/*
================
parseArguments
================
*/
function parseArguments( args ) {
	const parsed = {
		planOnly: false,
		force: false,
		jobs: undefined,
		regionIds: [],
		help: false
	};

	for ( const argument of args ) {
		if ( argument === "--plan" ) {
			parsed.planOnly = true;
		} else if ( argument === "--force" ) {
			parsed.force = true;
		} else if ( argument === "--help" || argument === "-h" ) {
			parsed.help = true;
		} else if ( argument.startsWith( "--jobs=" ) ) {
			parsed.jobs = Number.parseInt( argument.slice( "--jobs=".length ), 10 );
		} else if ( argument.startsWith( "--region=" ) ) {
			parsed.regionIds.push(
				...argument
					.slice( "--region=".length )
					.split( "," )
					.map( ( value ) => value.trim() )
					.filter( Boolean )
			);
		} else {
			throw new Error( `Unknown extended world build option ${argument}` );
		}
	}

	return parsed;
}
