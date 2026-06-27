# v355 Spec: Aura Soft Lights

Status: Approved for implementation  
Date: 2026-06-26  
Codename: `aura-soft-lights`  
Baseline: v354 `dungeon-torch-lights`

## Purpose

Replace in-world **aura** visual cues on heroes and monsters with a single soft `OmniLight3D` radius per
entity, colored per aura type. Auras are area or self-buff presentation states driven by existing
server-owned `effect_ids` and skill events — not single-target debuffs.

Hero auras (`rage`, `holy_shield`, `sanctuary`) and monster command auras (`elite_command` on
followers, elite-leader radius preview when followers are buffed) lose their emissive ring/dome/crown
mesh markers and instead emit a data-driven colored point light whose range is derived from shared
rules (skill effect radius, sanctuary/holy-shield helpers, `dungeon_generation.monster_placement.elite_aura.radius`).

Holy Shield cast feedback becomes a brief light energy/range tween on the caster — no disc mesh pulse.

Debuff/status paints (burning, rogue mark, pinning root, stun, poison tint, ice slow, etc.) are
unchanged.

## Non-goals

- No server, protocol, `effect_ids`, or golden changes.
- No replacement of debuff/status markers or entity tint painting.
- No HUD status-bar, skill-icon, or rage visual-scale changes.
- No new aura types, aura roll tables, combat tuning, or minimap/nameplate/tooltip work.
- No production authored VFX meshes, particles, shaders, or external plugins.
- No Settings quality toggle for aura lights.
- No full forward_plus perf re-baseline (optional manual follow-up).

## Aura inventory (in scope)

| Kind | Entity | Trigger | Replaces |
|------|--------|---------|----------|
| Self buff | Hero | `rage` active on local player / `effect_ids` on remote players | `RageVisualEffect` ring + flame meshes |
| Self buff | Hero | `holy_shield` in `effect_ids` | `HolyShieldEffect` ring + column meshes |
| Area buff | Hero | `sanctuary` in `effect_ids` | `SanctuaryDomeEffect` dome + floor meshes |
| Command aura | Monster follower | `elite_command` in `effect_ids` | `EliteCommandVisualEffect` ring + crown meshes |
| Command aura preview | Monster leader | same-pack follower has `elite_command` | `EliteCommandRadiusPreview` torus ring |

## Aura selection rule

At most **one aura light** per entity at a time. When multiple aura sources could apply, use this
fixed priority (highest wins):

1. `sanctuary` (largest area buff)
2. `holy_shield`
3. `rage`
4. `elite_command_radius_preview` (leaders only)
5. `elite_command` (followers only)

Implementation may encode priority in the shared presentation catalog rather than hardcoding in
GDScript.

## Client asset / plugin decision

| Option | Decision |
|--------|----------|
| In-repo `OmniLight3D` + JSON presentation (hero fog light, dungeon torches) | **Adopt** |
| Emissive primitive aura meshes (`player_status_effect_markers.gd` factories) | **Replace** for auras only |
| External VFX plugins or authored mesh/particle packs | **Reject** |

## Acceptance criteria

- [ ] `shared/assets/aura_light_presentation.v0.json` (+ schema) defines per-aura color, energy,
  attenuation, `height_offset`, `shadow_enabled` (default false), and optional `hero` / `monster`
  multiplier blocks.
- [ ] Loader follows the existing `class_name` + `ensure_loaded()` singleton pattern.
- [ ] Each in-scope aura renders **only** an `OmniLight3D` child — no leftover aura mesh markers
  (`RageVisualEffect`, `HolyShieldEffect`, `SanctuaryDomeEffect`, `EliteCommandVisualEffect`,
  `EliteCommandRadiusPreview`).
- [ ] Debuff markers (`BurningVisualEffect`, `PinningRootVisualEffect`, `StunStarsVisualEffect`,
  `RogueMarkSkullEffect`, poison tint, etc.) are untouched.
- [ ] Light `omni_range` is rule-derived:
  - `sanctuary` / `holy_shield` cast pulse: skill `effect.radius` via existing client helpers
  - `rage`: personal radius from presentation catalog (no gameplay AoE today)
  - `elite_command` follower: presentation catalog personal radius
  - leader preview: `dungeon_generation.monster_placement.elite_aura.radius`
