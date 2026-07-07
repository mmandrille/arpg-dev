# v448 Spec: Quest Steward Hunt Loop

Status: Approved for implementation
Date: 2026-07-07
Codename: `quest-steward-hunt-loop`
Baseline: v447 `class-gear-retune`

## Purpose

Ship the first playable **Quest Steward hunt loop**:

1. Some generated non-boss dungeon floors roll a steward hunt quest.
2. On entering that floor, the client shows a **red hunt banner** (same HUD band as elite objectives).
3. One monster on the floor is flagged as the hunt target; killing it drops a **quest trophy** (heart/head/etc.).
4. Returning to the Quest Steward with the trophy opens **five family choices** (sword, bow, helm, …).
5. Picking a family consumes the trophy and grants one **magic-or-better** rolled item at the **hunt source depth**.

Session-scoped state only; v291 gold `quest_leaf` turn-in remains as legacy/debug.

## Non-goals

- Durable quest persistence, repeat limits, anti-farming, branching dialog, portraits
- Replacing v155 quest-reward chest floors or v158 elite objectives
- Co-op trophy/reward splitting
- Production NPC art or external assets

## Acceptance criteria

- Shared `quest_steward.v0.json` owns hunt floor chance, trophy catalog, family template pools, choice count, and min rarity.
- Hunt floors deterministically flag one monster (`steward_hunt_target`) and store hunt metadata on the level.
- Killing the hunt target drops the configured trophy with `quest_source_depth` metadata.
- `steward_hunt_started` event fires when entering an active hunt floor.
- Quest Steward click with a hunt trophy opens server-generated offers; without trophy rejects `missing_quest_item`.
- `quest_steward_pick_intent` consumes trophy and grants magic+ item at source depth.
- Client red banner + journal/tracker show hunt objective text.
- Client quest steward panel shows five family choices and sends pick intent.
- Go tests + extended protocol/client bot scenarios prove the full loop.

## Scope and files

- Shared: `quest_steward.v0.json`, schema, `items.v0.json`, protocol v8 extensions, `worlds.v0.json` lab
- Server: `dungeon_steward_hunt.go`, `quest_steward.go`, `quest_steward_rewards.go`, `quest_turn_in.go`, handlers, rules load
- Client: `steward_hunt_banner.gd`, `quest_steward_panel.gd`, state sync in focused helpers
- Bot: `120_steward_hunt_quest.json`, `98_steward_hunt_quest.json` (client), extended tier

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game -run 'StewardHunt|QuestSteward' -count=1
make bot scenario=120_steward_hunt_quest
make bot-client scenario=98_steward_hunt_quest HEADLESS=1
make maintainability
```

Manual visual: `make bot-visual scenario=98_steward_hunt_quest`

## Asset decision

- Adopt: Quest Steward model, elite tracker placement, shop/mystery row patterns, draggable window chrome
- Reject: external assets/plugins

## Open questions

- None blocking; defaults from v448 brief apply.
