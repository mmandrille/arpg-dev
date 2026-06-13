# v113 Spec: Elite Aura Readability

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-13
Codename: `elite-aura-readability`

## Purpose

Make the v112 elite command aura readable to players. Same-pack follower monsters that are currently
buffed by a nearby living pack leader should expose `elite_command` through the existing
`effect_ids` entity field, and the Godot client should render a small display-only marker on those
monsters.

## Non-goals

- No new aura type, aura roll table, combat tuning, damage formula change, or protocol version bump.
- No monster nameplates, tooltips, minimap markers, audio, production VFX, or aura radius preview.
- No client-side decision about whether the aura is active; the server owns the active effect id.

## Acceptance criteria

- Server monster entity views include `effect_ids: ["elite_command"]` only when a non-leader pack
  follower currently benefits from the v112 aura.
- Aura `effect_ids` disappear when the follower is outside radius, its leader is dead, it is itself
  the leader, or it has no pack metadata.
- The Godot client renders a small elite command marker for monsters with the `elite_command`
  effect id and removes it when the id disappears.
- Client debug state exposes the marker so bot/unit tests can assert presentation without pixel
  matching.
- A focused protocol/client proof covers generated dungeon aura readability.

## Scope and likely files

- Server game: `server/internal/game/elite_aura.go`, `server/internal/game/sim.go`,
  `server/internal/game/elite_aura_test.go`
- Client presentation: `client/scripts/player_status_effect_markers.gd`, `client/scripts/main.gd`,
  client tests/bot runner as needed
- Bot: new `tools/bot/scenarios/client/34_elite_aura_readability.json`
- Lifecycle docs: `PROGRESS.md`, `docs/as-built/v113_elite-aura-readability.md`

## Test and bot proof

- `cd server && go test ./internal/game -run TestEliteAura -count=1`
- `make client-unit`
- `make bot-client scenario=34_elite_aura_readability`
- `make maintainability`
- `make ci`

Visual verification command for manual inspection:

```bash
make bot-visual scenario=34_elite_aura_readability
```

## Plugin / shortcut note

Client work is a small code-native marker layered onto the existing status-effect marker helpers.
Do not adopt an external plugin or asset pack for this slice; the plan should record this as
`reject external plugin/assets`.

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | Is a protocol bump required? | No. Reuse the existing `effect_ids` entity field. |
| R-1 | Client could infer aura state incorrectly. | It must render only server-provided `effect_ids`. |
| R-2 | Generated dungeon positioning may make bot proof brittle. | Use the existing generated-pack scenario seed and assert at least one presented marker. |
