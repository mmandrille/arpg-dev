# v464 Plan — Combat Input Flow Polish

Status: Ready for implementation
Goal: Make ordinary basic combat clicks, buffered follow-ups, target changes, and movement retargets feel continuous without changing authoritative outcomes.
Architecture: This slice stays client-side and presentation-owned. Existing Godot helpers own attack buffering, sticky target, movement retarget grace, local attack presentation, and melee lunge; the server remains authoritative for accepted attacks, damage, death, and loot. Shared changes are limited to existing presentation data if tuning is needed, with no protocol/schema bump.
Tech stack: Godot client, shared presentation JSON, client bot scenarios, lifecycle docs.

## Baseline and shortcut decision

Builds on v463 `dungeon-surface-detail-overlays` and reuses prior combat-feel work from v428 plus
locomotion smoothing from v461.

Asset/plugin decision:
- Adopt: existing `combat_feel_presentation.v0.json`, `CombatInputBuffer`, `CombatStickyTarget`,
  `CommandRetargetGrace`, `CombatLocalAttackPresentation`, `MeleeLungePresentation`, and client bot
  visual scenario infrastructure.
- Borrow: current `77_input_buffering` and `82_melee_lunge_micro_step` scenario patterns.
- Reject: external input plugins, camera/combat addons, imported assets, production VFX/audio, and
  server-side balance retuning.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/attack_move_input_coordinator.gd` | Replace stale pending combat intent when target or movement intent changes. |
| Modify | `client/scripts/combat_input_buffer.gd` | Track one buffered target safely and expose focused debug state if needed. |
| Modify | `client/scripts/combat_sticky_target.gd` | Track target replacement and support explicit replacement tests if needed. |
| Modify | `client/scripts/command_retarget_grace.gd` | Ensure movement retarget can clear combat intent through caller-owned hooks. |
| Modify | `client/tests/test_sustained_input.gd` | Unit coverage for buffer/sticky target replacement and stale target cleanup. |
| Modify | `client/tests/test_command_retarget_grace.gd` | Unit coverage for movement retarget replacing pending combat intent via callback. |
| Modify | `client/tests/test_melee_lunge_presentation.gd` | Keep lunge bounded and recovering after repeated local melee starts if changed. |
| Create | `tools/bot/scenarios/client/103_combat_input_flow_polish.json` | Focused visual proof for target replacement / movement retarget in live client. |
| Modify | `docs/specs/v464_spec-combat-input-flow-polish.md` | Mark complete once shipped. |
| Create | `docs/as-built/v464_combat-input-flow-polish.md` | Record shipped behavior and proof. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v464 lifecycle row. |
| Modify | `PROGRESS.md` | Update current status and deferred scope. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd` — touched for camera validity guard; baseline exception documented.
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [ ] Extract focused helper/module/test file as part of this slice, or
- [x] Defer extraction with rationale: this slice stayed inside already-extracted combat-input helpers; the only `main.gd` touch was a small camera validity guard needed to keep bot visuals running.

Verification:

```bash
make maintainability
```

## Task 1 — Combat input helper behavior

Files:
- Modify: `client/scripts/attack_move_input_coordinator.gd`
- Modify: `client/scripts/combat_input_buffer.gd`
- Modify: `client/scripts/combat_sticky_target.gd`
- Modify: `client/scripts/command_retarget_grace.gd`
- Modify: `client/tests/test_sustained_input.gd`
- Modify: `client/tests/test_command_retarget_grace.gd`

- [x] Step 1.1: Ensure monster clicks replace stale buffered/sticky target intent with the newest living monster target.
- [x] Step 1.2: Ensure floor retarget during local recovery can clear pending combat intent before queuing or sending movement.
- [x] Step 1.3: Add focused unit tests for target replacement and movement retarget cleanup.

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
godot --headless --path client --script res://tests/test_command_retarget_grace.gd
```

## Task 2 — Local melee feedback remains bounded

Files:
- Modify: `client/scripts/combat_local_attack_presentation.gd`
- Modify: `client/scripts/melee_lunge_presentation.gd`
- Modify: `client/tests/test_melee_lunge_presentation.gd`
- Modify: `shared/assets/combat_feel_presentation.v0.json` only if presentation tuning is needed

- [x] Step 2.1: Keep local attack start feedback immediate while avoiding duplicate result animation for the same attack result.
- [x] Step 2.2: Prove repeated melee presentation starts replace prior lunge tween and settle back to the model base.
- [x] Step 2.3: If any presentation value changes, keep it in `combat_feel_presentation.v0.json` and validate shared assets.

```bash
make validate-shared
godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd
godot --headless --path client --script res://tests/test_combat_feel_presentation_loader.gd
```

## Task 3 — Client bot visual proof

Files:
- Create: `tools/bot/scenarios/client/103_combat_input_flow_polish.json`
- Modify: client bot assertion/action files only if existing steps cannot prove replacement

- [x] Step 3.1: Add a focused Godot-client scenario in `combat_control_lab` that clicks one monster, replaces pending combat intent or retargets movement during recovery, and observes later combat/movement presentation.
- [x] Step 3.2: Keep the scenario `ci_tier` as `extended`; do not add it to the fast CI pack unless existing coverage is insufficient.
- [x] Step 3.3: Re-run existing combat feel scenarios plus the new proof.

```bash
HEADLESS=1 make bot-visual scenario=77_input_buffering
HEADLESS=1 make bot-visual scenario=82_melee_lunge_micro_step
HEADLESS=1 make bot-visual scenario=103_combat_input_flow_polish
```

## Task 4 — Lifecycle docs and closeout

Files:
- Modify: `docs/specs/v464_spec-combat-input-flow-polish.md`
- Create: `docs/as-built/v464_combat-input-flow-polish.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `PROGRESS.md`

- [x] Step 4.1: Mark the spec complete and write the as-built with shipped behavior, verification, and deferred balance scope.
- [x] Step 4.2: Add v464 to the lifecycle table and update `PROGRESS.md` current status to v464 complete / v465 TBD.
- [x] Step 4.3: Keep broader cooldown, movement-speed, and attack-speed retuning deferred.

```bash
make maintainability
```

## Final verification

- [x] `make validate-shared`
- [x] `godot --headless --path client --script res://tests/test_sustained_input.gd`
- [x] `godot --headless --path client --script res://tests/test_command_retarget_grace.gd`
- [x] `godot --headless --path client --script res://tests/test_melee_lunge_presentation.gd`
- [x] `godot --headless --path client --script res://tests/test_combat_feel_presentation_loader.gd`
- [x] `make client-unit`
- [x] `HEADLESS=1 make bot-visual scenario=77_input_buffering`
- [x] `HEADLESS=1 make bot-visual scenario=82_melee_lunge_micro_step`
- [x] `HEADLESS=1 make bot-visual scenario=103_combat_input_flow_polish`
- [x] `make maintainability`
- [x] `make ci` during finish before commit

## Deferred scope

- Authoritative basic-attack cooldown rebalance.
- Shared-rules movement-speed or acceleration retuning.
- Skill/ranged/boss combat flow redesign.
- Production combat VFX/audio.
