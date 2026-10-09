/*
===========================================================================

experimental-extended.test.mjs - the extended-content row is opt-in only

The live-2026 graft rides the Experimental window like every other
non-native addition: only an explicit true enables it, the window lists
it under its own tab, and the draft never touches saved preferences
before Confirm. Native (everything off) stays the default.

===========================================================================
*/
import "../helpers/native-source-loader.mjs";
import { test } from "node:test";
import assert from "node:assert/strict";
const { experimentalOptions } = await import( "../../src/engine/foundation/ui/experimental-options.ts" );
const { createExperimentalHud, EXPERIMENTAL_TABS } = await import(
	"../../src/engine/runtime/ui/hud/experimental-hud.ts"
);

test("extended content stays off unless explicitly true", () => {
	assert.equal( experimentalOptions().extendedContent, false );
	assert.equal( experimentalOptions( {} ).extendedContent, false );
	assert.equal( experimentalOptions( { extendedContent: "true" } ).extendedContent, false );
	assert.equal( experimentalOptions( { extendedContent: 1 } ).extendedContent, false );
	assert.equal( experimentalOptions( { extendedContent: true } ).extendedContent, true );
});

test("the experimental window lists the extended content row", () => {
	const rows = EXPERIMENTAL_TABS.flatMap( ( tab ) => tab.rows );
	const row = rows.find( ( candidate ) => candidate.key === "extendedContent" );
	assert.ok( row, "the window must expose the extended content row" );
	const ids = rows.map( ( candidate ) => candidate.id );
	assert.equal( new Set( ids ).size, ids.length, "row control ids must be unique" );
	const anyOptions = experimentalOptions( { extendedContent: true } );
	for ( const candidate of rows ) {
		assert.ok( candidate.key in anyOptions, `row key ${candidate.key} must be a real preference` );
	}
});

test("toggling extended content changes the draft, never the saved value", () => {
	const hud = createExperimentalHud();
	hud.open();
	assert.equal( hud.state().draft.extendedContent, false );
	hud.toggle( "extendedContent" );
	assert.equal( hud.state().draft.extendedContent, true );
	assert.equal( hud.state().saved.extendedContent, false, "confirm is what persists" );
	assert.equal( hud.confirm().extendedContent, true );
});
