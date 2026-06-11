# Spec: `paladin-holy-shield`

Status: Complete
Date: 2026-06-11
Codename: `paladin-holy-shield`
Slice: v81 - paladin holy shield
Baseline: v80 `combat-threat-readability`

## Purpose

Add a second Paladin active skill that feels protective in co-op and visibly distinct in the client.
`Holy Shield` is a server-authoritative area ally buff: the Paladin casts around self or a targeted
ally, affected living heroes gain a short armor/block defensive effect, and every affected hero gets
a shining client-side halo while the effect is active.

The skill must also be visible in the same active-effect presentation path as Rage, so players can
see that Holy Shield is running near the skill/hotbar UI instead of relying only on the world halo.

## Non-goals

- No production VFX, audio, particle packs, or external Godot plugin dependency.
- No invulnerability, reflect damage, thorns, taunt, enemy blind, or absorb shield resource.
- No passive skill tree, respec, or new class selection behavior.
- No protocol rewrite; additive event/entity/effect metadata is acceptable if required.
- No final balance pass for Paladin survivability.

## Acceptance Criteria

1. `shared/rules/skills.v0.json` contains Paladin-owned `holy_shield` with rank-scaled requirements,
   mana cost, range/radius, duration, and defensive buff values.
2. Skill schema and Go validation support a closed area ally buff effect; malformed or unsupported
   effect definitions fail validation.
3. Paladins can learn and cast `holy_shield`; non-Paladins are rejected by existing class gates.
4. Casting Holy Shield spends mana, starts cooldown, emits `skill_cast`, emits effect-start events
   for every affected living allied player, and does not affect monsters or dead players.
5. Affected heroes gain server-authoritative defensive benefit for the configured duration, at
   minimum armor and/or block improvement, and the benefit expires deterministically.
6. The effect is represented in authoritative entity state so reconnect/snapshot refresh restores
   the affected heroes' visible effect state while it is active.
7. Godot shows a shining/holy visual around every affected hero and removes it when the effect ends
   or the entity refresh no longer includes the effect.
8. The active effect appears in the same status-effect/hotbar-adjacent presentation used for Rage,
   with label/icon/remaining-time data derived from server events.
9. Existing Rage, Heal, and Magic Bolt behavior remains covered and unchanged.
10. Protocol bot proof learns/casts Holy Shield and verifies effect events plus defensive stat
    improvement/expiry.
11. Client unit or client-bot proof verifies the world shine and status-effect UI state for Holy
    Shield.

## Scope and Likely Files

- Shared rules/assets: `shared/rules/skills.v0.json`, `shared/rules/skills.v0.schema.json`,
  `shared/assets/skill_presentations.v0.json`, validation cross-checks if needed.
- Server: `server/internal/game/rules.go`, `server/internal/game/handlers.go`,
  `server/internal/game/sim.go`, Go tests.
- Protocol examples/schemas only if additive effect metadata needs schema coverage.
- Client: `client/scripts/main.gd`, `client/scripts/status_effects_bar.gd`, a small in-repo holy
  shield effect script if useful, and focused GDScript tests.
- Bot/tools: new protocol scenario and, if stable enough, focused client scenario.
- Docs: v81 plan, as-built, `PROGRESS.md`.

## Test and Bot Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'HolyShield|Rage|Skill'`
- `make client-unit`
- Focused protocol bot scenario for Paladin Holy Shield.
- Focused client proof for Holy Shield shine/status-effect presentation.
- `make ci`

## Open Questions and Risks

- Default tuning: rank 1 uses a compact area around the Paladin, short duration, moderate armor and
  block improvement, and rank scaling on the defensive amount. Exact values are content data, not
  hardcoded assertions.
- Existing skill effect state is mostly local-player/Rage-shaped. The slice may need to generalize
  active skill effects from one global map to per-player effect state; that is in scope because it
  is required for an area ally buff.
- Adoption checklist: reject external plugins/assets. This is small code-native presentation using
  existing client effect/debug patterns, with server authority unchanged.
