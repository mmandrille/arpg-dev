# v448 — Quest steward hunt loop

## What it proved

- Generated non-boss dungeon floors can roll a **Quest Steward hunt** (~10% via `quest_steward.v0.json`).
- Entering a hunt floor emits `steward_hunt_started` and flags one monster with `steward_hunt_target`.
- Killing the target drops a quest trophy carrying `quest_source_depth`.
- Town Quest Steward click with a hunt trophy opens five deterministic family offers (`quest_steward_offers_opened`).
- `quest_steward_pick_intent` consumes the trophy and grants a magic+ rolled item at source depth.
- Legacy `quest_leaf` gold turn-in remains unchanged.
- Client shows a red hunt banner + journal objective; steward family picker sends pick intent.
- Extended bot proofs: `120_steward_hunt_quest` (protocol), `98_steward_hunt_quest` (client banner/journal).

## Key decisions

- Hunt tuning and trophy/family pools live in `shared/rules/quest_steward.v0.json`.
- Pinned bot seed: `v448_steward_probe_17` (bat wing trophy on depth 1).
- Protocol extended in place on v8 (no v9 bump).

## Deferred

- Co-op trophy/reward splitting, repeat limits, durable quest persistence.
- Full client bot proof of family picker click (banner/journal only in client scenario).

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'StewardHunt|QuestSteward|PinnedSteward' -count=1
make bot scenario=120_steward_hunt_quest
make bot-client SCENARIO=98_steward_hunt_quest HEADLESS=1
```

Visual: `make bot-visual scenario=98_steward_hunt_quest`
