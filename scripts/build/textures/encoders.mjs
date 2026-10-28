/*
===========================================================================

encoders.mjs - the pinned texture encoders of the asset build

Port-only. The texture variant passes (buildTextureVariants.mjs) re-encode
the retail BC textures: ASTC for devices without BC, BC7 for the BC2 alpha
textures where BC is sampled (client foundation/assets/texture-variants.ts).
Their encoders are third-party tools fetched from their official GitHub
releases at the exact versions and SHA-256 digests below, kept in the hub's
.tools/ (shared by every worktree, never shipped to players) and used only
at build time.

- astcenc 5.7.0, Arm Limited, Apache-2.0
  (https://github.com/ARM-software/astc-encoder). The digest is the one
  Arm publishes in the release's release-sha256.txt. The release archive
  carries no licence file; licences/astcenc-5.7.0-LICENSE.txt is the
  LICENSE.txt of the 5.7.0 tag.

- texconv may2026, Microsoft, MIT
  (https://github.com/microsoft/DirectXTex). The digest is the one the
  release publishes as the asset's sha256 on GitHub. The executable is a
  standalone release asset, not an archive; licences/texconv-may2026-
  LICENSE.txt is the LICENSE of the tag. The BC7 pass forces its CPU codec
  (-nogpu): the DirectCompute codec a driver offers is not reproducible
  across machines, and a pinned build must be.

===========================================================================
*/
import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { existsSync, mkdirSync, renameSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import { MAIN_CHECKOUT_ROOT } from "../../lib/generatedRoot.mjs";

/*
================
ENCODERS

Each pinned tool: the release archive, its digest, and the executable
inside it.
================
*/
export const ENCODERS = Object.freeze( {
	astcenc: Object.freeze( {
		version: "5.7.0",
		licence: "Apache-2.0",
		url: "https://github.com/ARM-software/astc-encoder/releases/download/5.7.0/astcenc-5.7.0-windows-x64.zip",
		sha256: "cbde7c78adcc90d0d90ecf10f32d03f3340fbe1720cb6d6c892147d625e74d2d",
		executable: "bin/astcenc-avx2.exe"
	} ),
	texconv: Object.freeze( {
		version: "may2026",
		licence: "MIT",
		url: "https://github.com/microsoft/DirectXTex/releases/download/may2026/texconv.exe",
		sha256: "dcfdec10244e02cf5037fba089c55fb7e1326b1c8181742d77d15fa5cb5eef06",
		executable: "texconv.exe"
	} )
} );

/*
================
toolsRoot
================
*/
export function toolsRoot( mainCheckout = MAIN_CHECKOUT_ROOT ) {
	return path.join( mainCheckout, ".tools" );
}

/*
================
encoderPath

The installed executable of a pinned encoder, fetching and verifying its
release first when it is not installed yet. A pin whose archive is the
executable itself (texconv ships standalone .exe release assets) installs
by digest check alone; an archive pin extracts beside the final folder and
renames it into place, so a broken extraction never leaves a half-installed
tool behind.
================
*/
export async function encoderPath( name, root = toolsRoot() ) {
	const pin = ENCODERS[name];
	if ( !pin ) throw new Error( `Unknown encoder ${name}` );
	if ( process.platform !== "win32" ) throw new Error( `${name}: only the windows-x64 release is pinned` );
	const home = path.join( root, `${name}-${pin.version}` ), executable = path.join( home, pin.executable );
	if ( existsSync( executable ) ) return executable;
	const response = await fetch( pin.url );
	if ( !response.ok ) throw new Error( `${name}: download failed (${response.status})` );
	const archive = Buffer.from( await response.arrayBuffer() );
	const digest = createHash( "sha256" ).update( archive ).digest( "hex" );
	if ( digest !== pin.sha256 ) throw new Error( `${name}: SHA-256 ${digest} is not the pinned ${pin.sha256}` );
	if ( path.extname( pin.url ).toLowerCase() === ".exe" ) {
		mkdirSync( home, { recursive: true } );
		writeFileSync( executable, archive );
		writeFileSync(
			path.join( home, "PINNED.json" ),
			JSON.stringify( { name, ...pin, installedAt: new Date().toISOString() }, null, 2 ) + "\n"
		);
		return executable;
	}
	// Extract beside the final folder and rename it into place, so a broken
	// extraction never leaves a half-installed tool behind.
	const staging = home + ".partial", zip = home + ".zip";
	rmSync( staging, { recursive: true, force: true } );
	mkdirSync( staging, { recursive: true } );
	writeFileSync( zip, archive );
	// Windows' own tar (bsdtar) reads zip archives; a Git or MSYS tar first
	// on PATH does not.
	execFileSync( path.join( process.env.SystemRoot ?? "C:\\Windows", "System32", "tar.exe" ), [
		"-xf",
		zip,
		"-C",
		staging
	] );
	rmSync( zip, { force: true } );
	writeFileSync(
		path.join( staging, "PINNED.json" ),
		JSON.stringify( { name, ...pin, installedAt: new Date().toISOString() }, null, 2 ) + "\n"
	);
	if ( !existsSync( path.join( staging, pin.executable ) ) ) {
		throw new Error( `${name}: ${pin.executable} is not in the release archive` );
	}
	renameSync( staging, home );
	return executable;
}
