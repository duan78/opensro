# Graphics modernization - 2026-10-06, wave two

Four more opt-in stages under Experimental > Video, plus the heightfield
normals that feed one of them. The renderer stays a faithful 2005 D3D9
reconstruction: off is the native frame (the pinned 45-degree light, the
flat NOLIGHT ground, the fog band at the horizon, the quantized one-level
bloom), and each stage here is a deliberate, documented deviation with one
named constant and one revert line in its owning file. Wave one (the
presentation pass, anisotropy, height fog, water fresnel, garment sheen,
authored block textures) lives in `GRAPHICS_UPGRADE_2026-10-05.md`; the
upstream roadmap is tracked in opensro-dev/opensro#273.

`env.stages` grew to four switches: x height fog, y sun direction,
z terrain relief, w textured horizon. The environment block grew one vec4
(`Environment.sunDirection`, packed by `world-environment.ts`) ahead of
the stages; nothing else about the uniform layout moved.

## 1. Sun direction

`pipelines.ts` (`lightDirection()`), `world-environment.ts` (packing)

Retail shades every surface with one pinned diagonal light
(`dot(normal, vec3(0.707, 0.707, 0))`, both the vs_1_1 oD0 vertex path and
the per-pixel world path) while only its colour animates through the day.
The stage replaces the direction with the same X-Y arc the sky's sun quad
rides (`skyTime.x`, native 8cbe10), so dawn and dusk rake across geometry
and noon lights from above. Below the horizon the anti-solar point takes
over - the night's light arrives from where the moon sits, and the flip
lands exactly at horizon-crossing where the diffuse term is already zero,
so no frame pops. The vertex and pixel paths share one `lightDirection()`
helper; deleting it (or clearing stage y) restores the pinned diagonal.

## 2. Terrain relief

`terrain-associations.ts` (`terrainBlockNormals`), the asset worker's
terrain emission, `pipelines.ts` (the relief block)

Retail terrain is NOLIGHT: flat albedo that the lightmap then multiplies,
with authored normals of (0,1,0) that nothing reads. Two changes:

- The worker now computes one normal per 17x17 height sample (central
  differences over the 20-unit cells, one-sided at the borders, shared by
  every association pass and detail level) and writes it into the terrain
  vertex stream. This is data, not a look change: with the stage off the
  values are unused and the frame is byte-identical.
- The stage shades slopes against `lightDirection()` through those
  normals, levelled against the flat-ground term
  (`ambient + diffuse * 0.707`, the retail diagonal's ground response) so
  level terrain keeps its exact retail brightness and only slopes move.
  `TERRAIN_RELIEF` (0.45) is the whole tuning surface; 0 restores flat
  ground.

The lightmap still multiplies on top unchanged: relief modulates albedo,
the lightmap stays the baked lighting it always was.

## 3. Textured horizon

`pipelines.ts` (the `distant` flag)

Retail discards the lightmap and flat-fogs terrain past the detail band
(8ABFD0): 2005 minification aliased out there. Every terrain texture now
carries its authored-plus-generated mip chain (the block-container work),
so the band's texels are stable - the stage simply keeps the textured fog
blend (`mix(lit, terrainFogColor, fog)`) instead of the flat return. The
cost is the texture fetches those fragments were already skipping; the
seam behaviour is unchanged because both fog targets still agree. With
height fog on, distant peaks now keep their texture as they rise out of
the haze. Revert: drop `&& env.stages.w < 0.5` from the flag.

## 4. High dynamic glow

`bloom.ts`

The native glow chain is a byte-exact 8A99B0 replica, including its
per-tap 8-bit re-quantization and its single 512 blur level - which bands
on wide gradients. The float stage keeps every authored constant (capture
scale 128/255, threshold 40/255, kernel 80/70/50, composite alpha 198,
blend byte 192) but carries them through `rgba16float` targets and adds a
second 256 ping-pong level, so large glows keep their gradient instead of
stepping. The capture target stays in the canvas format (the main pass
renders into it); only the blur levels are float. Turning the stage off
hands quality straight back to the native chain with no reallocation
beyond the targets' own rebuild.

Soft particles were evaluated and deferred: they need the scene depth
sampled mid-frame, which means splitting the main render pass - a frame
order change through `execution-contract.json`, its own wave.

## Verification

- `pnpm --filter @sro/client-next run verify:quick` (typecheck, ownership,
  capabilities, execution map): PASS.
- `tests/runtime/world-environment.test.mjs` pins the grown block (88
  floats) and the arc: dawn east, noon zenith, dusk west, midnight
  anti-solar. `tests/runtime/terrain-normals.test.mjs` pins the normal
  table: flat stays (0,1,0), ramps tilt by gradient, borders use their
  actual span, everything unit length.
- `tests/browser/environment-stages.test.mjs` (against a local dev
  server): the native frame is byte-stable across repeat submits; the
  three world stages together move the Constantinople dock frame by more
  than 500 pixels of the 400x300 probe; the float bloom chain composites
  (dark stays 0, mid-gray gains glow, bright saturates), survives resize
  and hands quality back without a device error.
- The device compiles the modified geometry uber-shader and both bloom
  chains before its running phase, so every booting browser test also
  proves the WGSL.
