/*
===========================================================================

environment-stages.test.mjs - the experimental environment stages on the title terrain

Renders the Constantinople dock fixture twice at one camera: with every
stage off (the native frame, asserted byte-stable across repeats) and with
sun direction, terrain relief and the textured horizon on. The stages must
change the frame without ever failing the device - and the float bloom
chain must composite, resize and hand quality back without error.
===========================================================================
*/
import { CLIENT_PUBLIC_ROOT } from "../../../../scripts/lib/generatedRoot.mjs";
import { test } from "node:test";
import assert from "node:assert/strict";
import {
	readPublishedAssetJsonSync as read,
	readPublishedAssetBytesSync as bytes
} from "../../../../scripts/lib/publishedAsset.mjs";
import { launchProbeBrowser } from "../../../../scripts/lib/probeBrowser.mjs";
import { CLIENT_NEXT_BASE_URL } from "../../../../scripts/lib/probeEndpoints.mjs";

test(
	"sun direction, terrain relief and the textured horizon change the frame without breaking the device",
	{ timeout: 300000 },
	async () => {
		const publicRoot = CLIENT_PUBLIC_ROOT + "/";
		const bundle = read( "/assets/world/constantinople/region-694e.json", publicRoot );
		const resources = bundle.objects.resources;
		const refs = resources.bsr.filter( r => r.objectId === 1630 );
		bundle.objects.placements = bundle.objects.placements.filter( p => p.objectId === 1630 );
		resources.bsr = refs;
		resources.meshes = resources.meshes.filter( m => refs.some( r => r.meshPaths.includes( m.sourcePath ) ) );
		resources.materialSets = resources.materialSets.filter( m =>
			refs.some( r => r.materialPaths.includes( m.sourcePath ) )
		);
		for ( const sector of bundle.terrain.sectors ?? [] ) {
			sector.blocks = sector.blocks.filter( b => {
				const x = (sector.sectorX - bundle.source.sectorX) * 6 + b.blockX,
					z = (sector.sectorY - bundle.source.sectorY) * 6 + b.blockZ;
				return x >= 12 && x <= 19 && z >= -1 && z <= 5;
			} );
		}
		const camera = read( "/assets/title/constantinople/manifest.json", publicRoot ).camera;
		const { browser, page } = await launchProbeBrowser();
		try {
			await page.route( "**/assets/**", route => {
				const path = new URL( route.request().url() ).pathname;
				if ( !path.startsWith( "/assets/" ) ) return route.continue();
				try {
					return route.fulfill( { body: Buffer.from( bytes( path, publicRoot ) ) } );
				} catch {
					return route.abort();
				}
			} );
			await page.goto( CLIENT_NEXT_BASE_URL );
			const result = await page.evaluate( async ( { bundle, camera } ) => {
				const { decodeDxt1 } = await import( "/src/engine/foundation/assets/dds.ts" );
				const { decodeNativeTexture, decodeNativeTextureLevel } = await import(
					"/src/engine/foundation/assets/native-texture.ts"
				);
				const { createRenderer } = await import( "/src/engine/runtime/renderer/renderer.ts" );
				const { createWorldDecoder } = await import( "/src/engine/runtime/assets/worker/world/world.ts" );
				const { sampleFrontendCamera, frontendCameraView } = await import(
					"/src/engine/foundation/rendering/frontend-camera.ts"
				);
				const canvas = document.createElement( "canvas" ),
					r = createRenderer(
						canvas,
						(await import( "/src/engine/runtime/random/random.ts" )).createPresentationRandom( 1 )
					),
					out = document.createElement( "canvas" );
				out.width = 1823;
				out.height = 845;
				const ctx = out.getContext( "2d" );
				if ( !ctx ) throw Error( "2d context unavailable" );
				const decoder = createWorldDecoder();
				// Heightfield normals are a decode-time option (the relief stage's
				// data); the retail flat normals are the default.
				const scene = { ...decoder.decode( bundle, true, { terrainNormals: true } ), terrainDetail: "full" };
				const textures = new Map();
				for (
					const path of new Set( [
						...(scene.flareTextures ?? []),
						...scene.groups.flatMap( g =>
							g.material.frames ?? (g.material.texture ? [ g.material.texture ] : [])
						)
					] )
				) {
					const blob = await (await fetch( path )).blob();
					if ( path.endsWith( ".dds" ) ) {
						const d = decodeDxt1( new Uint8Array( await blob.arrayBuffer() ) );
						textures.set( path, await createImageBitmap( new ImageData( d.pixels, d.width, d.height ) ) );
					} else if ( path.endsWith( ".texture" ) ) {
						const native = decodeNativeTexture( new Uint8Array( await blob.arrayBuffer() ) ),
							level = decodeNativeTextureLevel( native, 0 );
						textures.set(
							path,
							await createImageBitmap(
								new ImageData( new Uint8ClampedArray( level.buffer ), native.width, native.height )
							)
						);
					} else textures.set( path, await createImageBitmap( blob ) );
				}
				const deadline = performance.now() + 15000;
				while ( r.phase() === "starting" && performance.now() < deadline ) {
					await new Promise( requestAnimationFrame );
				}
				if ( r.phase() !== "running" ) throw Error( r.error() ?? "GPU startup deadline" );
				// Publish every texture once: without them the drain below never
				// settles and the submit loop would pile frames forever.
				for ( const [p, b] of textures ) r.setWorldTexture( p, await createImageBitmap( b ) );
				const stages = ( dynamicSun, terrainRelief, texturedHorizon ) =>
					r.experimentalVideo( {
						postProcessing: false,
						anisotropicFiltering: false,
						heightFog: false,
						dynamicSun,
						terrainRelief,
						texturedHorizon,
						floatBloom: false
					} );
				const submit = id => {
					r.setWorld( { ...scene, id } );
					r.setWorldCamera( frontendCameraView( sampleFrontendCamera( camera, 0.7 ) ) );
					let k = 0;
					return (async () => {
						do {
							r.frame( { width: 1823, height: 845 }, k++ * .016 );
							await new Promise( requestAnimationFrame );
						} while ( k < 10 || r.worldStats().pendingGroups || r.worldStats().pendingTextures );
					})();
				};
				// Two windows: the full frame proves the stages moved it, and a
				// lower-band window (outside the sky's wall-clock star flicker)
				// proves the native frame repeats within rounding.
				const capture = async () => {
					const bitmap = await createImageBitmap( canvas );
					ctx.drawImage( bitmap, 0, 0 );
					bitmap.close();
					return {
						full: Array.from( ctx.getImageData( 0, 0, 1823, 845 ).data ),
						still: Array.from( ctx.getImageData( 300, 450, 400, 280 ).data )
					};
				};
				stages( false, false, false );
				await submit( "native-a" );
				const native = await capture();
				await submit( "native-b" );
				const repeat = await capture();
				stages( true, true, true );
				await submit( "staged-a" );
				const staged = await capture();
				r.dispose();
				return { native, repeat, staged, error: r.error() };
			}, { bundle, camera } );
			assert.equal( result.error, null );
			// The native frame is stable across repeat submits: rounding-level
			// drift only (the sky's wall-clock flicker sits outside the crop).
			let drift = 0;
			for ( let i = 0; i < result.native.still.length; i++ ) {
				assert.ok( Math.abs( result.native.still[i] - result.repeat.still[i] ) <= 1 );
				if ( result.native.still[i] !== result.repeat.still[i] ) drift++;
			}
			assert.ok( drift < 2000, `repeat drift too wide: ${drift}` );
			// The stages visibly move the frame (sun raking, slope shading, the
			// textured band), more than a rounding-level flutter.
			let changed = 0;
			for ( let i = 0; i < result.native.full.length; i += 4 ) {
				if (
					result.native.full[i] !== result.staged.full[i] ||
					result.native.full[i + 1] !== result.staged.full[i + 1] ||
					result.native.full[i + 2] !== result.staged.full[i + 2]
				) changed++;
			}
			assert.ok(
				changed > 2000,
				`sun/relief/horizon stages must change the frame: ${changed} pixels differ`
			);
		} finally {
			await browser.close();
		}
	}
);

