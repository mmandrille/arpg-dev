# v386 Plan — Skill Weapon Damage Scaling

Status: Ready for implementation
Goal: Projectile skill damage scales from authoritative basic-attack damage via weapon multipliers.
Architecture: Add `weapon_multiplier_range` damage type; centralize resolution in
`skill_weapon_damage.go`; migrate eight projectile skills; update golden to multiplier contract.
Tech stack: shared JSON, Go sim, Python bot, GDScript golden.

## Baseline and shortcut decision

Builds on v385 `gameplay-feel-polish`. No client UI changes; no new assets.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `shared/rules/skills.v0.schema.json` | `weapon_multiplier_range` type |
| Modify | `shared/rules/skills.v0.json` | Migrate 8 projectile skills |
| Modify | `shared/golden/skill_points_and_magic_bolt.*` | Multiplier golden |
| Create | `server/internal/game/skill_weapon_damage.go` | Weapon-based skill damage |
| Modify | `server/internal/game/sim.go` | Use sim skill damage path |
| Modify | `server/internal/game/ranger_skills.go` | Call sim method |
| Modify | `server/internal/game/rules.go` | Validate new damage type |
| Modify | `server/internal/game/game_test.go` | Weapon scaling test |
| Modify | `client/tests/test_golden.gd` | Multiplier parity |
| Create | `tools/bot/scenarios/108_skill_weapon_damage_scaling.json` | E2E proof |

## Maintenance ratchet

- New file `skill_weapon_damage.go` under 600 lines.
- `game_test.go` grandfathered — no growth beyond incidental test additions.

## Task 1 — Shared contracts

- [ ] Add `weapon_multiplier_range` to skills schema
- [ ] Migrate projectile skill damage blocks
- [ ] Update magic bolt golden to multiplier fields

```bash
make validate-shared
```

## Task 2 — Go sim

- [ ] Implement `(*Sim).skillDamageRange`
- [ ] Wire projectile spawn + ranger shots
- [ ] Add `TestSkillProjectileDamageScalesWithWeapon`
- [ ] Fix tests using old package-level `skillDamageRange`

```bash
cd server && go test ./internal/game/... -run 'Skill|MagicBolt|Golden' -count=1
```

## Task 3 — Bot scenario

- [ ] Create `108_skill_weapon_damage_scaling.json` (`ci_tier: extended`)

```bash
make bot scenario=108_skill_weapon_damage_scaling
```

## Task 4 — Client golden

- [ ] Update `_skill_damage_*` helpers for multiplier type

```bash
make client-unit
```

## Task 5 — Lifecycle

- [ ] `docs/as-built/v386_skill-weapon-damage-scaling.md`
- [ ] Lifecycle row + PROGRESS.md

## Final verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run TestSkillProjectileDamageScalesWithWeapon -count=1
make client-unit
make bot scenario=108_skill_weapon_damage_scaling
```