- [ ] Exactly one aura light per entity; priority table applied when multiple aura states overlap.
- [ ] Holy Shield ally cast uses a brief light tween (energy and/or range); no `HolyShieldAuraPulse`
  or `HolyShieldTargetPulse` disc meshes.
- [ ] Works for local player, remote player entities, and monsters.
- [ ] `elite_aura_preview_sync.gd` drives leader preview via aura light sync, not torus mesh.
- [ ] Perspective fog culling path updated for aura lights (reuse or simplify
  `elite_aura_preview_sync` LOS/distance culling; emissive aura meshes no longer bypass lighting).
- [ ] Bot presentation debug exposes aura-light presence and approximate range/color per entity
  (existing keys may be repurposed or aliased, e.g. `has_rage_effect` → aura light active for rage).
- [ ] `make validate-shared`, `make client-unit`, and targeted bot scenarios green.
- [ ] `player_status_effect_markers.gd` does not exceed maintainability baseline; extract
  `aura_soft_lights.gd` (or similar) if needed.

## Scope and likely files

| Area | Paths |
|------|--------|
| Shared data | `shared/assets/aura_light_presentation.v0.json`, `shared/assets/aura_light_presentation.v0.schema.json` |
| Client loader | `client/scripts/aura_light_presentation_loader.gd` (new) |
| Aura lights | `client/scripts/aura_soft_lights.gd` (new, preferred) or refactored `player_status_effect_markers.gd` |
| Elite preview | `client/scripts/elite_aura_preview_sync.gd` |
| Coordinator | `client/scripts/main.gd` (holy-shield pulse hook, sync call sites) |
| Bot debug | `client/scripts/bot_presentation_debug.gd` |
| Tests | `client/tests/test_status_effect_presentation.gd` (aura cases only) |
| Bot scenarios | `tools/bot/scenarios/client/34_elite_aura_readability.json`, `37_elite_aura_radius_preview.json`; optional extended `88_aura_soft_lights.json` |
| Docs | plan + as-built on completion |

**Reuse patterns:** `client/scripts/dungeon_torch_lights.gd`, `client/scripts/fog_of_war_overlay.gd`
(hero `OmniLight3D` tuning).

## Test and bot proof

**Unit (update aura tests only):**

```bash
make validate-shared
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_status_effect_presentation.gd
make client-unit
```

Assert `OmniLight3D` presence, aura type, and semantic range — not emissive mesh child names.

**Bot (extend existing scenarios):**

```bash
HEADLESS=1 make bot-client SCENARIO=34_elite_aura_readability
HEADLESS=1 make bot-client SCENARIO=37_elite_aura_radius_preview
```

Optional extended proof (default `"ci_tier": "extended"`):

```bash
HEADLESS=1 make bot-client SCENARIO=88_aura_soft_lights
```

**Visual verification:**

```bash
make bot-visual scenario=34_elite_aura_readability
make bot-visual scenario=37_elite_aura_radius_preview
```

**Gate:**

```bash
make maintainability
make ci
```

## Open questions and risks

| # | Item | Resolution |
|---|------|------------|
| Q-1 | Replace debuff meshes too? | **No.** Auras only; debuffs keep current mesh/tint cues. |
| Q-2 | Multiple auras on one entity? | **One light max;** fixed priority table above. |
| Q-3 | Effect-light shadows? | **Off** (`shadow_enabled: false`). |
| Q-4 | Holy Shield cast pulse? | **Light tween;** no mesh pulse. |
| Q-5 | Hero vs monster tuning? | **Single catalog** with optional hero/monster multiplier blocks. |
| R-1 | Many concurrent aura lights in large packs | Forward+ (v348) helps; cap is one light per entity. Monitor in manual play. |
| R-2 | `player_status_effect_markers.gd` size | Extract aura-light module; debuff factories stay in place. |
| R-3 | Headless CI uses `gl_compatibility` | Bot assertions must use debug state, not pixel/lighting screenshots. |

## ADR alignment

- **ADR-0001 / ADR-0007:** Client-only presentation; server `effect_ids` and events unchanged.
- **ADR-0006:** Adopt in-repo `OmniLight3D` + schema-backed JSON; reject external VFX.
- **v331 as-built:** Addresses emissive aura markers ignoring fog/lighting in perspective mode.
- **v354 as-built:** Reuses established torch/hero local-light presentation patterns.
