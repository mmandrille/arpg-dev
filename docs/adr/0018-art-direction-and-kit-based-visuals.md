# ADR-0018: Art Direction and Kit-Based Visuals (KayKit)

- **Status:** Accepted (2026-09-28). P0 checklist resolved in v469 ([findings](../researchs/v469_kaykit-p0-findings.md)). One open item
  carries into P3: the rig of the itch.io Adventurers 2.0 / Character Animations 1.1 packs.
- **Date:** 2026-09-28
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** art-direction, assets, rendering, lighting, animation, dungeon-visuals, vfx
- **Extends:** ADR-0001 D1 (3D low-poly, flat-ish shading), ADR-0006 (glTF asset pipeline)
- **Preserves:** ADR-0007 (client-only, event-driven animation), ADR-0001 D2 (authoritative server)
- **Reverses:** the "no imported texture packs / shader addons / marketplace kits" constraints in
  `docs/specs/v461_spec-dungeon-surface-kit.md` and `docs/specs/v464_spec-surface-material-kit.md`

---

## Context

Around 460 slices in, the gameplay architecture is mature. The visuals still look like a prototype,
and nothing in the current plan changes that.

- **Heroes:** one CC0 base mesh (`assets/characters/base_human/base_human_mesh.glb`) is
  vertex-morphed into five classes (`tools/assets/class_body_morph.py`). It is auto-weighted to a
  17-bone translation-only skeleton (`tools/assets/canonical_skeleton.py`). Its albedo texture is
  2×2, so each class renders as one flat color tinted at runtime (`client/scripts/class_body_tint.gd`).
- **Animation:** a few Euler keyframes per bone are hand-typed in `client/tools/build_animations.gd`.
  The file header says the motion is "crude on purpose (art is a non-goal)".
- **Dungeon:** walls, columns, obstacles, doors, chests and town props are `BoxMesh` /
  `CylinderMesh` / `PlaneMesh` built in code. Their textures are 64×64 images painted per pixel in
  GDScript, filtered with `TEXTURE_FILTER_NEAREST` (`client/scripts/ground_wall_factory.gd`,
  `client/scripts/wall_renderer.gd`).
- **Monsters:** most are "user-provided" trimesh exports (up to ~39k triangles, 2048² textures)
  with license `unconfirmed-user-provided`. They sit next to 4k-triangle heroes. Three unused
  4.7–6.9 MB GLBs are committed and referenced by nothing.
- **Rendering:** one `DirectionalLight3D` with shadows disabled and a bare `WorldEnvironment`
  (`client/scripts/main.gd` `_build_scene`, commented "placeholder primitives"). There is no
  tonemap, SSAO, glow, fog or anti-aliasing. There are no particles and no 3D shaders. The renderer
  is the implicit engine default.
- **Visual regression (before v469):** nothing enforced what the game looks like.
  - `make regen-screenshots` had no recorded real run.
  - Its dry-run test polluted `.artifacts/` with empty directories.
  - Two focuses (`town`, `stairs`) had been saving blank frames as "ok" since `ba083f77`.
  - No suite captured a dungeon, the town, monsters or lighting.
  - v469 fixed all of this: 220/220 captures pass, GDScript errors fail a capture, and a `scenes`
    suite was added. See D9.

Recent visual slices (v308, v461–v464) kept polishing primitives because their specs ruled out
imported kits, texture packs and shaders. That constraint is the root cause. Procedural geometry and
per-pixel textures written by an agent have hit their quality ceiling. More slices under the same
rules will not produce a good-looking game.

The project needs one cohesive art source, a real lighting and post-processing baseline, and a
screenshot gate. The source must be permissively licensed, consistent in style, low-poly, and ship
with a rig and animation set.

---

## Decisions

### D1 — Art direction: one cohesive stylized low-poly family (KayKit)

The game's visual family is **KayKit** (Kay Lousberg), which is CC0 low-poly. It uses a shared
gradient-atlas texturing style and a common character rig. The packs in scope are:

| Pack | Replaces |
|------|----------|
| Adventurers | the five hero bodies |
| Character Animations (optional; 1.0 characters already embed 76 clips) | `build_animations.gd` character clips |
| Skeletons | undead monster families |
| Dungeon Remastered | walls, floors, columns, doors, chests, torches, barrels, dungeon props |

The class mapping, verified in v469:

