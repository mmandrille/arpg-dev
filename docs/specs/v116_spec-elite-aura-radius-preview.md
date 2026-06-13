# v116 Spec: Elite Aura Radius Preview

Status: Approved for implementation from `$autoloop 3`
Date: 2026-06-13
Codename: `elite-aura-radius-preview`

## Purpose

Make the existing elite command aura easier to read by drawing a display-only radius preview around
elite pack leaders when at least one nearby follower currently exposes the server-owned
`elite_command` effect. The server remains the authority for whether a follower is buffed; the
client only visualizes the current relationship already present in entity snapshots.

## Non-goals

- No new aura type, combat tuning, damage formula change, or protocol schema bump.
- No client-side authority over whether a follower is buffed.
- No production VFX/audio, monster nameplates, tooltips, minimap markers, or aura roll table.
- No static/lab monster aura behavior beyond the existing generated-pack command aura.

## Acceptance criteria

- The client renders a subtle radius preview around an elite leader that has at least one visible
  same-pack follower with the `elite_command` effect.
- The preview disappears when no follower currently has the effect, when the leader is not visible,
  or when follower/leader pack metadata is missing.
- Preview radius is read from shared dungeon-generation rules, not hardcoded in client presentation.
- Client debug state exposes the preview count/radius for bot assertions without pixel matching.
- A focused Godot client bot scenario proves the aura marker and radius preview appear in the
  generated dungeon pack scenario.

## Scope and likely files

- Client presentation: `client/scripts/player_status_effect_markers.gd`, `client/scripts/main.gd`
- Client bot: `client/scripts/bot_scenario_runner.gd`
- Bot scenario: `tools/bot/scenarios/client/37_elite_aura_radius_preview.json`
- Docs: `docs/plans/v116_2026-06-13-elite-aura-radius-preview.md`,
  `docs/as-built/v116_elite-aura-radius-preview.md`, `PROGRESS.md`

## Test and bot proof

- `make client-unit`
- `make bot-client scenario=37_elite_aura_radius_preview`
- `make maintainability`
- `make ci`

Manual visual command:

```bash
make bot-visual scenario=37_elite_aura_radius_preview
```

## Plugin / shortcut note

Reject external plugins/assets. This is a small code-native marker layered on existing status-effect
marker helpers and shared rule loading.

## Open questions and risks

| # | Question / risk | Resolution |
|---|-----------------|------------|
| Q-1 | Should the client infer aura state from geometry? | No. It may only use geometry to find the leader for followers already marked with server-owned `elite_command`. |
| R-1 | Generated dungeon placement can be brittle for bot proof. | Reuse the existing v112 deterministic pack seed and assert at least one preview. |
| R-2 | Large Godot coordinator files are already over the ratchet target. | Keep changes narrow and prefer helper extraction if implementation would grow `main.gd` materially. |

## ADR alignment

- ADR-0001: preserves server authority; the client renders state already authored by the server.
- ADR-0007: keeps aura preview as client-only presentation state.

