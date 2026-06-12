# v89 As-Built - Class Second Combat Skills

Date: 2026-06-12
Baseline: v88 `skill-visual-rank-seeding`

## What Shipped

- Added `cleave` for Barbarian and `ice_shard` for Sorcerer to the shared skill catalog and presentation catalog.
- Added schema-backed skill contracts for cone attacks, cold projectile slows, and deterministic shard fan-out.
- Made server-owned skill damage use an always-hit path by default, separate from basic attack miss/block/crit resolution.
- Implemented Cleave as a 3-unit, 50-degree cone using weapon damage and deterministic 1-3 unit pushback for surviving targets.
- Implemented Ice Shard as a cold projectile that applies `ice_slow`, stacks movement slow to a 75% cap, and spawns 2-5 deterministic shard projectiles with divided damage.
- Added optional `skill_cast` geometry fields (`position`, `direction`, `range`, `angle_degrees`) to v8 state schemas for client presentation.
- Rendered Cleave as a transient red cone and Ice Shard slow as a light-blue enemy tint driven by authoritative effect IDs.
- Extended dynamic skill visual tooling so `make skill-visual-list` lists all six configured skills, including `cleave` and `ice_shard`.

## Proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'Cleave|IceShard|Skill|MagicBolt'`
- `.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_skill_demo.py tools/bot/test_skill_visual.py -q`
- `make bot scenario=45_class_second_combat_skills.json`
- `make skill-visual-list`
- `make skill-visual skill=ice_shard dry=1`
- `make client-unit`
- `make ci`

## Notes

- The v89 protocol scenario proves the Barbarian Cleave class gate, rank allocation, cast, damage, and cooldown through the normal session path. Cleave push and Ice Shard slow/shatter are covered by focused Go tests and the reusable skill visual runner.
- A standalone `make client-smoke` requires a live server on `localhost:8888`; the direct run reached the live smoke step and failed with `not connected: 4` when no server was running. The same smoke phase passed inside `make ci` with the CI-managed server.
- A broader `scripts/bot_client_local.sh` run was stopped after an unrelated existing `client_combat_feedback` movement timeout; that same scenario passed inside `make ci`.
- `make maintainability` required a documented file-size ratchet exception for the existing sim/rules/client coordinator hotspots touched by this slice, plus the pre-existing `server/internal/http/auth_session_test.go` drift reported by the checker.

## Deferred

- Production VFX/audio for both skills.
- Ground-targeted skill casting.
- Full balance pass for rank scaling and mana/cooldown tuning.
- Recursive shard shattering; shard projectiles damage and slow but do not shatter again.
