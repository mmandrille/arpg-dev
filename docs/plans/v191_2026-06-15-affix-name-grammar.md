# v191 Plan - Affix Name Grammar

Status: Ready for implementation
Goal: Add deterministic affix-style display names for rolled non-unique equipment.
Architecture: The server already persists rolled item metadata in `ItemRollPayload`. This slice
keeps the protocol unchanged and changes only the display name generated at roll time. Existing
inventory, stash, shop, market, and client panels continue to render `display_name`.
Tech stack: Go sim item rolling, Python bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v190. No client UI or art is in scope, so the Godot plugin adoption checklist is not
required for implementation.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `server/internal/game/affix_names.go` | Generate deterministic affix display names. |
| Modify | `server/internal/game/shop.go` | Use affix display names from the roll payload path. |
| Modify | `server/internal/game/item_rarity_test.go` | Prove rare skill-affix naming. |
| Modify | `shared/golden/item_rolls.json` | Update rolled item display-name golden. |
| Modify | `shared/golden/shop_offers.json` | Update generated shop display-name golden. |
| Modify | `tools/validate_shared.py` | Keep the deterministic shop-offer mirror aligned with affix names. |
| Modify | `tools/bot/scenarios/80_skill_affix_rolls.json` | Assert generated display name. |
| Modify | `PROGRESS.md` | Mark v191 complete at finish. |
| Add | `docs/as-built/v191_affix-name-grammar.md` | Record what shipped. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: `server/internal/game/shop.go`
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract focused helper/module/test file as part of this slice.

Verification:
```bash
make maintainability
```

## Task 1 - Server affix naming

Files:
- Create: `server/internal/game/affix_names.go`
- Modify: `server/internal/game/shop.go`
- Modify: `server/internal/game/item_rarity_test.go`
- Modify: `shared/golden/item_rolls.json`
- Modify: `shared/golden/shop_offers.json`
- Modify: `tools/validate_shared.py`

- [x] Add a deterministic display-name helper for non-unique, non-set rolled item payloads.
- [x] Prefer rolled stat families over base stats when choosing the affix word.
- [x] Add focused Go coverage for a rare skill-affix staff name.

```bash
cd server && go test ./internal/game -run 'TestAffixName|TestSkillAffix' -count=1
```

## Task 2 - Bot proof

Files:
- Modify: `tools/bot/scenarios/80_skill_affix_rolls.json`

- [x] Assert the deterministic display name in the existing skill-affix protocol scenario.
- [x] Run the focused scenario.

```bash
make bot scenario=80_skill_affix_rolls.json
```

## Task 3 - Lifecycle docs and CI

Files:
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v191_affix-name-grammar.md`

- [x] Add lifecycle row and as-built note.
- [x] Run final verification.

```bash
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game -run 'TestItemRollsGolden|TestShopGeneratedOfferGolden|TestAffixName|TestSkillAffix' -count=1`
- [x] `make bot scenario=80_skill_affix_rolls.json`
- [x] `make ci`
