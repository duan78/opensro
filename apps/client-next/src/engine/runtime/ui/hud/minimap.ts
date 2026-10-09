import type { AssetOwner } from "@/engine/contracts/assets";
import { minimapNpcPositions } from "@/engine/foundation/ui/minimap-catalog";
import { minimapDungeons, minimapArt } from "@/engine/foundation/ui/minimap-tiles";
import { experimentalOptions } from "@/engine/foundation/ui/experimental-options";
// Extended content (isro-live-2026), port-only, not v1.150-native: the art
// overlay the extended world lane publishes. The server announces the
// extended set by serving it; a native tree has none and the minimap stays
// exactly native (the charter's fallback).
const EXTENDED_MINIMAP_CATALOG_PATH = "/assets/data/extended-minimap.json";
// Catalog jobs share the UI lifetime and the existing asset scheduler.
export function createMinimapResources(
	assets: Pick<AssetOwner, "available" | "request" | "take" | "cancel">,
	base: string
) {
	const dungeon = createDungeonResources( assets, base );
	let state:
		| { kind: "idle"; }
		| { kind: "loading"; id: number; }
		| { kind: "ready"; positions: ReturnType<typeof minimapNpcPositions>; }
		| { kind: "failed"; error: string; }
		| { kind: "disposed"; } = { kind: "idle" };
	return {
		step( needed: boolean, worldVisible = false ) {
			const changed = dungeon.step( worldVisible );
			if ( state.kind === "idle" && needed && assets.available() > 0 ) {
				state = {
					kind: "loading",
					id: assets.request( new URL( "/assets/data/npcpos.json", base ).href, 4 << 20 )
				};
			}
			if ( state.kind === "loading" ) {
				const result = assets.take( state.id );
				if ( result ) {
					try {
						if ( result.kind !== "bytes" ) {
							throw Error(
								result.kind === "error" ? result.error : "Invalid minimap catalog response"
							);
						}
						state = {
							kind: "ready",
							positions: minimapNpcPositions(
								JSON.parse( new TextDecoder( "utf-8", { fatal: true } ).decode( result.buffer ) )
							)
						};
					} catch ( error ) {
						state = { kind: "failed", error: String( error ) };
					}
					return true;
				}
			}
			return changed;
		},
		positions: () => state.kind === "ready" ? state.positions : undefined,
		dungeons: dungeon.data,
		art: dungeon.art,
		error: () => state.kind === "failed" ? state.error : dungeon.error(),
		dispose() {
			dungeon.dispose();
			if ( state.kind === "loading" ) assets.cancel( state.id );
			state = { kind: "disposed" };
		}
	};
}
function createDungeonResources( assets: Pick<AssetOwner, "available" | "request" | "take" | "cancel">, base: string ) {
	let state: "idle" | "loading" | "overlay" | "ready" | "failed" | "disposed" = "idle",
		id = 0,
		overlayId = 0,
		data: ReturnType<typeof minimapDungeons> | undefined,
		error: string | null = null;
	let art: ReturnType<typeof minimapArt> | undefined;
	return {
		step( needed: boolean ) {
			if ( state === "idle" && needed && assets.available() > 0 ) {
				id = assets.request( new URL( "/assets/data/mission-dungeon-minimap.json", base ).href, 4 << 20 );
				state = "loading";
			}
			if ( state === "loading" ) {
				const result = assets.take( id );
				if ( result ) {
					try {
						if ( result.kind !== "bytes" ) {
							throw Error(
								result.kind === "error" ? result.error : "Invalid minimap response"
							);
						}
						const value = JSON.parse( new TextDecoder( "utf-8", { fatal: true } ).decode( result.buffer ) ),
							nextArt = minimapArt( value ),
							nextData = minimapDungeons( value );
						art = nextArt;
						data = nextData;
						if ( experimentalOptions().extendedContent && assets.available() > 0 ) {
							overlayId = assets.request( new URL( EXTENDED_MINIMAP_CATALOG_PATH, base ).href, 4 << 20 );
							state = "overlay";
						} else state = "ready";
					} catch ( e ) {
						error = String( e );
						state = "failed";
					}
					return true;
				}
			}
			if ( state === "overlay" ) {
				const result = assets.take( overlayId );
				if ( result ) {
					if ( result.kind === "bytes" ) {
						try {
							art = new Set( [
								...art ?? [],
								...minimapArt(
									JSON.parse( new TextDecoder( "utf-8", { fatal: true } ).decode( result.buffer ) )
								)
							] );
						} catch ( e ) {
							console.warn( "extendedContent: invalid extended minimap catalog ignored", e );
						}
					} else {console.warn(
							"extendedContent: no extended minimap catalog served; the minimap stays native"
						);}
					state = "ready";
					return true;
				}
			}
			return false;
		},
		data: () => data,
		art: () => art,
		error: () => error,
		dispose() {
			if ( state === "loading" ) assets.cancel( id );
			if ( state === "overlay" ) assets.cancel( overlayId );
			state = "disposed";
			data = undefined;
			art = undefined;
		}
	};
}
