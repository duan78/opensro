/*
===========================================================================

texture-variants.ts - re-encoded textures, for devices without BC and for BC

Port-only, not native. The retail textures are BC (DXT1/3/5) and the
client uploads them unchanged where the GPU samples BC. A device without
BC (iPhone, iPad, most Android) otherwise decodes every level to RGBA8:
four to eight times the memory, which is what ends a phone tab. With the
Experimental "Compressed textures" row on, such a device is served the
asset build's re-encoding of each texture in a format it does sample
(ASTC 4x4; ETC2 is defined, not built), published beside the original under
TEXTURE_VARIANT_ROOT so every path keeps one owning pack group. With the
Experimental "BC7 alpha textures" row on, a device that samples BC is
served the build's BC7 re-encoding of the retail BC2 alpha textures - the
same eight bits a texel without BC2's 4-bit alpha - from the same root.

The asset build (scripts) and the asset worker read the same rules here.

===========================================================================
*/

// Where a variant set is published: TEXTURE_VARIANT_ROOT + format + the
// original path below /assets.
export const TEXTURE_VARIANT_ROOT = "/assets/texture-variants/";

/*
================
TextureVariant

A published variant set. ASTC 4x4 is 8 bits a texel for every source; ETC2
keeps BC1's 4 bits for colour textures (RGB8A1) and 8 for alpha. BC7 is
8 bits a texel like the BC2 it replaces, and is served where BC itself is
sampled: the desktop and console sets, not the phone path.
================
*/
export type TextureVariant = "etc2" | "astc" | "bc7";

// The adapter features the choice reads (named as WebGPU names them).
export type TextureFeature = "texture-compression-bc" | "texture-compression-etc2" | "texture-compression-astc";

/*
================
textureVariantOrder

The sets the asset build publishes for devices WITHOUT BC, in the order
such a device is served them. Only ASTC 4x4 is built (scripts/build/textures):
it is what iPhone and iPad sample. ETC2 joins this list when its set is
built; a set listed here but not published would send such a device to
paths that do not exist. The BC7 set is not in this list: it answers to
the BC feature, not the phone path, and chooseTextureVariant serves it
from its own row.
================
*/
export function textureVariantOrder(): readonly TextureVariant[] {
	return [ "astc" ];
}

/*
================
textureVariantFeature

The WebGPU feature a variant set needs.
================
*/
export function textureVariantFeature( variant: TextureVariant ): TextureFeature {
	if ( variant === "etc2" ) return "texture-compression-etc2";
	return variant === "bc7" ? "texture-compression-bc" : "texture-compression-astc";
}

/*
================
chooseTextureVariant

The variant a device is served. With the BC7 row on, a device that samples
BC is served the BC7 set - it replaces the retail BC2 alpha textures at
the same bytes, and a source whose variant is absent falls back to its
original (the asset registry). Otherwise a BC device keeps the retail
textures, and the compressed switch is the phone path: off is native,
else the first variant whose feature it has.
================
*/
export function chooseTextureVariant(
	features: ReadonlySet<string>,
	enabled: boolean,
	bc7AlphaTextures = false
): TextureVariant | null {
	if ( bc7AlphaTextures && features.has( "texture-compression-bc" ) ) return "bc7";
	if ( !enabled || features.has( "texture-compression-bc" ) ) return null;
	return textureVariantOrder().find( variant => features.has( textureVariantFeature( variant ) ) ) ?? null;
}

/*
================
textureVariantPath

The published path of path's variant, or null for a path that has none
(only native .texture containers under /assets/ are re-encoded).
================
*/
export function textureVariantPath( path: string, variant: TextureVariant ): string | null {
	if ( !path.startsWith( "/assets/" ) || !path.endsWith( ".texture" ) || path.startsWith( TEXTURE_VARIANT_ROOT ) ) {
		return null;
	}
	return TEXTURE_VARIANT_ROOT + variant + path.slice( "/assets".length );
}
