# v223 Plan - Unique Non-Damage Skill Modifier

Status: Complete
Goal: Add a named Blood Price unique item and prove it changes skill use through HP-for-mana payment.
Architecture: Reuse the existing server-authoritative `blood_price` unique effect. Add a named
unique data package, then prove named payload generation, equipped behavior, protocol inventory
payloads, and client unique-chest presentation. No protocol or formula changes are planned.
Tech stack: Shared JSON rules, Go game tests, Python protocol bot scenario, Godot client bot proof,
lifecycle docs.

## Baseline and shortcut decision

Builds on v108's live Blood Price effect and v82/v136 unique chest/named unique presentation. The
slice chooses data/package exposure plus tests instead of introducing a new unique hook.

Asset/plugin decision: reject external assets/plugins. Reuse existing unique chest and stash tooltip
presentation.

## Spec Review

- Baseline: v223 follows v222 on `main`.
- Scope: named unique package plus proofs; no new mechanics or protocol fields.
- Contracts: shared unique item data changes only; schema already supports the package.
- Server authority: skill cast, HP payment, mana checks, and item effects remain server-owned.
- ADR alignment: advances ADR-0014 D5 with a non-damage behavior hook.
- Bot proof: protocol and client scenarios assert the new named item and effect summary.
- Maintainability: touched files are small or already grandfathered; no coordinator growth is needed.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/unique_items.v0.json` | Add `bloodbound_sigil` named unique package. |
| Modify | `server/internal/game/unique_chest_test.go` | Assert named payload identity/stats/effect IDs. |
| Modify | `server/internal/game/unique_effects_test.go` | Prove named Blood Price item pays missing mana with HP. |
| Add | `tools/bot/scenarios/91_unique_non_damage_skill_modifier.json` | Protocol chest/inventory proof. |
| Modify | `tools/bot/scenarios/client/40_unique_chest_client_proof.json` | Client unique chest summary proof. |
| Add | `docs/as-built/v223_unique-non-damage-skill-modifier.md` | Record proof and deferred scope. |
| Modify | `PROGRESS.md` | Advance current status after the slice. |
| Modify | `docs/progress/slice-lifecycle.md` | Add v223 lifecycle row. |
| Modify | `docs/progress/slice-codename-index.md` | Add v223 codename mapping. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `server/internal/game/unique_effects_test.go` is grandfathered; add focused coverage only.
- [x] Other touched source/test/tool files should remain within their current baselines.
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Defer extraction with rationale: v223 adds narrow cases to existing unique-effect/chest tests
  and does not introduce a new implementation coordinator.

Verification:

```bash
make maintainability
```

## Task 1 - Add named unique package

Files:
- Modify: `shared/rules/unique_items.v0.json`

- [x] Add `bloodbound_sigil` using a compatible ring/amulet template, fixed stats, and
  `fixed_effect_ids: ["blood_price"]`.

```bash
make validate-shared
```

## Task 2 - Server proof

Files:
- Modify: `server/internal/game/unique_chest_test.go`
- Modify: `server/internal/game/unique_effects_test.go`

- [x] Extend named unique payload coverage for `Bloodbound Sigil`.
- [x] Add or update Blood Price behavior coverage so the named item package, when equipped, pays
  missing mana with HP and allows the skill cast.

```bash
cd server && go test ./internal/game -run 'NamedUnique|BloodPrice|UniqueTestChest' -count=1
```

## Task 3 - Bot and client presentation proof

Files:
- Add: `tools/bot/scenarios/91_unique_non_damage_skill_modifier.json`
- Modify: `tools/bot/scenarios/client/40_unique_chest_client_proof.json`

- [x] Add protocol scenario to take `Bloodbound Sigil` and assert `blood_price` effect IDs.
- [x] Add client unique chest assertion for the Blood Price summary.

```bash
make bot scenario=unique_non_damage_skill_modifier
make bot-client scenario=unique_chest_client_proof
```

## Task 4 - Lifecycle docs and close-out

Files:
- Add: `docs/as-built/v223_unique-non-damage-skill-modifier.md`
- Modify: `PROGRESS.md`
- Modify: `docs/progress/slice-lifecycle.md`
- Modify: `docs/progress/slice-codename-index.md`
- Modify: `docs/specs/v223_spec-unique-non-damage-skill-modifier.md`
- Modify: `docs/plans/v223_2026-06-16-unique-non-damage-skill-modifier.md`

- [x] Mark this plan complete as tasks pass.
- [x] Mark the spec complete.
- [x] Update current status to v223 complete and selected feature queue complete.
- [x] Add the lifecycle row and as-built summary.

```bash
make maintainability
```

## Final verification

- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'NamedUnique|BloodPrice|UniqueTestChest' -count=1`
- [x] `make bot scenario=unique_non_damage_skill_modifier`
- [x] `make bot-client scenario=unique_chest_client_proof`
- [x] `make maintainability`

Batch-level `make ci` remains deferred to `$autoloop` after v223 is committed.
