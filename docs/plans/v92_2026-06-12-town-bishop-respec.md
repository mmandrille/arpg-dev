# v92 Plan — Town Bishop Respec

Status: Complete
Goal: Add a red town bishop NPC that heals on interaction and offers a 250 gold full respec.
Architecture: `town_bishop` is a server-owned interactable service, not a shop. The Go sim owns HP,
mana, gold, stat reset, skill refund, cooldown clearing, and emits protocol events/changes. The
Godot client renders the bishop and shows a display-only service panel that sends one intent.
Tech stack: shared JSON rules and schemas, Go sim/protocol schemas, Python protocol bot, Godot client.

## Baseline and shortcut decision

Builds on v91 `spanish-language-selector` on branch `main`. Reuses the existing interactable,
action-intent, progression, skill progression, gold, cooldown, and service-panel patterns.

Godot plugin/asset checklist:
- License: no external asset/plugin.
- Godot version: in-repo primitive GDScript node only.
- Authoritative boundary: client sends intent; server owns all outcomes.
- Agent ergonomics: text-only GDScript and JSON.
- Maintenance: no new dependency.
- Integration cost: small panel plus primitive model.
- Slice scope decision: reject external plugins/assets; build a minimal red bishop primitive/model.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `docs/specs/v92_spec-town-bishop-respec.md` | Slice contract |
| Modify | `shared/rules/main_config.v0.json` and schema | Data-driven respec cost |
| Modify | `shared/rules/interactables.v0.json` and schema | `town_bishop` service metadata |
| Modify | `shared/rules/worlds.v0.json` | Town/lab bishop placement |
| Modify | `shared/protocol/*.schema.json` | Respec intent and bishop events |
| Modify | `server/internal/game/*.go` | Server authority for bishop heal/respec |
| Modify | `server/internal/game/*_test.go` | Focused bishop/respec unit coverage |
| Modify | `tools/bot/run.py`, `tools/bot/scenarios/*.json` | Protocol proof |
| Modify | `client/scripts/main.gd`, optional helper scene/script | Bishop model/menu and intent |
| Modify | `client/scripts/bot_controller.gd`, client scenarios/tests | Headless client proof |
| Modify | `PROGRESS.md`, `docs/as-built/v92_town-bishop-respec.md` | Lifecycle close-out |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `server/internal/game/game_test.go`
- [x] `tools/bot/run.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: check before editing

Decision:
- [x] Extract focused helper/module/test file where practical. Keep `main.gd` edits minimal and put
      new Go tests in a focused test file if that avoids growing `game_test.go`.
- [x] Documented maintenance exception: this slice updates the grandfathered baselines for
      `client/scripts/main.gd`, `server/internal/game/handlers.go`, `server/internal/game/sim.go`,
      and `tools/bot/run.py` because bishop service integration crosses the existing client,
      sim, handler, and protocol-bot coordinators. New reusable UI/test/scenario code was split
      into focused files where practical.

Verification:
```bash
make maintainability
```

## Task 1 — Shared Rules And Protocol

Files:
- Modify: `shared/rules/main_config.v0.json`
- Modify: `shared/rules/main_config.v0.schema.json`
- Modify: `shared/rules/interactables.v0.json`
- Modify: `shared/rules/interactables.v0.schema.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `shared/protocol/envelope.v*.schema.json`, `shared/protocol/messages.v*.schema.json`, `shared/protocol/session_snapshot.v*.schema.json` as current schema set requires

- [x] Add `services.respec_cost_gold: 250` or equivalent shared config.
- [x] Add `town_bishop` interactable with ready initial state and a bishop service marker.
- [x] Place the bishop in town and in a compact lab world for bot/client proof.
- [x] Add `bishop_respec_intent` payload contract and bishop service/respec event fields.

```bash
make validate-shared
```

## Task 2 — Server Authority

Files:
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/handlers.go`
- Modify: `server/internal/game/sim.go`
- Modify/Create: `server/internal/game/bishop_test.go`

- [x] Load and validate the data-driven respec cost and bishop service interactable.
- [x] On bishop activation, fill HP/mana and emit a service-opened event with cost and affordability.
- [x] Register and implement `bishop_respec_intent`.
- [x] Deduct gold, reset class base stats, refund level-earned stat points, refund spent skill ranks,
      clear skill ranks/cooldowns, refill resources, and emit all related changes/events.
- [x] Reject unaffordable respec with `not_enough_gold` before mutation.
- [x] Add focused tests for heal-on-open, successful respec, unaffordable reject, and class baseline reset.

```bash
cd server && go test ./internal/game/... -run 'TestBishop|TestRules'
```

## Task 3 — Protocol Bot Proof

Files:
- Modify: `tools/bot/run.py`
- Create: `tools/bot/scenarios/45_town_bishop_respec.json`
- Modify: `tools/bot/test_protocol.py` if helper validation needs unit coverage

- [x] Add bot steps to open bishop service, request respec, assert bishop events, assert HP/mana full,
      assert gold decrease, and assert progression/skill reset.
- [x] Add scenario setup that has enough gold, spent stats, and spent skill points before respec.
- [x] Add an unaffordable branch or follow-up session that proves `not_enough_gold`.

```bash
make bot
```

## Task 4 — Godot Client Presentation

Files:
- Modify: `client/scripts/main.gd`
- Modify: `client/scripts/bot_controller.gd`
- Modify/Create: `tools/bot/scenarios/client/??_town_bishop_respec.json` or focused client unit test

- [x] Render `town_bishop` as a red non-merchant NPC/model.
- [x] Add a compact bishop service menu with one `Respec` action.
- [x] Send `bishop_respec_intent` from the menu and reflect server events/status.
- [x] Add headless client proof that clicking the bishop opens the service menu and exposes Respec.

```bash
make client-unit
```

## Task 5 — Lifecycle Docs And CI

Files:
- Modify: `docs/plans/v92_2026-06-12-town-bishop-respec.md`
- Modify: `docs/specs/v92_spec-town-bishop-respec.md`
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v92_town-bishop-respec.md`

- [x] Mark plan tasks complete as they pass.
- [x] Update spec status and lifecycle docs.
- [x] Add as-built summary and any deferred scope.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make maintainability`
- [x] `make validate-shared`
- [x] `cd server && go test ./internal/game/... ./internal/inputdecode/...`
- [x] `make bot scenario=town_bishop_respec`
- [x] `make client-unit`
- [x] `make ci`

Manual visual check:
```bash
make bot-visual scenario=45_town_bishop_respec.json
```
