# v455 Spec — combat skill processing budget

## Goal

Cap per-tick skill/combat work so overlapping casts cannot blow the 100ms session tick budget. Overflow casts defer deterministically to the next tick (visible sub-tick delay, never silent drop).

## Ship

- Data-driven caps in `shared/rules/combat.v0.json` → `combat_processing`.
- Sim FIFO defer queue for `cast_skill_intent` when `skill_resolutions_per_tick` exhausted.
- Projectile spawn cap hook for remaining simulated projectile skills.
- Unit test `TestSkillBudgetDefersOverflowCastsDeterministically`.
- Extended scenario `crowded_skill_overlap_lab`.

## Non-goals

- True six-client soak (v456).
- Damage-event soft-cap enforcement on every combat event (deferred).
