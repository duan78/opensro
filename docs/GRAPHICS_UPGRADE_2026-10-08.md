# Graphics modernization - 2026-10-08

Three stages of the renderer roadmap's third track (issue #273): the HDR
intermediate with a filmic tone map, one sun shadow cascade, and per-pixel
character and object lighting. The renderer stays a faithful 2005 D3D9
reconstruction; every stage here is a deliberate, documented deviation,
tuned through the named constants in its owning file.

**Every stage is an opt-in under Experimental, off by default.** Off is
the native 2005 frame, byte for byte: the same pipelines (no rgba16float
sibling ever compiles), the same pass order, and `env.stages2` zeroed in
the shader. The window's Image tab gains its fourth row (HDR tone map) and
a new Lighting tab holds the two direct-light stages; World keeps its four
rows, Chat and Developer shift one tab later.

The environment block's port-only tail grows by the second stages vec4
(byte 368: x HDR, y sun shadow, z per-pixel lighting) and the cascade
matrix (byte 384); the layout table lives in
`apps/client-next/src/engine/runtime/renderer/device/environment-block.ts`.

## 1. HDR intermediate and filmic tone map

`apps/client-next/src/engine/runtime/renderer/device/hdr.ts`,
`pipelines.ts`, `bloom.ts`, `geometry.ts`, `device.ts`, `frame.ts`

Experimental > Image > HDR tone map renders the scene into an
`rgba16float` intermediate instead of the 8-bit frame, and one fullscreen
pass resolves it through Narkowicz's fitted ACES curve (2015) into the
presented frame. The 2005 frame is authored display-referred LDR, so the
only values above 1.0 are the stages that lift them there - here the sun
disc gain (`SUN_HDR_GAIN`, pipelines.ts; the retail clamp stands when the
stage is off) - and there is no linear-to-sRGB re-encode on output; the
curve's job is the highlight roll-off and the filmic mid lift, and
`TONEMAP_EXPOSURE` is the whole compensation surface.

The structural cost is honest: the scene's geometry pipelines need
rgba16float siblings, and they compile lazily on the stage's first enabled
use (the float bloom rule - a native-only device never compiles them).
`geometry.sceneBindings` walks every live draw to its float pipeline
sibling and rebinds it (the `textureOptions` walk's pattern), the water
mirror and the portrait targets reformat to follow, and the silhouette
receivers build their float variant on demand. The UI never touches the
intermediate: with bloom on, the chain's copy-back step becomes the tone
map (`bloom.ts`'s lazily compiled `original` sibling, so the glow still
composites after the roll-off); without bloom, `hdr.ts`'s own pass runs
between the scene and the lens flares. The character preview composes
into the intermediate ahead of the tone map, so preview characters roll
off with the frame.

## 2. One sun shadow cascade

`apps/client-next/src/engine/runtime/renderer/device/sun-shadow.ts`,
`sun-shadow-math.ts` (foundation), `pipelines.ts`, `geometry.ts`,
`renderer.ts`

Experimental > Lighting > Sun shadows renders one depth-only pass of the
frame's opaque casters - world objects and characters; terrain, water and
everything blended stays out - from the shading light's position, and the
world shader resolves receivers with a 3x3 kernel of hardware comparison
taps. The cascade follows the same light the shader shades with (the
pinned retail diagonal, or the packed arc direction under Moving
sunlight), covers `SHADOW_EXTENT` metres around the eye at `SHADOW_SIZE`
taps, and its centre snaps to whole cascade texels (`sun-shadow-math.ts`)
so a walking camera does not shimmer the edges. The caster pass carries
the world shader's D3D9 alpha test (12-bit truncation, compare byte), so
leaf cards cast leaves, with `SUN_SHADOW_BIAS` slope-scaled bias against
acne.

Receivers keep their ambient: the cascade shades the diffuse term of the
world path, the object path and the lightmap terrain alike (the baked sum
is terrain's diffuse), fading to unshadowed at the cascade border
(`SUN_SHADOW_FADE`). Water and sky never receive. The native per-character
silhouette projections stand untouched beside it - this complements, it
does not replace. Every world draw binds the cascade at bindings 13/14;
while the stage is off those name a 1x1 dummy depth view no gated branch
ever samples, and the whole owner (shader, pipelines, cascade texture)
builds only on first enabled use.

Known follow-up (this wave's scope): the caster list is the frame's opaque
draws without distance culling; a city scene pays for every object again
in the depth pass.

## 3. Per-pixel character and object lighting

`apps/client-next/src/engine/runtime/renderer/device/pipelines.ts`

Experimental > Lighting > Per-pixel character light. The retail vs_1_1
path saturates each vertex's lighting before interpolation (`oD0`,
A5D450's diagonal in object space), so a 2005 character shows its
triangles under strong light. The stage re-evaluates the same terms - the
same diffuse and ambient factors, the same light - per pixel from the
interpolated, skinned world normal the shader already carries. Nothing
else changes: the vertex point-light path
(`NATIVE_CHARACTER_LIGHTING`, a compatibility mode of its own, off) stays
per-vertex.

## Off-path proof

The full runtime and architecture suites' failure set on this head is
byte-identical to naked main `2bb8c88b` on this machine (the licensed-data
set; the retail extraction is absent). In the browser
(`tests/browser/hdr-stages.test.mjs`), a synthetic lit scene - ground,
smooth-normal dome, casting box - cycles each stage off-on-off through the
production renderer: every enabled stage changes the frame, every disabled
capture restores the native pixels exactly. The shipped stages'
`environment-stages` A-B-A and the `game-options` Experimental window
round-trip (new rows, new tab order) pass unchanged.
