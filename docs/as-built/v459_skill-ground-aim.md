# v459 — Skill Ground Aim

Spec: [`docs/specs/v459_spec-skill-ground-aim.md`](../specs/v459_spec-skill-ground-aim.md)  
Plan: [`docs/plans/v459_2026-07-08-skill-ground-aim.md`](../plans/v459_2026-07-08-skill-ground-aim.md)

## What it proved

- Directional skills use shared `targeting: direction` and always cast via `cast_skill_intent.direction`, not homing `target_id`.
- LMB sticky attack (v296) unchanged; only `revive` / `direction_or_target` skills keep entity-pick `target_id`.
- Server rejects `target_id` on `direction` skills with `invalid_targeting`.
- Client RMB and hotbar use crosshair aim; direction casts no longer drive attack-style foe health-bar highlight.

## Verification

```bash
make validate-shared
cd server && go test ./internal/game -run 'TestSkillCastDirection|TestReviveSkill' -count=1
make bot scenario=skill_ground_aim_lab
make client-unit
```

Visual check (optional): `make play` — RMB a skill over a monster and confirm no foe lock/highlight; cast still hits if aim line intersects.
