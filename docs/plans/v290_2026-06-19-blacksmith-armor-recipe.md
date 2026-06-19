# v290 Plan — Blacksmith Armor Recipe

Status: Complete
Goal: Add a third blacksmith recipe that reinforces armor through the existing authoritative upgrade
flow.
Architecture: Keep upgrade outcomes server-authoritative and reuse the current gold/resource,
success, pity, and item-level mechanics. Add only recipe identity and eligibility rules: the server
owns acceptance, while the client mirrors eligibility for preview and button state. Extract client
recipe metadata into a focused helper before expanding options so `blacksmith_panel.gd` stays under
the maintainability ratchet.
Tech stack: Go HTTP/store integration, Godot blacksmith panel/helper, client bot scenario, SDD docs.

## Baseline and shortcut decision

Builds on v238 recipe selector, v245 `weapon_honing`, v246 upgrade history, and ADR-0012. The slice
adds `armor_reinforcement` without changing upgrade resources, costs, success formulas, or the
generic upgrade mutation algorithm.

Asset/plugin decision:

- Adopt: existing blacksmith panel, upgrade preview/history, item icon presentation, vendor-lab loot,
  and inventory upgrade route.
- Borrow: existing weapon-honing HTTP test and `blacksmith_second_recipe` client bot structure.
- Reject: external assets/plugins, new art/icons/audio, new model pipelines, per-recipe tuning, and a
  full shared blacksmith recipe catalog.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `server/internal/http/account_stash.go` | Add armor recipe ID, validation, and eligibility. |
| Modify | `server/internal/http/blacksmith_recipe_test.go` | Cover armor success and weapon rejection. |
| Create | `client/scripts/blacksmith_recipes.gd` | Focused recipe metadata and eligibility helper. |
| Modify | `client/scripts/blacksmith_panel.gd` | Delegate recipe options/labels/eligibility to helper. |
| Modify | `client/tests/test_blacksmith_panel.gd` | Cover third option and armor/weapon gating. |
| Create | `tools/bot/scenarios/client/70_blacksmith_armor_recipe.json` | Client proof for selecting and using armor recipe. |
| Create during finish | `docs/as-built/v290_blacksmith-armor-recipe.md` | Record proof and deferred scope. |
| Modify during finish | `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/progress/slice-codename-index.md` | Lifecycle updates. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:

- [x] `client/scripts/blacksmith_panel.gd` — currently at the 600-line target; must shrink before
  adding behavior.
- [x] `client/scripts/bot_scenario_runner.gd` — not expected to change.
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected.
- [x] Did every touched grandfathered file stay at or below its baseline? Not applicable unless bot
  coordinator changes become necessary.

Decision:

- [x] Extract recipe constants, options, labels, eligibility, and rejection copy into
  `blacksmith_recipes.gd`.
- [x] Keep server recipe changes inside `account_stash.go`, which is below the target.

Verification:

```bash
make maintainability
```

## Task 1 — Server armor recipe eligibility

Files:

- Modify: `server/internal/http/account_stash.go`
- Modify: `server/internal/http/blacksmith_recipe_test.go`

- [x] Step 1.1: Add `armor_reinforcement` as a valid recipe ID.
- [x] Step 1.2: Add server eligibility for armor-bearing templates: positive armor base stat and
  armor slots (`off_hand`, `head`, `chest`, `gloves`, `belt`, `boots`).
- [x] Step 1.3: Extend the focused HTTP test to prove shield/mail armor can be upgraded and weapons
  are rejected by the armor recipe.

Verify:

```bash
(cd server && go test ./internal/http -run Blacksmith -count=1)
```

## Task 2 — Client recipe helper and panel gating

Files:

- Create: `client/scripts/blacksmith_recipes.gd`
- Modify: `client/scripts/blacksmith_panel.gd`
- Modify: `client/tests/test_blacksmith_panel.gd`

- [x] Step 2.1: Extract recipe IDs, option metadata, label lookup, eligibility text, item acceptance,
  and rejection message into `blacksmith_recipes.gd`.
- [x] Step 2.2: Add `armor_reinforcement` / `Reinforce Armor` / `Eligible: Armor pieces only`.
- [x] Step 2.3: Delegate blacksmith panel recipe selector, preview, button gating, and rejection copy
  to the helper.
- [x] Step 2.4: Extend the focused panel test to assert three options, armor recipe label/eligibility,
  armor item enabled, and weapon disabled.

Verify:

```bash
godot --headless --path client --script res://tests/test_blacksmith_panel.gd
```

## Task 3 — Client bot proof

Files:

- Create: `tools/bot/scenarios/client/70_blacksmith_armor_recipe.json`

- [x] Step 3.1: Reuse `vendor_lab` to pick up an upgrade shard and enough sellable loot for the
  existing upgrade cost.
- [x] Step 3.2: Select `armor_reinforcement`, stage `cave_mail`, and assert recipe-specific preview
  text.
- [x] Step 3.3: Upgrade once and assert the armor item is upgraded and the wallet resource is spent.

Verify:

```bash
make bot-client scenario=70_blacksmith_armor_recipe HEADLESS=1
```

## Task 4 — Docs and lifecycle

Files:

- Existing: `docs/specs/v290_spec-blacksmith-armor-recipe.md`
- Existing: `docs/plans/v290_2026-06-19-blacksmith-armor-recipe.md`
- Create during finish: `docs/as-built/v290_blacksmith-armor-recipe.md`
- Modify during finish: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/progress/slice-codename-index.md`

- [x] Step 4.1: Record focused checks, client bot proof, visual command, and deferred scope in the
  as-built note.
- [x] Step 4.2: Update lifecycle/current status during finish.

## Task 5 — Final verification

- [x] `(cd server && go test ./internal/http -run Blacksmith -count=1)`
- [x] `godot --headless --path client --script res://tests/test_blacksmith_panel.gd`
- [x] `make bot-client scenario=70_blacksmith_armor_recipe HEADLESS=1`
- [x] `make maintainability`

Full `make ci` is deferred to the end of the selected `$autoloop` queue.

Manual visual command:

```bash
make bot-visual scenario=70_blacksmith_armor_recipe
```
