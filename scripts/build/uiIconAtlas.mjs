/*
===========================================================================

uiIconAtlas.mjs - the UI icon family's atlas step

Port-only. Wraps ui_icon_atlas.py (the pack itself) for the resource
build: the step runs after every producer has published the icon PNGs,
writes the sheet and its manifest beside them, and claims both in the
open publication so the claim audit keeps them. The per-icon files stay
published: a client without the manifest reads them unchanged.

===========================================================================
*/
import { execFile } from "node:child_process";
import { existsSync } from "node:fs";
import path from "node:path";
import { promisify } from "node:util";
import { CLIENT_PUBLIC_ROOT } from "../lib/generatedRoot.mjs";
import { claimPublicFile, isPublicationOpen } from "./shared/publicationLedger.mjs";

const run = promisify( execFile );

/*
================
buildUiIconAtlas

The atlas for the published icon family, skipped where the pinned Python
cannot run it (the fallback is the per-icon path the client already
reads).
================
*/
export async function buildUiIconAtlas( flags = {} ) {
	if ( process.platform !== "win32" ) {
		console.log( "ui icon atlas: skipped (the pinned Pillow environment is windows-x64)" );
		return { skipped: true };
	}
	const out = flags.out ? [ "--out", flags.out ] : [];
	const python = process.env.SRO_PYTHON ?? "python";
	await run( python, [ path.resolve( "scripts/build/ui_icon_atlas.py" ), ...out ], {
		cwd: path.resolve( "." )
	} );
	if ( !flags.out && isPublicationOpen() ) {
		for ( const name of [ "manifest.json", "sheet-0.png" ] ) {
			const target = path.join( CLIENT_PUBLIC_ROOT, "assets", "ui-atlas", name );
			if ( existsSync( target ) ) claimPublicFile( target );
		}
	}
	return { skipped: false };
}
