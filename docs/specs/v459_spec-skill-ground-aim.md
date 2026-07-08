# v459 Spec: Skill Ground Aim

Status: Complete  
Date: 2026-07-08  
Codename: `skill-ground-aim`  
Baseline: v457 `live-combat-transport-stability`

## Purpose

Establish a durable skill-aim rule for PvE and future PvP: directional skills always aim at the
player's click/crosshair point (world direction), never at a homing foe lock or attack-style target
highlight. LMB basic attack keeps v296 sticky targeting unchanged.

Encode the rule in shared skill targeting data (`direction`) and enforce it on client input and
server cast validation.

## Non-goals

- Latency compensation / rewind-at-intent-time
- Skill travel-time differentiation (volley instant vs snipe/fireball travel + lead)
- PvP arenas, friendly fire, or PvP damage scaling
- Protocol version bump (additive rules + reject reason only)
- Removing `target_id` from `cast_skill_intent` globally (`revive` keeps entity pick)

## Acceptance criteria

- RMB cast over a living monster sends `direction` only, not `target_id`
- Skill-bar / hotkey cast uses mouse/crosshair aim, not nearest-monster fallback, for `direction` skills
- Directional skill casts do not drive attack-style foe health-bar highlight via `pending_skill_casts`
- `shared/rules/skills.v0.json` uses `targeting: direction` for projectile/cone/mobility skills; `revive` keeps `direction_or_target`
- Server rejects `target_id` on `direction` skills with `invalid_targeting`
- Protocol bot proves direction cast damages a lined-up target and `target_id` is rejected on `direction` skills
- Client unit test proves RMB and hotbar payloads for `magic_bolt` omit `target_id`
- `78_attack_move_sticky_targeting` client regression unchanged (LMB only)

## Scope and files

| Area | Files |
|------|-------|
| ADR | `docs/adr/0017-skill-aim-and-cast-resolution.md` |
| Shared | `shared/rules/skills.v0.json`, `shared/rules/skills.v0.schema.json` |
| Server | `server/internal/game/sim.go`, `skill_rules_validation.go`, tests |
| Client | `client/scripts/skill_aim_input.gd`, `main.gd`, `enemy_health_bar_visibility.gd`, `client/tests/test_coop_client.gd` |
| Bot | `tools/bot/run.py`, `tools/bot/scenarios/skill_ground_aim_lab.json` |
| Docs | lifecycle, as-built, `PROGRESS.md` |

## Test and bot proof

```bash
make validate-shared
cd server && go test ./internal/game/... -run SkillCast -count=1
make bot scenario=skill_ground_aim_lab
make client-unit
```

## Asset and plugin decision

- Adopt: existing aim helpers (`_aim_direction_from_mouse`, `DirectionalAttackInput`, `SkillRulesLoader`)
- Borrow: `skill_progression_lab` world for protocol proof
- Reject: external input plugins or new reticle art in this slice

## Open questions

None — batch defaults from `/next` brief apply.
