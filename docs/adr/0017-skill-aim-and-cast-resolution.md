# ADR-0017: Skill Aim and Cast Resolution (Proposed)

Status: Proposed  
Date: 2026-07-08

## Context

PvP preparation (ADR-0014 D11) requires execution-first combat. v296 sticky targeting applies to
LMB basic attack only. Prior to v459, RMB skills and hotbar casts could send homing `target_id` or
nearest-monster fallback, mimicking tab-target locks and polluting attack-style foe highlights.

Future slices will add skill-dependent travel time and latency rewind; they depend on honest
ground/crosshair aim first.

## Decision

### D1 — Basic attack vs skill aim

- **LMB `action_intent`:** sticky foe lock + health-bar highlight (v296) — unchanged.
- **Skills (RMB, hotbar, directional intents):** aim at click/crosshair world direction via
  `cast_skill_intent.direction`; no homing lock or attack-style highlight for `direction` skills.

### D2 — Shared targeting enum

- `direction` — required for projectile, cone, mobility, and other ground-aim actives.
- `direction_or_target` — reserved for entity-pick skills (`revive` companion corpse).
- Existing `self`, `direction_or_target_area`, `self_or_ally_area` unchanged.

### D3 — Server enforcement

- `direction` skills reject `target_id` with `invalid_targeting`.
- Outcomes remain server-authoritative; client aim is input only.

## Consequences

- Bot `cast_skill` steps with `monster_def_id` resolve to direction for `direction` skills.
- Future travel-time and latency-compensation slices build on `direction` payloads.

## Non-goals (deferred slices)

- Rewind / hitscan at intent time
- Per-skill travel vs instant resolution beyond current ADR-0016 baseline
- PvP-specific scaling or arenas