| Class | Adventurers 1.0 (GitHub, staged) | Adventurers 2.0 (itch.io, owner download) |
|-------|----------------------------------|-------------------------------------------|
| paladin | Knight | Knight |
| barbarian | Barbarian | Barbarian |
| sorcerer | Mage | Mage |
| rogue | Rogue | Rogue |
| ranger | Rogue_Hooded | **Ranger** |

Adventurers 1.0 has no bow, so the ranger bow is a gap until 2.0 is verified. The 1.0 characters
embed 76 clips each. The core clip set therefore does not depend on the Character Animations pack;
that pack is an optional extension. Details: [findings](../researchs/v469_kaykit-p0-findings.md) §2–3.

**Coherence rule:** a new runtime model must come from the KayKit family, or be visibly
style-matched to it and approved in a spec. "Visibly style-matched" means a before/after capture
beside a KayKit reference, same flat/gradient shading, and a comparable triangle budget. Mixing art
families is the failure mode this ADR exists to stop.

**Mood:** the dark ARPG tone comes from lighting, fog, biome palettes and VFX (D7). It does not come
from switching to realistic assets.

**Rationale:**
- A shared rig and atlas remove most of the per-asset rigging work.
- The flat/gradient style matches ADR-0001 D1 ("limited palette, flat-ish shading").
- CC0 removes license risk.
- Coherence matters more than fidelity for a readable isometric ARPG.

### D2 — Sourcing, vendoring and provenance

**Sources the owner has authorized (2026-09-28):** any legally free resource made by artists is
allowed: 3D models, textures, sprites and particle sheets, fonts, sounds. KayKit (D1) stays the
primary visual family.

Secondary sources follow the D1 coherence rule. Examples:
- **Kenney, Quaternius:** models and particle sprites.
- **Poly Haven, ambientCG:** PBR textures, HDRIs.
- **OpenGameArt, itch.io free assets:** each asset needs its own license check.

**License allow-list:**

| License | Status |
|---------|--------|
| CC0 / public domain | Allowed |
| CC-BY 3.0 / 4.0 | Allowed only with an entry in the committed credits file |
| Other free licenses that permit commercial use and modification (e.g. MIT-style asset licenses) | Allowed only with an entry in the committed credits file |
| Non-commercial (NC) | **Rejected** |
| No-derivatives (ND) | **Rejected** |
| Share-alike on art assets (SA) | **Rejected** |
| "Free for personal use" | **Rejected** |
| Unclear or missing license | **Rejected** |
| AI-scraped redistributions and asset rips from commercial games | **Rejected** |

The credits file is `assets/CREDITS.md`, created by the first slice that imports a CC-BY asset.
It lists the asset, the author, a link to the license and the source URL.

**Vendoring and provenance**
- KayKit GLB/glTF files are committed under a KayKit subtree of `client/assets/`. P0 fixes the
  exact layout.
- Each runtime file gets an `assets/manifests/assets.v0.json` entry with provenance, following
  ADR-0006 D2: pack name, pack version, source URL, `license: CC0-1.0`, `sha256`.
- The pack's license text is committed next to the vendored files.
- Agents may download allow-listed resources from the original author's or distributor's official
  page. Before each download or batch, the agent states the filename, source and size in chat, and
  records the license evidence. There is no runtime or build-time fetch (ADR-0006 D1 still holds).
- **License hygiene:**
  - Runtime assets with `unconfirmed-user-provided` provenance are replaced or removed by the
    monster slice (P4).
  - Committed GLBs that nothing references are deleted.
  - After P4, no runtime asset may carry an unconfirmed license.

### D3 — What is now allowed

| Allowed | Still rejected unless amended |
|---------|-------------------------------|
| Imported modular kits and their textures (KayKit per D1) | Third-party Godot addons/plugins and shader packs |
| In-repo `.gdshader` files authored for this project | Runtime-downloaded art |
| `GPUParticles3D` / `CPUParticles3D` VFX | Mixing art families without a spec (D1) |
| Forward+ rendering features (D7) | |

Existing specs that rejected imported kits (v461, v464) are superseded on that point only.

### D4 — Rig and animation: adopt the kit rig; retire the canonical 17-bone skeleton

**Rig and clips**
- Heroes and humanoid monsters use the KayKit rig (41 joints; the deform chain is
  `root/hips/spine/chest/head` plus limb chains ending in `handslot.{l,r}` and `toes.{l,r}`).
  They use the clips embedded in each kit character file, and optionally Character Animations
  pack clips on the same rig. Everything is imported through Godot's import pipeline.
