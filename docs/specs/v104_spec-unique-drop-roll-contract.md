# v104 Spec: Unique Drop Roll Contract

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-12
Codename: `unique-drop-roll-contract`

## Purpose

Make unique a real item roll outcome. When an item rolls unique, the server rolls normal
depth/template stats first, then attaches exactly one compatible global unique effect id from
`unique_effects.v0.json`.

## Non-goals

- No combat execution for unique effects.
- No client VFX or burning visual cue.
- No shop or mystery-seller unique eligibility.
- No final unique drop chance tuning.

## Acceptance Criteria

- `unique` is a valid equipment rarity with its own name prefix and stat-roll count.
- Unique item rolls attach exactly one compatible effect id.
- Non-unique item rolls keep their existing effect ids.
- Item roll payloads continue using the existing `effect_ids` field; no protocol schema bump is
  required.
- The deterministic item-roll golden includes a unique case proving unique effect attachment.
- Server tests prove unique effects are loaded and incompatible effects are not selected.
- `make validate-shared`, `cd server && go test ./internal/game/...`, `make maintainability`, and
  `make ci` pass.

## Scope And Likely Files

- `shared/rules/item_templates.v0.json`
- `server/internal/game/rules.go`
- `server/internal/game/shop.go`
- `server/internal/game/game_test.go`
- `shared/golden/item_rolls.json`
- `tools/validate_shared.py`
- lifecycle docs

## Test And Bot Proof

- Go golden coverage proves the unique roll contract deterministically.
- No bot scenario is required in v104 because the effect is not executable yet; v105 adds the
  combat and visual bot proof.

## Open Questions And Risks

| Risk | Mitigation |
|------|------------|
| Adding `unique` changes deterministic roll outcomes. | Update the named item-roll golden intentionally and rely on CI for broader drift. |
| Unique effects could attach to incompatible item types. | Loader/test filters by `compatible_item_types`. |

## ADR Alignment

- ADR-0014 D2/D5: adds loot hope through behavior-changing unique effects while preserving normal
  stats and item scaling.
- ADR-0012/0013: leaves economy and mystery-seller eligibility deferred.
