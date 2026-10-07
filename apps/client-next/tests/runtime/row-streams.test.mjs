/*
===========================================================================

row-streams.test.mjs - the per-batch instance streams and the portrait view

The batch's fading opacities and hit point lights are written into retained
scratches, once per batch instead of once per primitive. These pins hold
that contract across membership changes: a batch that grows, shrinks and
loses its lights must hand updateInstances exactly what fresh allocations
would, and the portrait target must return one stable view per slot and
follow the texture when the slot is replaced.
===========================================================================
*/
import "../helpers/native-source-loader.mjs";
import { pathToFileURL as sourceFileUrl } from "node:url";
import { test } from "node:test";
import assert from "node:assert/strict";

const { fillRowOpacities, fillRowPointLights } = await import(
	sourceFileUrl( "src/engine/runtime/renderer/characters/characters.ts" ).href
);
const { createUiResources } = await import( sourceFileUrl( "src/engine/runtime/renderer/device/ui.ts" ).href );

const actor = ( gid, opacity, pointLight ) => ({
	gid,
	opacity,
	pointLight,
	pose: { regionId: 0, x: gid * 10, y: 0, z: 0, yaw: 0 }
});
const light = ( ambient, diffuse ) => ({
	pose: { regionId: 0, x: 1, y: 2, z: 3 },
	attenuation: 0.5,
	ambient,
	diffuse
});

test("a batch's opacity scratch matches a fresh allocation through growth and shrink", () => {
	const batch = {};
	const rows = [ actor( 1, 1 ), actor( 2, .5 ), actor( 3, .25 ), actor( 4, .75 ) ];
	const opacity = a => a.opacity ?? 1;
	// Grow: 2 rows, then 4 (past the initial power-of-two), then shrink to 1.
	const two = fillRowOpacities( batch, rows.slice( 0, 2 ), opacity );
	assert.deepEqual( [ ...two ], [ 1, .5 ] );
	const four = fillRowOpacities( batch, rows, opacity );
	assert.deepEqual( [ ...four ], [ 1, .5, .25, .75 ] );
	const one = fillRowOpacities( batch, rows.slice( 0, 1 ), opacity );
	assert.deepEqual( [ ...one ], [ 1 ] );
	// The shrunk view never leaks the retired rows' values.
	assert.equal( one.length, 1 );
	// A fresh scratch would produce the same bytes at every size.
	assert.deepEqual( [ ...four ], rows.map( a => a.opacity ) );
	assert.equal( two.length, 2 );
});

test("the light scratch zeroes retired rows when the lights disappear", () => {
	const batch = {};
	const lit = [
		actor( 1, 1, light( [ .1, .2, .3 ], [ .4, .5, .6 ] ) ),
		actor( 2, 1, light( [ .7, .8, .9 ], [ .15, .25, .35 ] ) )
	];
	const both = fillRowPointLights( batch, lit, 0 );
	assert.equal( both.length, 24 );
	// Layout per row: placed position (x,y,z), attenuation, ambient rgb,
	// zero padding, diffuse rgb, zero padding - Float32 rounded.
	const f = v => Math.fround( v );
	assert.deepEqual( [ ...both.slice( 0, 12 ) ], [
		1,
		2,
		3,
		f( .5 ),
		f( .1 ),
		f( .2 ),
		f( .3 ),
		0,
		f( .4 ),
		f( .5 ),
		f( .6 ),
		0
	] );
	// Slots 7 and 11 stay zero padding.
	assert.equal( both[7], 0 );
	assert.equal( both[19], 0 );
	// The second row keeps its own light.
	assert.deepEqual(
		[ ...both.slice( 12, 24 ) ],
		[ 1, 2, 3, f( .5 ), f( .7 ), f( .8 ), f( .9 ), 0, f( .15 ), f( .25 ), f( .35 ), 0 ]
	);
	// Rows that lose their lights: the retired span reads zero, and a shorter
	// view never exposes stale bytes.
	const unlit = fillRowPointLights( batch, [ actor( 1, 1, undefined ) ], 0 );
	assert.equal( unlit.length, 12 );
	assert.deepEqual( [ ...unlit ], new Array( 12 ).fill( 0 ) );
});

test("the portrait target returns one view per slot and follows a replaced texture", async () => {
	const oldBuffer = globalThis.GPUBufferUsage,
		oldTexture = globalThis.GPUTextureUsage,
		oldStage = globalThis.GPUShaderStage;
	globalThis.GPUShaderStage = { VERTEX: 1, FRAGMENT: 2, COMPUTE: 4 };
	globalThis.GPUBufferUsage = { STORAGE: 1, COPY_DST: 2, UNIFORM: 4 };
	globalThis.GPUTextureUsage = { TEXTURE_BINDING: 1, COPY_DST: 2, RENDER_ATTACHMENT: 4 };
	let textureCounter = 0;
	let views = new Map();
	const device = {
		createBindGroupLayout: () => ({}),
		createPipelineLayout: () => ({}),
		createShaderModule: () => ({}),
		createRenderPipelineAsync: async () => ({ getBindGroupLayout: () => ({}) }),
		createBuffer: () => ({ destroy() {} }),
		createSampler: () => ({}),
		createTexture: () => {
			const texture = { id: ++textureCounter, destroy() {} };
			texture.createView = () => {
				if ( !views.has( texture ) ) views.set( texture, { texture } );
				return views.get( texture );
			};
			return texture;
		},
		createBindGroup: () => ({}),
		queue: { writeTexture() {}, writeBuffer() {}, copyExternalImageToTexture() {} }
	};
	const ui = createUiResources( device, "rgba8unorm" );
	await ui.ready;
	// The white fallback takes texture 1; the first portrait slot is 2.
	const first = ui.portraitTarget( "__p", 128, 128 );
	const again = ui.portraitTarget( "__p", 128, 128 );
	assert.equal( again, first, "the same slot returns the same view object" );
	// A resize replaces the slot: a new view bound to the new texture.
	const resized = ui.portraitTarget( "__p", 64, 64 );
	assert.notEqual( resized, first );
	assert.equal( resized.texture.id, 3 );
	// Returning to the original size replaces it again.
	const restored = ui.portraitTarget( "__p", 128, 128 );
	assert.notEqual( restored, resized );
	assert.equal( restored.texture.id, 4 );
	assert.notEqual( restored, first );
	globalThis.GPUBufferUsage = oldBuffer;
	globalThis.GPUTextureUsage = oldTexture;
	globalThis.GPUShaderStage = oldStage;
});
