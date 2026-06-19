# v291 As Built: Quest Town Turn-In

Date: 2026-06-19
Spec: [`docs/specs/v291_spec-quest-town-turn-in.md`](../specs/v291_spec-quest-town-turn-in.md)
Plan: [`docs/plans/v291_2026-06-19-quest-town-turn-in.md`](../plans/v291_2026-06-19-quest-town-turn-in.md)

## What shipped

- Added `town_quest_giver` as a ready town interactable with service `quest_turn_in`.
- Main gameplay config now owns `quest_turn_in_item_def_id` and `quest_turn_in_reward_gold`, with
  schema, Python validation, and Go load-time validation that the configured item is a quest item.
- Added `quest_turn_in_lab` plus reachable quest-giver placements in default town and vendor lab.
  The new town placements are appended after existing scenario-critical entities to preserve old
  hardcoded unique-chest IDs.
- Server turn-in logic consumes one configured quest item, awards configured gold, emits
  `quest_turn_in_completed`, and updates inventory, gold, and character progression changes.
- Added focused Go tests for success, missing-item rejection, wrong-service rejection, config-owned
  reward behavior, and invalid tuning.
- Added a code-native `QuestSteward` town model and a focused Godot visual test for its scroll and
  quest marker.
- Added protocol and headless client bot scenarios `97_quest_town_turn_in` and
  `75_quest_town_turn_in`.

## Proof

Focused verification:

```bash
make validate-shared
(cd server && go test ./internal/game -run 'QuestTurnIn|MainConfig|Interactable' -count=1)
godot --headless --path client --script res://tests/test_item_visuals.gd
godot --headless --path client --script res://tests/test_quest_giver_visual.gd
make bot scenario=97_quest_town_turn_in
make bot-client scenario=75_quest_town_turn_in HEADLESS=1
make maintainability
```

Result: green on 2026-06-19.

Regression checks after preserving unique-chest IDs:

```bash
make bot scenario=61_purple_town_unique_chest
make bot scenario=82_unique_skill_modifier
make bot scenario=83_second_set_package
make bot scenario=91_unique_non_damage_skill_modifier
```

Result: green on 2026-06-19.

Full verification:

```bash
make ci
```

Result: attempted on 2026-06-19 and red on non-v291 gates: the isolated
`TestEliteMinionFollowsLeaderWithoutPassiveAggro` Go test, protocol scenarios
`88_mercenary_hiring_board`, `90_mercenary_death_loss`, and `95_second_boss_template`, plus older
client combat/boss timing scenarios. The new quest protocol and client scenarios both passed inside
that full sweep.

## Manual visual command

```bash
make bot-visual scenario=75_quest_town_turn_in
```

## Deferred

- Durable quest state, quest offers, quest completion flags, repeat limits, anti-farming rules,
  branching dialog, portraits, VO/audio, and quest-log panel work remain deferred.
- XP rewards, item rewards, account persistence, and broader quest reward economy remain deferred.
