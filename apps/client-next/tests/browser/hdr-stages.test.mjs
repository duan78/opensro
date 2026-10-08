/*
===========================================================================

hdr-stages.test.mjs - the 2026-10-08 lighting stages in isolation

A synthetic, textureless scene owns its geometry and its light: a lit
ground quad, a dome whose smooth normals separate per-vertex from
per-pixel lighting, and a box that casts. Each stage cycles off-on-off;
every disabled capture must restore the native pixels exactly.

===========================================================================
*/
import { test } from "node:test";
import assert from "node:assert/strict";
import { launchProbeBrowser } from "../../../../scripts/lib/probeBrowser.mjs";
import { CLIENT_NEXT_BASE_URL } from "../../../../scripts/lib/probeEndpoints.mjs";

/*
================
captureLightingStages
================
*/
async function captureLightingStages() {
	const { createRenderer } = await import( "/src/engine/runtime/renderer/renderer.ts" );
	const { createPresentationRandom } = await import( "/src/engine/runtime/random/random.ts" );
	const width = 192, height = 120;
	const canvas = document.createElement( "canvas" );
	const output = document.createElement( "canvas" );
	output.width = width;
	output.height = height;
	const context = output.getContext( "2d", { willReadFrequently: true } );
	if ( !context ) throw Error( "No capture context" );
	const renderer = createRenderer( canvas, createPresentationRandom( 1 ) );
	const identity = Float32Array.of( 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1 );
	const options = {
		postProcessing: false,
		anisotropicFiltering: false,
		heightFog: false,
		dynamicSun: false,
		terrainRelief: false,
		texturedHorizon: false,
		floatBloom: false,
		hdrToneMap: false,
		sunShadow: false,
		perPixelLighting: false
	};
	// The dome: one apex, eight rim vertices with smooth outward normals,
	// so the saturated per-vertex colours and the per-pixel re-evaluation
	// disagree across its faces. The box sits sunward (+x) of the dome, so
	// the retail diagonal light casts it across the dome's ground.
	const SEGMENTS = 8;
	const positions = [ 0, 10, 0 ], normals = [ 0, 1, 0 ], indices = [];
	for ( let i = 0; i <= SEGMENTS; i++ ) {
		const angle = i / SEGMENTS * Math.PI * 2, x = Math.cos( angle ) * 18, z = Math.sin( angle ) * 18;
		positions.push( x, 0, z );
		const length = Math.hypot( x, 8, z );
		normals.push( x / length, 8 / length, z / length );
	}
	for ( let i = 0; i < SEGMENTS; i++ ) indices.push( 0, 1 + i, 2 + i );
	// A cube: 8 corners, 12 triangles, flat outward normals.
	const cube = [], cubeNormals = [], cubeIndices = [];
	const corners = [
		[ 0, 0, 0 ],
		[ 8, 0, 0 ],
		[ 8, 0, 8 ],
		[ 0, 0, 8 ],
		[ 0, 8, 0 ],
		[ 8, 8, 0 ],
		[ 8, 8, 8 ],
		[ 0, 8, 8 ]
	];
	const faces = [
		[ [ 0, 3, 2, 1 ], [ 0, -1, 0 ] ],
		[ [ [ 4 ], [ 5 ], [ 6 ], [ 7 ] ].flat(), [ 0, 1, 0 ] ],
		[ [ 0, 1, 5, 4 ], [ 0, 0, -1 ] ],
		[ [ [ 2 ], [ 3 ], [ 7 ], [ 6 ] ].flat(), [ 0, 0, 1 ] ],
		[ [ 1, 2, 6, 5 ], [ 1, 0, 0 ] ],
		[ [ [ 0 ], [ 4 ], [ 7 ], [ 3 ] ].flat(), [ -1, 0, 0 ] ]
	];
	for ( const [quad, normal] of faces ) {
		const base = cube.length / 3;
		for ( const corner of quad ) {
			cube.push( corners[corner][0] + 14, corners[corner][1], corners[corner][2] - 4 );
			cubeNormals.push( normal[0], normal[1], normal[2] );
		}
		cubeIndices.push( base, base + 1, base + 2, base, base + 2, base + 3 );
	}
	const lit = shade => ({
		color: shade,
		unlit: false,
		objectLight: 1,
		alphaCutoff: 0,
		blend: false,
		doubleSided: true
	});
	try {
		const deadline = performance.now() + 15000;
		while ( renderer.phase() === "starting" && performance.now() < deadline ) {
			await new Promise( requestAnimationFrame );
		}
		if ( renderer.phase() !== "running" ) throw Error( renderer.error() ?? "GPU startup deadline" );
		renderer.setWorldCamera( { eye: [ 6, 26, -64 ], target: [ 4, 2, 0 ], fov: 1.1, near: 1, far: 500 } );
		renderer.setWorld( {
			id: "hdr-stages",
			originRegion: 257,
			warnings: [],
			environment: {
				startTimeOfDay: .5,
				ratePerSecond: 0,
				tracks: {
					color0xf0: [ { t: 0, r: 1, g: 1, b: 1 } ],
					color0x124: [ { t: 0, r: .55, g: .55, b: .55 } ]
				}
			},
			groups: [
				{
					id: "ground",
					center: [ 0, 0, 0 ],
					radius: 300,
					material: lit( [ .72, .62, .5, 1 ] ),
					geometry: {
						world: true,
						positions: Float32Array.of( -120, 0, -120, 120, 0, -120, 120, 0, 120, -120, 0, 120 ),
						normals: Float32Array.of( 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0 ),
						uvs: Float32Array.of( 0, 0, 1, 0, 1, 1, 0, 1 ),
						indices: Uint32Array.of( 0, 1, 2, 0, 2, 3 ),
						instances: identity,
						transform: identity
					}
				},
				{
					id: "dome",
					center: [ 0, 5, 0 ],
					radius: 24,
					material: lit( [ .8, .45, .35, 1 ] ),
					geometry: {
						world: true,
						positions: new Float32Array( positions ),
						normals: new Float32Array( normals ),
						uvs: new Float32Array( positions.length / 3 * 2 ),
						indices: new Uint32Array( indices ),
						instances: identity,
						transform: identity
					}
				},
				{
					id: "box",
					center: [ 18, 4, 0 ],
					radius: 12,
					material: lit( [ .5, .55, .65, 1 ] ),
					geometry: {
						world: true,
						positions: new Float32Array( cube ),
						normals: new Float32Array( cubeNormals ),
						uvs: new Float32Array( cube.length / 3 * 2 ),
						indices: new Uint32Array( cubeIndices ),
						instances: identity,
						transform: identity
					}
				}
			]
		} );
		const rows = [];
		for ( const mode of [ "off", "hdrToneMap", "off", "sunShadow", "off", "perPixelLighting", "off" ] ) {
			renderer.experimentalVideo( { ...options, ...(mode === "off" ? {} : { [mode]: true }) } );
			for ( let frame = 0; frame < 5; frame++ ) await renderer.frame( { width, height }, .16 );
			if ( renderer.error() ) throw Error( `${mode}: ${renderer.error()}` );
			const image = await createImageBitmap( canvas );
			context.drawImage( image, 0, 0 );
			image.close();
			rows.push( { mode, rgba: [ ...context.getImageData( 0, 0, width, height ).data ] } );
		}
		return rows;
	} finally {
		renderer.dispose();
	}
}

test(
	"each lighting stage changes the native frame and disabling restores every pixel",
	{ timeout: 120000 },
	async () => {
		const { browser, page } = await launchProbeBrowser();
		const errors = [];
		page.on( "pageerror", error => errors.push( error.message ) );
		try {
			await page.goto( CLIENT_NEXT_BASE_URL );
			const rows = await page.evaluate( captureLightingStages );
			const native = rows[0].rgba;
			for ( const row of rows ) {
				if ( row.mode === "off" ) {
					assert.deepEqual( row.rgba, native, "Disabled stages must restore exact native pixels" );
				} else assert.notDeepEqual( row.rgba, native, `${row.mode} must independently affect the frame` );
			}
			// The witness must actually be lit geometry, not a black frame.
			assert.ok( native.some( ( value, i ) => i % 4 !== 3 && value > 24 ), "Witness must draw visible geometry" );
			assert.deepEqual( errors, [] );
		} finally {
			await browser.close();
		}
	}
);
