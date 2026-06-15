# v172 Plan - Loot Filter Persistence

Status: Complete
Goal: Persist the existing loot label rarity filter as a local client setting.
Architecture: Store the display-only mode label in `ClientSettings` alongside other local
preferences. Initialize `LootLabelFilter` from that value during client startup and save after each
cycle. Server authority and protocol contracts are unchanged.
Tech stack: Godot GDScript client settings, focused GDScript unit tests, lifecycle docs.

## Baseline and shortcut decision

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/client_settings.gd` | Persist and normalize `loot_filter_mode` |
| Modify | `client/scripts/loot_label_filter.gd` | Restore threshold from mode label |
| Modify | `client/scripts/main.gd` | Load setting into filter and save after cycling |
| Modify | `client/tests/test_client_bot.gd` | Cover settings parse/save/reload shape |
| Modify | `client/tests/test_loot_label_filter.gd` | Cover mode restore and invalid fallback |
| Add | `docs/as-built/v172_loot-filter-persistence.md` | Shipped proof |
| Modify | `PROGRESS.md` | Lifecycle and deferred scope |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [ ] `server/internal/game/game_test.go`
- [ ] `tools/bot/run.py`
- [ ] `tools/validate_shared.py`
- [ ] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected
- [x] Did every touched grandfathered file stay at or below its baseline (touch-to-shrink)?

Decision:
- [x] Extract focused helper/module/test file as part of this slice, or
- [ ] Defer extraction with rationale: `<why splitting now is riskier>`

Verification:
```bash
make maintainability
```

## Task 1 - Client settings and filter restore

Files:
- Modify: `client/scripts/client_settings.gd`
- Modify: `client/scripts/loot_label_filter.gd`
- Modify: `client/scripts/main.gd`

- [x] Add normalized `loot_filter_mode` load/save support to `ClientSettings`.
- [x] Add `LootLabelFilter.set_mode_label()` / restore behavior.
- [x] Initialize `_loot_filter` from loaded settings and save after cycle input.

```bash
make client-unit
```

## Task 2 - Focused tests

Files:
- Modify: `client/tests/test_client_bot.gd`
- Modify: `client/tests/test_loot_label_filter.gd`

- [x] Cover settings default, parse, save, and reload of `loot_filter_mode`.
- [x] Cover valid mode restore and invalid fallback in `LootLabelFilter`.

```bash
make client-unit
```

## Task 3 - Lifecycle docs and CI

Files:
- Add: `docs/as-built/v172_loot-filter-persistence.md`
- Modify: `PROGRESS.md`
- Modify: this plan

- [x] Record shipped proof and update lifecycle docs.
- [x] Run final verification.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`

## Deferred scope

- Category filtering and additional filter modes.
- Settings-panel controls for loot filtering.
- Account-synced settings.