- **Retired** once the hero slice (P3) lands:
  - the canonical 17-bone skeleton (`canonical_skeleton.py`, `rig_canonical_hero.py`,
    `rig_hero_glbs.py`)
  - `class_body_morph.py`
  - `class_body_tint.gd`
  - the character clips in `build_animations.gd`

**Clip mapping**
- Logical animation states (`idle`, `walk`, the attack variants, `hit`, `death`, and future skill
  clips) map to kit clip names through a **presentation catalog** under `shared/assets/`.
  The catalog is data, not code; P3 names the file.
- `animation_controller.gd` keeps its priority state machine. It resolves clip names through the
  catalog instead of hardcoded names.
- ADR-0007 is unchanged: animation stays client-only and is driven by existing authoritative events.

**Sockets**
- `shared/assets/gear_sockets.v0.json` is remapped to kit bone names. Socket **names** are unchanged
  (ADR-0006 D4), so `item_visuals` consumers are unaffected.

**Validator**
- `validate_assets.py` drops the hardcoded `hand_r` / `hand_l` requirement.
- It replaces it with: every bone referenced by `gear_sockets` must be a real skin joint of each
  rigged character.

**First-person view**
- The v468 real-body first-person view (`chest_view`) is re-verified against the kit rig in P3.
  The camera anchor bone moves into the socket/camera catalog. It is not hardcoded.

### D5 — Equipment visuals: attached meshes for silhouette slots, tints for armor

The paper-doll model changes by slot:

| Slot(s) | World visual |
|---------|--------------|
| `main_hand`, `off_hand` | Attached mesh on a socket (KayKit weapons/shields) |
| `head` | Attached mesh on a socket (KayKit helmets/hats), or the character's own head accessory toggled |
| `chest`, `gloves`, `boots`, `belt` | **Tint**: a color/material change on the character's own mesh regions. No separate mesh. |
| `ring_left`, `ring_right`, `amulet` | No world visual (UI only) |

**Tint data**
- A tint is presentation data in `shared/assets/item_visuals` (schema-versioned per the
  shared-contract rules). Each entry names a **target region** and a color, and optionally a
  roughness/metallic hint.
- Each class declares a region → mesh-part (or region-mask) mapping in a presentation catalog.

**Tint mechanism (resolved in v469): per-mesh-part material override.**
- Kit characters are split into `<Class>_Body`, `_ArmLeft`, `_ArmRight`, `_LegLeft`, `_LegRight`
  and `_Head` meshes.
- A duplicated `StandardMaterial3D` with a tinted albedo is set as `material_override`. The v469
  probe proved it.
- No region-mask shader is needed.

**Granularity (resolved):**

| Slot | Tint target |
|------|-------------|
| chest | `_Body` |
| gloves | both **whole arms** |
| boots | both **whole legs** |
| belt | shares `_Body`, so there is no separate belt visual; P3 decides the precedence against chest |

The owner accepted this tradeoff. Weapons, shields and class headgear already exist as accessory
nodes under the `handslot.r`, `handslot.l` and `head` bones, so equipping one can toggle node
visibility. Details: [findings](../researchs/v469_kaykit-p0-findings.md) §4.

**Retired**
- Chest, glove, boot and belt GLB attachments and their per-class `local_transform` tuning.

### D6 — Dungeon rendering: grid-snapped rectangles, auto-tiled into kit pieces

**Protocol and authority**
- The protocol is unchanged: walls stay `{id, position, size, source}` rectangles in the session
  snapshot, and the server stays authoritative for collision and reachability.

**Client rendering** (as built in v471: per-wall tiling at kit scale 1.0 instead of a global
occupancy auto-tiler, because kit pieces are natively 4 tall × 1 deep, matching `ceiling_height`
and `wall_thickness`, and per-wall tiling keeps each wall's occlusion/pick identity. See the v471
spec.)
- The client rasterizes wall rectangles into a grid occupancy map at a presentation tile size set in
  a `shared/assets/` catalog.
- It auto-tiles that map into kit pieces (straight, outer corner, inner corner, T-junction, end cap,
  pillar) using `MultiMeshInstance3D` batches per piece type.
- Floors use kit floor tiles in the same grid.
- Doors, chests, torches, stairs, water edges, holes and obstacles map to kit props through
  presentation catalogs.

**Retired**
- The procedural pixel-painted surface textures.
- Box/cylinder wall and corner meshes (v308, v461–v464 presentation).

Client-side colliders and occlusion (`wall_occlusion_fade.gd`) keep deriving from the rectangles,
not from kit meshes.

