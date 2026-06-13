# v113 As-Built: Elite Aura Readability

**Status:** Complete on `main`

## What shipped

- Active v112 elite command aura state now appears in monster entity `effect_ids` as
  `elite_command`.
- The server emits that id only for non-leader generated-pack followers with a living same-pack
  leader inside aura radius.
- The Godot client renders a compact blue command marker on monsters that carry the effect id.
- Client presentation debug state exposes `has_elite_command_effect` for bot assertions.
- Added client bot scenario `34_elite_aura_readability`.

## Proof

- `cd server && go test ./internal/game -run TestEliteAura -count=1`
- `make client-unit`
- `make bot-client scenario=34_elite_aura_readability`
- `make maintainability`
- `make ci`

## Deferred

Aura radius previews, nameplates/tooltips, audio, production VFX, additional aura types, and aura
roll variety remain deferred.
