# v355 Plan — Aura Soft Lights

Status: Ready for implementation  
Goal: Replace aura mesh markers on heroes and monsters with one data-driven soft `OmniLight3D` per entity.  
Architecture: Extract `aura_soft_lights.gd` and a shared presentation catalog. `AuraSoftLights` owns a single
named `OmniLight3D` child per entity, resolves the winning aura via catalog priority, and applies
rule-derived range plus hero/monster tuning multipliers. Debuff factories stay in
`player_status_effect_markers.gd`. Holy Shield cast feedback becomes a brief light tween instead of
disc meshes. Bot proof reuses scenarios 34/37 debug keys backed by light presence checks.  
Tech stack: shared JSON + schema, Godot client (`OmniLight3D`), Python/Godot client bot, docs only.

Spec: [`docs/specs/v355_spec-aura-soft-lights.md`](../specs/v355_spec-aura-soft-lights.md)  
Baseline: v354 `dungeon-torch-lights`

## Spec review (gate)

| Area | Result |
|------|--------|
| Baseline v355 / builds on v354 | OK |
| Scope / non-goals | OK — auras only; debuffs unchanged |
| Contracts | Client-only shared asset; no protocol/golden |
| Determinism | No Go changes |
| Server authority | Presentation only |
| Bot proof | Extend 34/37; optional extended 89 |
| Client assets | Adopt in-repo `OmniLight3D`; reject external VFX |
| Maintainability | Extract new module; shrink `player_status_effect_markers.gd` |
| Open questions | All resolved in spec |

## Baseline and shortcut decision

Reuse patterns from:

- `client/scripts/dungeon_torch_lights.gd` — `OmniLight3D` creation/tuning
- `client/scripts/fog_of_war_overlay.gd` — hero point light defaults
- `client/scripts/dungeon_torch_presentation_loader.gd` — `class_name` + `ensure_loaded()` loader
- `client/scripts/elite_aura_preview_sync.gd` — leader preview orchestration + perspective culling