**Server change**
- **Resolved in v469: no server change.**
  - The wall-grid audit covered 22,312 edges over 20 seeds × 10 levels. **100%** of wall edges lie
    on a 0.5 grid, and 91.6% on a 1.0 grid.
  - The kit modules are a 4-unit wall, a 2-unit half wall and a ~1-unit end cap.
  - So the environment renders at XZ scale **0.5**: half wall = 1.0, end cap ≈ 0.5, and the 1-deep
    wall is stretched to the server's 1.0 thickness.
  - P2 is client presentation only.
- A data-driven snap rule, with a regenerated `dungeon_obstacles.json` golden, is only reconsidered
  if P2 captures show seams.
- Proposed vertical scales, to be confirmed with P2/P3 captures: characters ×0.77 (keeps today's
  ~1.9 hero height), environment Y ×0.75.
- Details: [findings](../researchs/v469_kaykit-p0-findings.md) §5.

### D7 — Render baseline: Forward+, shadows, tonemap, AO, glow, fog, anti-aliasing

**Settings**
- `client/project.godot` declares the Forward+ renderer explicitly.
- The key light casts shadows.
- The environment enables filmic tonemapping (AgX or ACES), SSAO, glow for emissives, and
  depth/volumetric fog in dungeon biomes.
- MSAA or TAA is enabled.
- Kit textures use linear filtering with mipmaps.

**Data**
- Environment, light and fog parameters live in a presentation catalog under `shared/assets/`, per
  biome and per town day/night. This follows the existing `dungeon_depth_lighting.gd` / biome-palette
  pattern, not constants in `main.gd`.
- Torch lights may cast shadows under a per-quality-tier budget.

**Quality tiers**
- The v347 "Graphics quality" setting (Balanced / Performance) extends to toggle shadows, SSAO,
  volumetric fog and the AA mode.

**Fog of war**
- The canvas fog-of-war overlay is kept at first. P1 captures show whether real lighting and fog
  make parts of it redundant. Any removal is its own slice.

### D8 — Budgets, enforced by `validate_assets.py`

| Asset class | Triangle budget | Texture budget |
|-------------|-----------------|----------------|
| Hero / humanoid monster | ≤ 10k | shared kit atlas, or ≤ 1024² per asset |
| Large monster / boss | ≤ 20k | ≤ 1024² |
| Weapon / shield / helmet | ≤ 3k | shared kit atlas |
| Dungeon piece / prop | ≤ 4k (raised from 2k: measured kit max 3,499) | shared kit atlas |

- Triangle counts are read from GLB accessor data, and texture dimensions from embedded images.
- Budgets were measured in v469: kit characters are 4.6k–7.0k triangles including hidden accessory
  meshes, and every kit file uses one 1024² atlas. The first slice that imports kit files moves
  them into a data file; they do not go in validator code.

**Performance floor**
- At the Balanced tier, the v347 dungeon benchmark (24 monsters) must not regress below its
  recorded FPS floor.
- P1 re-baselines the floor with the new render settings and records it in its as-built.

### D9 — Verification: the screenshot harness is the visual gate

- Headless CI renders with `gl_compatibility`, so it cannot see Forward+ features.
- **P0 repairs `make regen-screenshots`**: dry-run must not write, the tool must be proven with a
  real full run, and a dungeon/town/monster scene suite is added.
- Every visual slice under this ADR records **before/after captures** in its as-built. Captures are
  taken with `showme` / `regen-screenshots` in a real render window, and the as-built attaches or
  links them.
- Headless checks still cover what they can: import validity, sockets and bones, catalog schemas,
  and budgets.

### D10 — Procedural generators become test-only

- `tools/assets/gen_glb.py` / `gen_glb_equipment.py` primitives remain only for deterministic test
  fixtures (for example `monster_dummy`) and explicit fallbacks.
- They are no longer a path to shipped visuals.
- `import_equipment_glb.py` (Poly Pizza fetch) is retired along with the Poly Pizza equipment.

---

## Rollout

Each row is its own slice under the SDD process. Rows marked "exempt" are client-presentation-only
and fall under the CLAUDE.md spec-gate exemption. They still write an as-built with captures.

