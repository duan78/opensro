/*
===========================================================================

surface.ts - the swapchain owner: scene target and final frame pair

acquire() reserves the frame's swapchain texture and answers the scene's
color target: the offscreen intermediate when the frame retains one
(presentation pass, deferred queries, a scaled scene), else the swapchain
view itself. The canvas always runs at the full backing-store size; a
render scale below 100 shrinks only the scene's intermediate and depth,
which the presentation upscales - the HUD composes at full resolution on
the final view either way.
final() hands the reserved pair to the frame, which presents the scene onto
the texture and composes the HUD on the view - both inside the frame's own
command buffer, so a frame is one submit.

===========================================================================
*/
import type {
	SurfaceCommands,
	SurfaceOwner,
	DepthTarget,
	ColorTarget
} from "@/engine/runtime/renderer/internal/gpu-contract";
export function createSurface(
	canvas: HTMLCanvasElement,
	commands: SurfaceCommands,
	format: GPUTextureFormat
): SurfaceOwner {
	const context = canvas.getContext( "webgpu" );
	if ( !context ) throw new Error( "WebGPU canvas unavailable" );
	let depth: DepthTarget | null = null, colorTarget: ColorTarget | null = null;
	let width = 0, height = 0, sceneWidth = 0, sceneHeight = 0, disposed = false;
	let frame: { texture: GPUTexture; view: GPUTextureView; } | null = null;
	return {
		depth() {
			if ( disposed || !depth ) throw new Error( "No active depth surface" );
			return depth.view;
		},
		color() {
			if ( disposed || !colorTarget ) throw new Error( "No retained frame color" );
			return colorTarget;
		},
		final() {
			if ( disposed || !frame ) throw new Error( "No acquired frame" );
			return frame;
		},
		acquire( viewport, retain = false, scale = 1 ) {
			if ( disposed ) throw new Error( "Disposed surface" );
			const nextSceneWidth = Math.max( 1, Math.round( viewport.width * scale ) ),
				nextSceneHeight = Math.max( 1, Math.round( viewport.height * scale ) );
			if (
				width !== viewport.width || height !== viewport.height ||
				sceneWidth !== nextSceneWidth || sceneHeight !== nextSceneHeight
			) {
				colorTarget?.dispose();
				colorTarget = null;
				width = viewport.width;
				height = viewport.height;
				sceneWidth = nextSceneWidth;
				sceneHeight = nextSceneHeight;
				canvas.width = width;
				canvas.height = height;
				commands.configure( context, format );
				const replacement = commands.createDepth( sceneWidth, sceneHeight );
				depth?.dispose();
				depth = replacement;
			}
			const texture = context.getCurrentTexture();
			frame = { texture, view: texture.createView() };
			if ( retain ) {
				if ( !colorTarget ) colorTarget = commands.createColor( sceneWidth, sceneHeight );
				return colorTarget.view;
			}
			return frame.view;
		},
		dispose() {
			if ( disposed ) return;
			disposed = true;
			depth?.dispose();
			depth = null;
			colorTarget?.dispose();
			colorTarget = null;
			frame = null;
			context.unconfigure();
		}
	};
}
