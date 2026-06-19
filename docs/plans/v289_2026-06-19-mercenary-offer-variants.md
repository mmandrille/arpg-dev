# v289 Plan — Mercenary Offer Variants

Status: Complete
Goal: Let the mercenary board hire one of multiple server-authored mercenary offers.
Architecture: Keep the board workflow one-click and server-authoritative. Add a shared mercenary
offer catalog consumed by Go only for now; the client receives the selected offer through existing
events. Use deterministic seed+board selection so bots and replay stay stable.
Tech stack: Shared JSON/schema, Go sim/rules loader, Godot panel/debug state, protocol bot, client
bot, asset catalog tooling.

## Baseline and shortcut decision

Builds on v206 fixed mercenary hiring, v207/v219/v239/v247 mercenary panel proofs, and ADR-0010.
The slice adds authored offer variety without implementing player-character-derived mercenary
listings.

Asset/plugin decision:

- Adopt: existing companion AI, mercenary board interaction, companion HUD/panel, and monster dummy
  asset.
- Borrow: existing mercenary hiring protocol/client bot scenario structure.
- Reject: external assets/plugins, new model pipelines, offer picker UI, player snapshots, ranged
  mercenary AI, and per-offer pricing.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `shared/rules/mercenaries.v0.json` | Authored mercenary offer catalog. |
| Create | `shared/rules/mercenaries.v0.schema.json` | Schema for offer ids and monster refs. |
| Modify | `shared/rules/monsters.v0.json` | Add `mercenary_scout` combat stats. |
| Modify | `shared/assets/monster_visuals.v0.json` | Map scout to an existing monster asset. |
| Modify | `shared/assets/model_preview_catalog.v0.json` | Regenerated preview catalog after visual mapping. |
| Modify | `shared/i18n/en.json`, `shared/i18n/es.json` | Localized scout name. |
| Create | `server/internal/game/mercenary_rules.go` | Types, loader validation, deterministic offer selection. |
| Modify | `server/internal/game/rules.go` | Add catalog field and loader hook while staying under ratchet. |
| Create | `server/internal/game/main_config_drop_rate.go` | Move existing drop-rate helper out of `rules.go` if needed for ratchet. |
| Modify | `server/internal/game/mercenary_hiring.go` | Select, emit, and spawn selected offer. |
| Modify | `server/internal/game/monster_companion_combat.go` | Emit variant-aware `mercenary_lost`. |
| Modify | `server/internal/game/mercenary_hiring_test.go` | Guard/scout selection, spawn, validation, loss coverage. |
| Modify | `client/scripts/mercenary_panel.gd` | Display scout name in panel/debug proof. |
| Modify | `client/tests/test_mercenary_panel.gd` | Cover scout display helper if needed. |
| Create | `tools/bot/scenarios/96_mercenary_offer_variants.json` | Protocol proof for scout offer. |
| Create | `tools/bot/scenarios/client/69_mercenary_offer_variant_ui.json` | Client UI proof for scout offer. |
| Create during finish | `docs/as-built/v289_mercenary-offer-variants.md` | Record proof and deferred scope. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `server/internal/game/rules.go` — only add the catalog field and loader hook.
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.

Decision:

- [x] Extract mercenary catalog types/validation/selection into `mercenary_rules.go`.
- [x] Move `applyMainConfigDungeonMonsterDropRate` from `rules.go` to
  `main_config_drop_rate.go` if the loader hook would exceed the ratchet allowance.

Verification:

```bash
make maintainability
```

## Task 1 — Shared offer catalog and scout content

Files:

- Create: `shared/rules/mercenaries.v0.json`
- Create: `shared/rules/mercenaries.v0.schema.json`
- Modify: `shared/rules/monsters.v0.json`
- Modify: `shared/assets/monster_visuals.v0.json`
- Modify: `shared/i18n/en.json`
- Modify: `shared/i18n/es.json`
- Modify generated: `shared/assets/model_preview_catalog.v0.json`

- [x] Step 1.1: Add guard and scout offers to a new mercenary catalog.
- [x] Step 1.2: Add `mercenary_scout` as a no-drop/no-XP companion-style monster with distinct
  shared stats.
- [x] Step 1.3: Map scout presentation to an existing in-repo monster asset and add i18n names.
- [x] Step 1.4: Regenerate the model preview catalog.

Verify:

```bash
make model-catalog-generate
make validate-shared
make validate-assets
```

## Task 2 — Server catalog loading and deterministic selection

Files:

- Create: `server/internal/game/mercenary_rules.go`
- Modify: `server/internal/game/rules.go`
- Create if needed: `server/internal/game/main_config_drop_rate.go`
- Modify: `server/internal/game/mercenary_hiring.go`
- Modify: `server/internal/game/monster_companion_combat.go`
- Modify: `server/internal/game/mercenary_hiring_test.go`

- [x] Step 2.1: Load and validate `mercenaries.v0.json` after monster rules are available.
- [x] Step 2.2: Select the board offer deterministically from session seed + board id.
- [x] Step 2.3: Replace hardcoded guard event/spawn fields with selected offer fields while keeping
  the existing configured hire cost.
- [x] Step 2.4: Preserve one active board-hired mercenary by pruning on the existing hire source.
- [x] Step 2.5: Make `mercenary_lost` report the selected offer and monster def for variant hires.
- [x] Step 2.6: Add focused Go tests for catalog validation, stable guard selection, scout selection,
  scout spawn stats, and scout loss event metadata.

Verify:

```bash
(cd server && go test ./internal/game -run 'TestMercenary' -count=1)
```

## Task 3 — Bot proofs

Files:

- Create: `tools/bot/scenarios/96_mercenary_offer_variants.json`
- Create: `tools/bot/scenarios/client/69_mercenary_offer_variant_ui.json`

- [x] Step 3.1: Add a protocol scenario using a pinned scout seed.
- [x] Step 3.2: Assert scout hire events, scout companion entity, and companion-sourced damage.
- [x] Step 3.3: Add a client scenario proving the mercenary panel and companion HUD show the scout
  variant.

Verify:

```bash
make bot scenario=96_mercenary_offer_variants
make bot-client scenario=69_mercenary_offer_variant_ui HEADLESS=1
```

## Task 4 — Client display polish

Files:

- Modify: `client/scripts/mercenary_panel.gd`
- Modify: `client/tests/test_mercenary_panel.gd`

- [x] Step 4.1: Ensure `mercenary_scout` displays as `Mercenary Scout` in offer/status/stats text.
- [x] Step 4.2: Extend the focused panel test if the helper needs explicit coverage.

Verify:

```bash
godot --headless --path client --script res://tests/test_mercenary_panel.gd
```

## Task 5 — Docs and lifecycle

Files:

- Existing: `docs/specs/v289_spec-mercenary-offer-variants.md`
- Existing: `docs/plans/v289_2026-06-19-mercenary-offer-variants.md`
- Create during finish: `docs/as-built/v289_mercenary-offer-variants.md`
- Modify during finish: `PROGRESS.md`

- [x] Step 5.1: Record focused checks, bot proof, and deferred scope in the as-built note.
- [x] Step 5.2: Update lifecycle/current status during finish.

## Task 6 — Final verification

- [x] `make validate-shared`
- [x] `make validate-assets`
- [x] `(cd server && go test ./internal/game -run 'TestMercenary' -count=1)`
- [x] `godot --headless --path client --script res://tests/test_mercenary_panel.gd`
- [x] `make bot scenario=96_mercenary_offer_variants`
- [x] `make bot-client scenario=69_mercenary_offer_variant_ui HEADLESS=1`
- [x] `make maintainability`