| Option | Decision |
|--------|----------|
| In-repo `OmniLight3D` + `aura_light_presentation.v0.json` | **Adopt** |
| Emissive aura mesh markers | **Replace** (auras only) |
| External VFX plugins | **Reject** |

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/assets/aura_light_presentation.v0.json` | Per-aura color, energy, attenuation, height, radius source, priority, hero/monster multipliers |
| Create | `shared/assets/aura_light_presentation.v0.schema.json` | Schema validation |
| Create | `client/scripts/aura_light_presentation_loader.gd` | Static loader singleton |
| Create | `client/scripts/aura_soft_lights.gd` | Single-light sync, priority resolution, holy-shield cast tween, debug helpers |
| Modify | `client/scripts/player_status_effect_markers.gd` | Remove aura mesh factories; keep debuff markers + effect-id constants; delegate aura `has_*` to `AuraSoftLights` |
| Modify | `client/scripts/elite_aura_preview_sync.gd` | Drive leader preview through `AuraSoftLights`; update culling marker list |
| Modify | `client/scripts/main.gd` | Route aura sync/pulse through `AuraSoftLights`; thin call-site replacements only |
| Modify | `client/scripts/bot_presentation_debug.gd` | Aura-light debug fields (`active_aura_id`, `aura_light_range`, optional color token) |
| Modify | `client/scripts/bot_scenario_runner.gd` | Optional assertions for new debug keys if scenario 89 added |
| Modify | `client/tests/test_status_effect_presentation.gd` | Aura cases assert lights/tweens, not mesh children; debuff tests unchanged |
| Create | `client/tests/test_aura_soft_lights.gd` | Focused unit tests for priority + light config (import module directly) |
| Modify | `scripts/client_smoke.sh` | Register `test_aura_soft_lights.gd` |
| Modify | `tools/bot/scenarios/client/34_elite_aura_readability.json` | Assert `active_aura_id == elite_command` (or keep `has_elite_command_effect`) |
| Modify | `tools/bot/scenarios/client/37_elite_aura_radius_preview.json` | Assert leader `active_aura_id == elite_command_radius_preview` + semantic radius |
| Create | `tools/bot/scenarios/client/89_aura_soft_lights.json` | Extended hero rage + holy shield proof (`"ci_tier": "extended"`) |
| Create | `docs/as-built/v355_aura-soft-lights.md` | Shipped summary |
| Modify | `PROGRESS.md` + `docs/progress/slice-lifecycle.md` | On `/finish` |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [ ] `client/scripts/main.gd` (baseline 6193) — call-site replacements only; no new coordinator logic
- [ ] `client/scripts/player_status_effect_markers.gd` (baseline 646) — **must shrink** by extracting aura code
- [ ] `client/scripts/bot_scenario_runner.gd` — only if scenario 89 needs new wait keys
- [ ] Other over-limit file: none expected

Decision:

- [x] Extract `aura_soft_lights.gd` + `aura_light_presentation_loader.gd` as part of this slice
- [x] `player_status_effect_markers.gd` ends at or below its 646-line baseline after aura removal

Verification:

```bash
make maintainability
```

## Task 1 — Shared aura light presentation catalog

Files:

- Create: `shared/assets/aura_light_presentation.v0.json`
- Create: `shared/assets/aura_light_presentation.v0.schema.json`

- [ ] Step 1.1: Define catalog entries for `sanctuary`, `holy_shield`, `rage`, `elite_command`,
  `elite_command_radius_preview` with color hex, `omni_energy`, `omni_attenuation`, `height_offset`,
  `shadow_enabled: false`, and optional `hero` / `monster` multiplier objects.
- [ ] Step 1.2: Encode aura priority order in JSON (`priority: ["sanctuary", "holy_shield", ...]`).
- [ ] Step 1.3: Define `radius_source` per aura:
  - `skill_effect_radius` + `skill_id` for sanctuary/holy_shield
  - `presentation_personal_radius` for rage and elite_command follower
  - `dungeon_elite_aura_radius` for leader preview
- [ ] Step 1.4: Define `cast_pulse` block for holy_shield (duration, energy peak multiplier).

```bash
make validate-shared
```

## Task 2 — Client loader + aura soft lights module

Files:

- Create: `client/scripts/aura_light_presentation_loader.gd`
- Create: `client/scripts/aura_soft_lights.gd`

- [ ] Step 2.1: Implement loader (`ensure_loaded`, `config`, `aura_entry(aura_id)`, `priority_list()`).
- [ ] Step 2.2: Implement `AuraSoftLights.sync_aura(root, state)` — creates/updates/removes single
  `OmniLight3D` named `AuraSoftLight`; stores `active_aura_id` meta for debug.
- [ ] Step 2.3: Implement priority resolver from `effect_ids`, local rage flag, leader preview flag.
- [ ] Step 2.4: Implement `apply_light_config(light, aura_id, entity_kind, radius)` using catalog +
  hero/monster multipliers; mirror torch omni defaults (`shadow_enabled = false`).
- [ ] Step 2.5: Implement `pulse_holy_shield_cast(source_root, affected_roots, radius)` — brief
  energy/range tween on caster (+ optional subtle bump on affected allies); no mesh children.
- [ ] Step 2.6: Expose debug helpers: `active_aura_id(root)`, `aura_light_range(root)`,
  `has_aura(root, aura_id)` for bot compatibility with existing `has_rage_effect` etc.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_aura_soft_lights.gd
```

## Task 3 — Wire coordinator + elite preview sync

Files:

- Modify: `client/scripts/player_status_effect_markers.gd`
- Modify: `client/scripts/elite_aura_preview_sync.gd`
- Modify: `client/scripts/main.gd`

- [ ] Step 3.1: Delete aura mesh factories (`make_rage_effect`, `make_holy_shield_effect`,
  `make_sanctuary_dome_effect`, `make_elite_command_effect`, `make_elite_command_radius_preview`,
  `make_holy_shield_aura_pulse`, `make_holy_shield_target_pulse`, `pulse_holy_shield_aura`).
- [ ] Step 3.2: Replace aura `sync_*` calls in `main.gd` with `AuraSoftLights.sync_aura` (or thin
  wrappers kept for readability). Preserve debuff `sync_*` unchanged.
- [ ] Step 3.3: Route `_pulse_holy_shield_aura` to `AuraSoftLights.pulse_holy_shield_cast`.
- [ ] Step 3.4: Update `elite_aura_preview_sync.gd` to set leader preview state then call
  `AuraSoftLights`; simplify `_all_marker_names()` to debuff markers + `AuraSoftLight` for culling.
