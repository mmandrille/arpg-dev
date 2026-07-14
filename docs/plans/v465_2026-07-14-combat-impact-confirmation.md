# v465 Plan — Combat Impact Confirmation

Status: Complete
Goal: Make authoritative combat outcomes visibly clearer without changing combat simulation.
Architecture: Keep the server as the only source of combat outcomes. Expand the existing Godot
`CombatOutcomePunch` presentation mapper so normal hits and kills join the already-supported miss,
block, immune, and critical result confirmations. Use focused Godot tests for presentation rules and
a client bot scenario for authoritative hit/kill flow through `combat_control_lab`.
Tech stack: Godot client scripts/tests, JSON bot scenarios, docs.

## Baseline and shortcut decision

Builds on v464 combat input flow polish, especially the local-start/result matching path that already
prevents duplicate local attack playback on authoritative results.

Asset/plugin decision:
- Adopt: existing procedural Godot outcome punch, combat text, model reaction, and combat lab assets.
- Borrow: existing `enemy_impact_feedback` gating and `wait_entity_reaction` bot proof pattern.
- Reject: new imported VFX/audio assets, third-party plugins, and new asset pipelines.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/combat_outcome_punch.gd` | Map normal hits and kills to bounded confirmation nodes. |
| Modify | `client/tests/test_combat_outcome_punch.gd` | Cover spawn rules, node shaping, and event-presentation integration. |
| Modify | `client/tests/test_impact_sparks.gd` | Assert new hit confirmation replaces the old no-sparks expectation and respects feedback toggle. |
| Add | `tools/bot/scenarios/client/104_combat_impact_confirmation.json` | Prove authoritative hit/kill flow in the combat control lab. |
| Modify | `docs/progress/scenario-movement-audit.tsv` | Register the new client bot scenario. |
| Add | `docs/as-built/v465_combat-impact-confirmation.md` | Record shipped behavior and proof. |
| Modify | `PROGRESS.md` | Close the slice during finish. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [ ] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [ ] Defer extraction with rationale: not expected; changes stay in existing focused presentation/test files.

Verification:

```bash
make maintainability
```

## Task 1 — Outcome Mapping

Files:
- Modify: `client/scripts/combat_outcome_punch.gd`

- [ ] Step 1.1: Extend outcome classification so `monster_damaged` normal hits and `monster_killed` results spawn.
- [ ] Step 1.2: Tune hit/kill ring and spark intensity within the existing short-lived procedural node.
- [ ] Step 1.3: Preserve existing miss/block/immune/critical distinctions.

```bash
godot --headless --path client --script res://tests/test_combat_outcome_punch.gd
```

## Task 2 — Presentation Tests

Files:
- Modify: `client/tests/test_combat_outcome_punch.gd`
- Modify: `client/tests/test_impact_sparks.gd`

- [ ] Step 2.1: Update spawn-rule tests to expect normal hit and kill confirmation.
- [ ] Step 2.2: Add integration assertions for `monster_damaged` and `monster_killed`.
- [ ] Step 2.3: Add disabled-toggle coverage proving `enemy_impact_feedback=false` blocks hit confirmation.

```bash
godot --headless --path client --script res://tests/test_combat_outcome_punch.gd
godot --headless --path client --script res://tests/test_impact_sparks.gd
```

## Task 3 — Bot Scenario

Files:
- Add: `tools/bot/scenarios/client/104_combat_impact_confirmation.json`
- Modify: `docs/progress/scenario-movement-audit.tsv`

- [ ] Step 3.1: Add a client scenario that picks up the training bow, damages the combat lab monster, waits for impact feedback, and continues until kill.
- [ ] Step 3.2: Keep the scenario `ci_tier` extended unless this reveals a merge-blocking gap.
- [ ] Step 3.3: Register the scenario in the movement audit.

```bash
HEADLESS=1 make bot-visual scenario=104_combat_impact_confirmation
```

## Task 4 — Lifecycle Docs and CI

Files:
- Add: `docs/as-built/v465_combat-impact-confirmation.md`
- Modify: `PROGRESS.md`

- [ ] Step 4.1: Document the shipped client behavior and exact verification commands.
- [ ] Step 4.2: Update `PROGRESS.md` current status, completed slice, open gaps, and checklist.
- [ ] Step 4.3: Run focused and final verification.

```bash
make maintainability
make client-unit
make ci
```

## Final verification

- [x] `godot --headless --path client --script res://tests/test_combat_outcome_punch.gd`
- [x] `godot --headless --path client --script res://tests/test_impact_sparks.gd`
- [x] `HEADLESS=1 make bot-visual scenario=104_combat_impact_confirmation`
- [x] `make maintainability`
- [x] `make client-unit`
- [x] `make ci`
