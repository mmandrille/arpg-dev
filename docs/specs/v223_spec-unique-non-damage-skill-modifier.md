# v223 Spec - Unique Non-Damage Skill Modifier

Status: Complete
Date: 2026-06-16
Codename: `unique-non-damage-skill-modifier`
Baseline: v222 `upgrade-result-preview`

## Purpose

Unique items should change build behavior, not only add damage. The current catalog has the live
`blood_price` effect, which lets skill casts pay missing mana with HP, but no named unique package
that clearly presents that non-damage skill modifier to players. This slice adds a named unique
item for that effect and proves the named item changes skill use behavior when equipped.

## Non-goals

- No new unique-effect hook or combat formula.
- No new active skill, passive tree, resource type, or skill UI redesign.
- No protocol/schema bump; existing rolled item `effect_ids`, unique chest, inventory, and skill-cast
  events are reused.
- No balancing pass for Blood Price beyond the existing effect data.
- No market restrictions, binding rules, or production art/audio for the new unique.

## Acceptance criteria

- Shared unique item rules include a named ready unique, `Bloodbound Sigil`, with exactly the
  `blood_price` fixed effect and valid fixed stats/requirements.
- Named unique payload generation includes the new item with stable display name, template, stats,
  requirements, and effect IDs.
- A server test equips the named unique package and proves a zero-mana skill cast succeeds by paying
  HP through the `blood_price` effect.
- A protocol bot scenario opens the purple town unique chest, takes `Bloodbound Sigil`, and asserts
  the inventory item has `effect_ids: ["blood_price"]`.
- The existing client unique chest proof shows the new named unique row with the Blood Price summary.
- The change aligns with ADR-0014 D5 by adding a behavior-changing non-damage unique, while preserving
  server authority over skill casts and resource payment.

## Scope and likely files

- `shared/rules/unique_items.v0.json` - add `bloodbound_sigil`.
- `server/internal/game/unique_chest_test.go` - named payload coverage for the new package.
- `server/internal/game/unique_effects_test.go` - named-item Blood Price behavior proof.
- `tools/bot/scenarios/91_unique_non_damage_skill_modifier.json` - protocol chest/inventory proof.
- `tools/bot/scenarios/client/40_unique_chest_client_proof.json` - client row/summary proof.
- Lifecycle docs: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/progress/slice-codename-index.md`, `docs/as-built/v223_unique-non-damage-skill-modifier.md`.

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'NamedUnique|BloodPrice|UniqueTestChest' -count=1`
- `make bot scenario=unique_non_damage_skill_modifier`
- `make bot-client scenario=unique_chest_client_proof`
- `make maintainability`

Manual visual proof, if desired:

```bash
make bot-visual scenario=unique_chest_client_proof
```

## Client asset/plugin decision

Reject external assets/plugins. The named item reuses existing unique chest, stash panel, tooltip,
and text-summary presentation.

## Open questions and risks

- No blocking questions. The existing Blood Price effect already owns the mechanic; this slice only
  gives it a named unique package and proves that named package drives the live behavior.
- Risk: adding another named item changes unique chest counts. Existing tests derive counts from
  rules, and the bot/client checks assert the named row directly.
