# v291 Plan — Quest Town Turn-In

Status: Complete
Goal: Add a reachable town quest giver that consumes `Quest Leaf` for a shared-config gold reward.
Architecture: Reuse `action_intent` on a ready town service interactable. Shared rules own the
service definition, world placement, required quest item, and reward amount. Server code remains
authoritative for inventory removal and gold updates; client work is limited to showing the new town
NPC/interactable in existing code-native presentation.
Tech stack: Shared rule catalogs, Go sim tests, Godot town presentation test, protocol/client bot
scenarios, SDD docs.

## Baseline and shortcut decision

Builds on v155 random quest reward floors, v174 quest journal foundation, and the existing
town-service patterns for vendors, bishop, blacksmith, market board, and mercenary board. This slice
adds a one-click turn-in service rather than a full quest system.

Asset/plugin decision:

- Adopt: existing `quest_leaf`, static town interactables, `action_intent`, inventory/gold changes,
  and bot event assertions.
- Borrow: bishop/mercenary service target resolution and vendor-lab scenario structure.
- Reject: external assets/plugins, production NPC art, dialog/portrait UI, new audio, durable quest
  state, repeat limits, XP/item rewards, and a quest panel.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/main_config.v0.json` | Add turn-in item and reward tuning. |
| Modify | `shared/rules/main_config.v0.schema.json` | Validate new required config fields. |
| Modify | `shared/rules/interactables.v0.json` | Add `town_quest_giver` service. |
| Modify | `shared/rules/worlds.v0.json` | Place quest giver in default town and vendor lab. |
| Modify | `server/internal/game/rules.go` | Load/validate quest turn-in config. |
| Modify | `server/internal/game/main_config_validation.go` | Reject invalid quest reward tuning. |
| Create | `server/internal/game/quest_turn_in.go` | Resolve service target and consume/reward. |
| Create | `server/internal/game/quest_turn_in_test.go` | Cover success and rejection paths. |
| Modify | `client/scripts/town_node_factory.gd` | Render quest giver with existing primitives. |
| Create | `client/tests/test_quest_giver_visual.gd` | Cover the quest giver node. |
| Create | `tools/bot/scenarios/97_quest_town_turn_in.json` | Protocol proof. |
| Create | `tools/bot/scenarios/client/75_quest_town_turn_in.json` | Client proof. |
| Create during finish | `docs/as-built/v291_quest-town-turn-in.md` | Record proof and deferred scope. |
| Modify during finish | `PROGRESS.md`, `docs/progress/slice-lifecycle.md`, `docs/progress/slice-codename-index.md` | Lifecycle updates. |

## Maintenance ratchet

Target: new source/test/tool files stay at or below 600 lines. Grandfathered files must not exceed
baseline + allowance.

Hotspot / over-limit files touched:

- [x] `server/internal/game/rules.go` is grandfathered but currently below its baseline; keep the
  config-field addition small.
- [x] Avoid touching `server/internal/game/sim.go`; it is already at its allowed line-count ceiling.
- [x] Keep turn-in logic in `quest_turn_in.go`, not `handlers.go` or `interactables.go`.
- [x] Avoid `tools/bot/run.py`; existing protocol/client bot actions covered the flow.

Verification:

```bash
make maintainability
```

## Task 1 — Shared quest giver and reward config

Files:

- Modify: `shared/rules/main_config.v0.json`
- Modify: `shared/rules/main_config.v0.schema.json`
- Modify: `shared/rules/interactables.v0.json`
- Modify: `shared/rules/worlds.v0.json`
- Modify: `server/internal/game/rules.go`
- Modify: `server/internal/game/main_config_validation.go`

- [x] Step 1.1: Add `quest_turn_in_item_def_id` and `quest_turn_in_reward_gold` to main gameplay
  config and schema.
- [x] Step 1.2: Validate non-empty known quest item and non-negative reward in server/shared checks.
- [x] Step 1.3: Add `town_quest_giver` with `service: "quest_turn_in"` and place it in reachable
  town/vendor-lab positions.

Verify:

```bash
make validate-shared
```

## Task 2 — Server turn-in behavior

Files:

- Create: `server/internal/game/quest_turn_in.go`
- Create: `server/internal/game/quest_turn_in_test.go`

- [x] Step 2.1: Resolve quest giver service targets on current town level and in dispatch range.
- [x] Step 2.2: On click with the required item, remove one inventory item, add configured gold,
  append inventory/gold/progression changes, and emit `quest_turn_in_completed`.
- [x] Step 2.3: Reject missing item, wrong service target, out-of-range, and non-ready target cases
  with stable reasons.
- [x] Step 2.4: Prove the reward follows `MainConfig.Gameplay.QuestTurnInRewardGold`.

Verify:

```bash
(cd server && go test ./internal/game -run 'QuestTurnIn|MainConfig|Interactable' -count=1)
```

## Task 3 — Client town presentation

Files:

- Modify: `client/scripts/town_node_factory.gd`
- Create: `client/tests/test_quest_giver_visual.gd`

- [x] Step 3.1: Add a code-native quest giver node using existing primitive material helpers.
- [x] Step 3.2: Add it to the town preview service list and interactable factory.
- [x] Step 3.3: Test that `town_quest_giver` creates the expected marker/node.

Verify:

```bash
godot --headless --path client --script res://tests/test_item_visuals.gd
godot --headless --path client --script res://tests/test_quest_giver_visual.gd
```

## Task 4 — Bot proof

Files:

- Create: `tools/bot/scenarios/97_quest_town_turn_in.json`
- Create: `tools/bot/scenarios/client/75_quest_town_turn_in.json`

- [x] Step 4.1: Reuse the existing protocol bot action/event assertions to click
  `town_quest_giver`, wait for `quest_turn_in_completed`, and check the gold delta.
- [x] Step 4.2: Add protocol proof that picks up `Quest Leaf`, turns it in, loses the item, and gains
  gold.
- [x] Step 4.3: Add client proof for the same visible town interaction.

Verify:

```bash
make bot scenario=97_quest_town_turn_in
make bot-client scenario=75_quest_town_turn_in HEADLESS=1
```

Manual visual command:

```bash
make bot-visual scenario=75_quest_town_turn_in
```

## Task 5 — Docs and lifecycle

Files:

- Existing: `docs/specs/v291_spec-quest-town-turn-in.md`
- Existing: `docs/plans/v291_2026-06-19-quest-town-turn-in.md`
- Create during finish: `docs/as-built/v291_quest-town-turn-in.md`
- Modify during finish: `PROGRESS.md`, `docs/progress/slice-lifecycle.md`,
  `docs/progress/slice-codename-index.md`

- [x] Step 5.1: Record focused checks, client visual command, and deferred quest-system scope in the
  as-built note.
- [x] Step 5.2: Update current status and lifecycle after verification.

## Task 6 — Final verification

- [x] `make validate-shared`
- [x] `(cd server && go test ./internal/game -run 'QuestTurnIn|MainConfig|Interactable' -count=1)`
- [x] `godot --headless --path client --script res://tests/test_item_visuals.gd`
- [x] `godot --headless --path client --script res://tests/test_quest_giver_visual.gd`
- [x] `make bot scenario=97_quest_town_turn_in`
- [x] `make bot-client scenario=75_quest_town_turn_in HEADLESS=1`
- [x] `make maintainability`

Full `make ci` was attempted at the end of the selected `$autoloop` queue and remains red on
non-v291 failures: `TestEliteMinionFollowsLeaderWithoutPassiveAggro`, `88_mercenary_hiring_board`,
`90_mercenary_death_loss`, `95_second_boss_template`, and older client combat/boss timing scenarios.