| Phase | Slice | Gate |
|-------|-------|------|
| **P0** | Repair screenshot harness; "before" captures; download packs (per D2); verify P0 checklist; record measured budgets | spec ([v469](../specs/v469_spec-art-baseline-and-kit-verification.md)) |
| **P1** | Render baseline (D7) | exempt — shipped [v470](../as-built/v470_render-baseline.md) (perf floor owed) |
| **P2** | Dungeon auto-tile renderer + kit props (D6) | walls/columns/floors shipped [v471](../as-built/v471_dungeon-kit-walls-floors.md); torches/treasure chests [v473](../as-built/v473_dungeon-kit-props.md) (no kit door; stairs deferred) |
| **P3** | Heroes + kit animation + clip catalog + socket remap + armor tints (D4, D5) | spec — P3a heroes/clips/sockets/weapons shipped [v475](../as-built/v475_kit-heroes.md); P3b armor tints + headgear; P3c legacy pipeline removal |
| **P4** | Kit monsters; purge unconfirmed-license and unused GLBs (D2) | spec — P4a skeletons + purge shipped [v474](../as-built/v474_kit-skeleton-monsters.md); P4b beasts shipped [v476](../as-built/v476_cc0-beasts.md) (Quaternius CC0 wolf + bat, palette-baked) |
| **P5** | Particle VFX + in-repo shader set: hit flash, death dissolve, highlight rim/outline (D3) | exempt |

### P0 verification checklist (resolved in v469, see [findings](../researchs/v469_kaykit-p0-findings.md))

All six items are resolved. The carried-over rig item is **resolved (2026-09-29)**:
- Adventurers 2.0 and Character Animations 1.1 share `Rig_Medium`.
- `Rig_Medium` has 23 joints: the 1.0 rig's deform joints under the same names, minus the IK
  helpers.
- So the 1.1 clips drive both the 2.0 heroes and the v474 kit monsters.

See [kaykit-asset-inventory.md](../researchs/kaykit-asset-inventory.md).


1. Confirm each pack's license text is on the D2 allow-list (KayKit is expected to be CC0) and
   record the pack versions.
2. Confirm the Adventurers pack contents and the final class mapping (D1).
3. Confirm the Adventurers characters and the Character Animations clips share one rig. List the
   bone names used for the hand, head and chest sockets and the first-person camera anchor.
4. Determine how kit characters are split into mesh parts. This picks the D5 tint mechanism and the
   final per-slot tint granularity.
5. Measure the Dungeon Remastered tile size. Check whether server-generated wall rectangles snap to
   it across a sample of seeds and levels (D6 server-change branch).
6. Measure kit triangle and texture sizes and finalize the D8 budgets.

---

## Consequences

**Positive**
- A consistent, readable style. The rigging, animation and texturing work mostly disappears.
- Most phases touch only client presentation. The server, protocol, shared gameplay rules and replay
  determinism are unaffected, except for the optional D6 snap rule.
- License risk is removed from runtime assets.
- Visual quality gets an actual gate (D9).

**Negative / costs**
- Expect large test churn in client tests that assert primitive meshes, specific clip names or
  legacy socket bones. Those assertions change to catalog-derived or semantic checks, per the Test
  Locking Policy.
- P3 is the highest-risk slice: the rig swap, the socket remap, the first-person view and the tint
  shader all land together. Heroes may need a temporary dual path (legacy and kit) behind a
  presentation flag until parity is verified.
- Armor is less expressive: chest, glove, boot and belt items no longer change the silhouette.
- CI cannot detect lighting or post-processing regressions. That relies on D9 captures.
- The repo grows by the vendored kit files. Unused GLB deletion partly offsets this.

---

## Rejected alternatives

- **Keep improving procedural geometry and GDScript-painted textures.** Rejected: this has already
  been tried across many slices and hit its ceiling. Agents write geometry well, but they do not
  author good art that way.
- **AI 3D generators (Meshy, Tripo, Rodin) for heroes and monsters.** Rejected for rigged
  characters: inconsistent topology, heavy triangle counts, and one-off styles. The current
  monsters show this. They may be reconsidered later for **static** props only, under the D1
  coherence rule.
- **Mixed CC0 sources (Poly Pizza, Quaternius, Sketchfab, Kenney).** Rejected: that is how the
  current mismatched look happened.
- **Quaternius modular packs as the primary family.** A viable alternative with similar licensing.
  Rejected because it is less cohesive across characters, dungeon and animation than KayKit.
- **Commissioned or realistic dark-fantasy art.** Rejected for now on cost and pipeline weight. Mood
  comes from lighting (D7). This can be revisited if the project moves beyond a prototype.
- **Separate armor meshes per slot on kit characters.** Rejected by the owner in favor of tints
  (D5). Attached torso and limb armor on chunky stylized bodies tends to clip and look glued on.