- [ ] Step 3.5: Repoint `has_rage_effect`, `has_holy_shield_effect`, `has_sanctuary_effect`,
  `has_elite_command_effect`, `has_elite_command_radius_preview` to light-based checks.

```bash
make client-unit
```

## Task 4 — Bot debug + unit tests

Files:

- Modify: `client/scripts/bot_presentation_debug.gd`
- Modify: `client/tests/test_status_effect_presentation.gd`
- Create: `client/tests/test_aura_soft_lights.gd`
- Modify: `scripts/client_smoke.sh`

- [ ] Step 4.1: Add `active_aura_id`, `aura_light_range` to local player + entity debug rows.
- [ ] Step 4.2: Update holy-shield test: assert cast pulse via tween counter or transient range bump,
  not `holy_shield_aura_pulses` mesh count (rename debug key if needed, e.g. `holy_shield_cast_pulses`).
- [ ] Step 4.3: Update rage/sanctuary/holy-shield/elite aura unit tests to assert `OmniLight3D` and
  `active_aura_id`; leave burning/pinning/stun/rogue-mark tests unchanged.
- [ ] Step 4.4: Add focused `test_aura_soft_lights.gd` for priority resolution (sanctuary beats
  holy_shield beats rage) and radius-source wiring.
- [ ] Step 4.5: Register new test in `scripts/client_smoke.sh`.

```bash
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_status_effect_presentation.gd
godot --headless --rendering-method gl_compatibility --path client --script res://tests/test_aura_soft_lights.gd
make client-unit
```

## Task 5 — Bot scenarios

Files:

- Modify: `tools/bot/scenarios/client/34_elite_aura_readability.json`
- Modify: `tools/bot/scenarios/client/37_elite_aura_radius_preview.json`
- Create: `tools/bot/scenarios/client/89_aura_soft_lights.json` (`"ci_tier": "extended"`)
- Modify: `client/scripts/bot_scenario_runner.gd` (only if new wait keys needed)

- [ ] Step 5.1: Keep scenario 34 passing via `has_elite_command_effect` (now light-backed) on a buffed
  follower in `dungeon_levels` / seed `v112_pack_metadata`.
- [ ] Step 5.2: Keep scenario 37 passing; assert leader `has_elite_command_radius_preview` and
  `elite_command_radius_min/max` against rule-derived radius (~4.0).
- [ ] Step 5.3: Add scenario 89 in `skill_progression_lab` (or `combat_control_lab`): cast `rage`,
  wait `has_rage_effect`; if paladin/holy shield available in lab, cast holy shield and assert
  `active_aura_id == holy_shield`. Classify extended only — do not add to `ci_pack.json`.
- [ ] Step 5.4: Add `bot_scenario_runner` support for `active_aura_id` wait filter if scenario 89
  uses it.

```bash
HEADLESS=1 make bot-client SCENARIO=34_elite_aura_readability
HEADLESS=1 make bot-client SCENARIO=37_elite_aura_radius_preview
HEADLESS=1 make bot-client SCENARIO=89_aura_soft_lights
```

Manual visual:

```bash
make bot-visual scenario=34_elite_aura_readability
make bot-visual scenario=37_elite_aura_radius_preview
make bot-visual scenario=89_aura_soft_lights
```

## Task 6 — Lifecycle docs and CI

Files:

- Create: `docs/as-built/v355_aura-soft-lights.md`
- Modify: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`

- [ ] Step 6.1: Record shipped behavior, proof commands, and deferred production VFX/audio.
- [ ] Step 6.2: Update progress tables (`Latest completed slice: v355`).

```bash
make maintainability
make ci
```

## Final verification

- [ ] `make validate-shared`
- [ ] `make client-unit`
- [ ] `HEADLESS=1 make bot-client SCENARIO=34_elite_aura_readability`
- [ ] `HEADLESS=1 make bot-client SCENARIO=37_elite_aura_radius_preview`
- [ ] `HEADLESS=1 make bot-client SCENARIO=89_aura_soft_lights`
- [ ] `make maintainability`
- [ ] `make ci`

## Deferred (explicit)

- Debuff mesh/tint presentation (burning, mark, stun, poison, ice slow)
- Production aura VFX/audio, nameplates, minimap markers
- Settings quality toggle for aura lights
- forward_plus perf re-baseline
- Protocol-visible aura state
