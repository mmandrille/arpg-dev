# v250 Spec - Boss-Specific Telegraph Decals

Status: Complete
Date: 2026-06-17
Codename: boss-specific-telegraph-decals

## Purpose

Improve boss attack readability by replacing the current generic circular telegraph marker with
shape-specific code-native decals. Cave Warden line, cone, summon-circle, and melee-contact warnings
should look and debug differently while still using the existing server-authored telegraph metadata.

## Non-goals

- No server/protocol changes, boss balance changes, timing changes, new boss patterns, imported VFX
  art, audio, particles, or exact aim projection for line/cone decals.
- No production decal asset pipeline. This slice uses lightweight Godot meshes keyed from existing
  `telegraph.hit_shape`, `telegraph.type`, and pattern id.

## Client Asset / Plugin Decision

- **Adopt:** Existing in-repo `BossVisualsController` marker path and boss pattern metadata.
- **Borrow:** Existing code-native mesh/material patterns for client-only presentation.
- **Reject:** External assets/plugins and imported decal textures.

## Acceptance Criteria

- Circle/summon, line, cone, and melee-contact telegraphs create distinct marker shapes.
- Decal color and radius continue to come from server telegraph metadata.
- Existing boss tinting, phase bar, and marker cleanup behavior remain unchanged.
- Entity presentation debug state exposes the current telegraph marker shape.
- Focused Godot tests prove marker shape selection and cleanup.
- A client bot scenario observes Cave Warden line, summon-circle, and cone telegraph decals.

## Scope and Likely Files

- Client: `client/scripts/boss_visuals_controller.gd`, `client/scripts/main.gd`,
  `client/scripts/bot_scenario_runner.gd`.
- Unit tests: `client/tests/test_factories.gd`.
- Bot/scenario: `tools/bot/scenarios/client/66_boss_telegraph_decals.json`.
- Docs: plan, as-built, progress lifecycle.

## Test and Bot Proof

- `godot --headless --path client --script res://tests/test_factories.gd`
- `godot --headless --path client --script res://tests/test_client_bot.gd`
- `make bot-client scenario=66_boss_telegraph_decals.json HEADLESS=1`
- `make maintainability`

## Open Questions and Risks

- No blocking questions.
- Risk: line/cone decals do not receive authoritative aim direction. This slice differentiates shape
  and footprint scale only; exact projected aim remains deferred.
