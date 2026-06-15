# v61 Spec — Rage and Heal Skills

Status: Draft
Date: 2026-06-10
Codename: rage-and-heal-skills

## Purpose

Add two first-tier active skills alongside Magic Bolt:

- Rage: a self buff that costs 10 mana, lasts 30 seconds, increases STR and VIT by 10% at rank 1 plus 10% per extra rank, and scales the character plus equipped gear by the same percentage while active.
- Heal: a ranged area heal that costs 10 mana and heals allied players in the target zone, including the caster when in zone. Rank 1 heals 25% of each target's max HP plus 10% per extra rank.

Skill behavior remains server-authoritative and data-driven. The skill catalog should grow from projectile-only definitions to a closed declarative `effects` model with supported effect types rather than arbitrary JSON plugin method names.

## Non-goals

- No user-authored executable plugin/method dispatch from JSON.
- No new protocol intent field for ground targeting; Heal reuses existing `target_id` or `direction` cast payloads for this slice.
- No balance pass beyond the requested mana costs, requirements, durations, and rank scaling.

## Acceptance Criteria

- `shared/rules/skills.v0.json` contains `magic_bolt`, `rage`, and `heal` in the initial skill-tree row with deterministic rank-scaled requirements.
- Skill schema and Go rule validation allow only supported skill kinds/effect types and reject malformed effects.
- Rage can be learned when STR and VIT meet `10 + 5 * (rank - 1)`, costs 10 mana, starts a cooldown, emits skill/buff events, boosts effective STR/VIT-derived combat stats for 450 ticks, and expires deterministically.
- Rage active presentation scales the local player and equipped visual root by `1 + percent / 100` and returns to normal on expiry or state refresh.
- Heal can be learned when MAGIC meets `10 + 5 * (rank - 1)`, costs 10 mana, starts a cooldown, and heals living allied players in the cast area by `floor(max_hp * percent / 100)` clamped to missing HP.
- Heal emits `player_healed` events with the healed target so the existing green `+N` floating text appears like potion healing.
- The skills panel can display and spend points into all three first-row skills, and the skill bar reflects the selected/right-click skill instead of assuming a single catalog entry.
- Existing Magic Bolt behavior remains covered and unchanged.

## Scope and Files Likely Touched

- Shared rules/assets: `shared/rules/skills.v0.json`, `shared/rules/skills.v0.schema.json`, `shared/assets/skill_presentations.v0.json`, `shared/assets/skill_presentations.v0.schema.json`, relevant golden fixtures.
- Server: `server/internal/game/rules.go`, `server/internal/game/sim.go`, `server/internal/game/handlers.go`, Go tests.
- Client: `client/scripts/skills_panel.gd`, `client/scripts/skill_bar.gd`, `client/scripts/main.gd`, GDScript unit tests.
- Bot/tools: protocol bot scenario for Rage/Heal and client scenario coverage if needed.
- Docs: v61 plan, as-built notes, `PROGRESS.md`.

## Test and Bot Proof

- `make validate-shared` validates the expanded skill schemas and catalogs.
- Go tests cover skill catalog validation, Rage rank scaling/stat effects/expiry, and Heal area ally targeting/clamping.
- Client unit tests cover multi-skill panel selection/spend behavior and skill-bar selected-skill state.
- Bot protocol scenario learns and casts Rage and Heal, waits for buff/heal events, and verifies Magic Bolt still casts.
- `make ci` is green before closeout.

## Open Questions and Risks

- Heal range and radius are not specified by the user; this slice uses conservative content defaults of 9.0 range and 3.0 radius.
- Direction-based Heal centers at max range along the cast direction unless a target entity is supplied. A later slice can add explicit ground-position targeting if playtesting needs finer control.
- Rage active stat effects are runtime-only and not durable progression; requirements continue to use unbuffed base stats.

## Shortcut Decision