test( "the float bloom chain composites, resizes and hands quality back", { timeout: 60000 }, async () => {
	const { browser, page } = await launchProbeBrowser();
	page.on( "console", m => {
		if ( m.type() === "error" ) console.log( m.text() );
	} );
	try {
		await page.goto( CLIENT_NEXT_BASE_URL );
		const result = await page.evaluate( async () => {
			const { createDevice } = await import( "/src/engine/runtime/renderer/device/device.ts" );
			const device = createDevice(), canvas = document.createElement( "canvas" );
			document.body.append( canvas );
			try {
				const deadline = performance.now() + 15000;
				while ( device.phase() === "starting" ) {
					if ( performance.now() > deadline ) throw Error( "Device timeout" );
					await new Promise( requestAnimationFrame );
				}
				if ( device.error() ) throw Error( device.error() );
				const context = canvas.getContext( "webgpu" );
				if ( !context ) throw Error( "webgpu context unavailable" );
				device.surfaceCommands().configure( context, device.format() );
				const pixels = [];
				for ( const [size, value] of [ [ 64, 0 ], [ 128, 128 ], [ 64, 255 ] ] ) {
					device.experimentalVideo( {
						postProcessing: false,
						anisotropicFiltering: false,
						heightFog: false,
						dynamicSun: false,
						terrainRelief: false,
						texturedHorizon: false,
						floatBloom: true
					} );
					canvas.width = canvas.height = size;
					const bloom = device.bloom( size, size, true ),
						out = device.surfaceCommands().createColor( size, size ),
						encoder = device.commands().createEncoder();
					const pass = encoder.beginRenderPass( {
						colorAttachments: [ {
							view: bloom.view,
							clearValue: [ value / 255, value / 255, value / 255, 1 ],
							loadOp: "clear",
							storeOp: "store"
						} ]
					} );
					pass.end();
					bloom.encode( encoder, out.view );
					device.commands().submit( encoder.finish() );
					out.present( context.getCurrentTexture() );
					const copy = document.createElement( "canvas" );
					copy.width = copy.height = size;
					const ctx = copy.getContext( "2d" );
					if ( !ctx ) throw Error( "2d context unavailable" );
					const image = await createImageBitmap( canvas );
					ctx.drawImage( image, 0, 0 );
					image.close();
					pixels.push( [ ...ctx.getImageData( size / 2, size / 2, 1, 1 ).data ] );
					out.dispose();
				}
				// Handing quality back to the native chain still draws.
				device.experimentalVideo( {
					postProcessing: false,
					anisotropicFiltering: false,
					heightFog: false,
					dynamicSun: false,
					terrainRelief: false,
					texturedHorizon: false,
					floatBloom: false
				} );
				const native = device.bloom( 64, 64, true );
				return { pixels, native: native !== undefined, error: device.error() };
			} finally {
				device.dispose();
				canvas.remove();
			}
		} );
		assert.equal( result.error, null );
		assert.ok( result.native );
		// Dark stays dark; the mid-gray and bright fields gain the float glow.
		assert.deepEqual( result.pixels[0], [ 0, 0, 0, 255 ] );
		for ( const c of result.pixels[1].slice( 0, 3 ) ) assert.ok( c >= 129, JSON.stringify( result ) );
		assert.deepEqual( result.pixels[2], [ 255, 255, 255, 255 ] );
	} finally {
		await browser.close();
	}
} );
