# v193 Plan - Unique Skill Modifier

Status: Complete
Goal: Add a named unique effect that modifies one configured skill's server-owned damage.
Architecture: The unique effect remains data-driven through `unique_effects.v0.json` and is applied
inside the existing skill damage path before hit resolution. The first hook supports a bounded
percent damage bonus for a configured `skill_id`; it does not alter generic skill affix stats,
basic attacks, or non-target skills. Existing unique chest payloads provide the protocol proof.
Tech stack: shared JSON/schema, Go sim, Python bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v192 and existing unique item/effect infrastructure. No client UI/art/camera work is in
scope; existing unique chest and tooltip payloads already render named unique effect summaries, so
the Godot plugin adoption checklist is not applicable.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/unique_effects.v0.schema.json` | Allow the new unique hook. |
| Modify | `shared/rules/unique_effects.v0.json` | Add a skill-specific damage modifier effect. |
| Modify | `shared/rules/unique_items.v0.json` | Add a named unique item that carries the effect. |
| Modify | `server/internal/game/item_skill_stats.go` | Apply skill-specific unique damage modifiers. |
| Modify | `server/internal/game/sim.go` / skill-specific helpers | Pass skill identity to the damage path. |
| Add/Modify | `server/internal/game/*unique*_test.go` | Focused modifier and chest coverage. |
| Add | `tools/bot/scenarios/82_unique_skill_modifier.json` | Protocol proof for named unique availability. |
| Modify | `PROGRESS.md` | Mark v193 complete at finish. |
| Add | `docs/as-built/v193_unique-skill-modifier.md` | Record what shipped. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/sim.go`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Keep the new logic in focused helper/test files; only thread skill identity through `sim.go`.

Verification:
```bash
make maintainability
```

## Task 1 - Shared unique contract

Files:
- Modify: `shared/rules/unique_effects.v0.schema.json`
- Modify: `shared/rules/unique_effects.v0.json`
- Modify: `shared/rules/unique_items.v0.json`

- [x] Add the `on_skill_damage_roll` hook.
- [x] Add a ready effect with `skill_id: magic_bolt` and `damage_bonus_percent`.
- [x] Add a named unique staff carrying the effect.

```bash
make validate-shared
```

## Task 2 - Server authority and tests

Files:
- Modify: `server/internal/game/item_skill_stats.go`
- Modify: `server/internal/game/sim.go`
- Modify: `server/internal/game/unique_effects_test.go`

- [x] Apply the unique modifier only when the cast/projectile skill id matches the effect.
- [x] Keep basic attacks and other skills on the baseline path.
- [x] Add focused Go tests for matching and non-matching skills.

```bash
cd server && go test ./internal/game -run 'UniqueSkill|UniqueChest|UniqueItemValidation' -count=1
```

## Task 3 - Bot proof

Files:
- Add: `tools/bot/scenarios/82_unique_skill_modifier.json`

- [x] Add a compact unique chest scenario that takes the new named unique.

```bash
make bot scenario=82_unique_skill_modifier.json
```

## Task 4 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v193_unique-skill-modifier.md`

- [x] Add lifecycle row and as-built note.
- [x] Run final verification.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'UniqueSkill|UniqueChest|UniqueItemValidation' -count=1`
- [x] `make bot scenario=82_unique_skill_modifier.json`
- [x] `make ci`
