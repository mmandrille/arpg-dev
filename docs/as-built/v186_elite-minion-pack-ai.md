# v186 As-Built: Elite Minion Pack AI

Date: 2026-06-15

## Shipped

- Added server-side elite minion AI helpers:
  - identify non-leader pack minions
  - resolve their living elite leader
  - copy the leader's chase target when the leader is engaged
  - otherwise follow a deterministic leader-adjacent slot
- Idle elite minions no longer use normal monster passive aggro while a living leader controls the pack.
- Idle elite minions do not opportunistically attack nearby players unless the leader is engaged.
- Existing hit/group aggro remains available so attacked packs still assist deterministically.
- Existing elite aura/objective metadata remains unchanged.
- Bot runtime `attack_until_event` can target `monster_pack_leader` selectors and combat-event assertions can filter source/target pack-leader state.
- Added protocol scenario `77_elite_minion_pack_ai.json`.

## Proof

- `cd server && go test ./internal/game`
- `python -m py_compile tools/bot/run.py`
- `make bot scenario=77_elite_minion_pack_ai.json`
- `make ci`

## Visual Check

```bash
make bot-visual scenario=77_elite_minion_pack_ai.json
```
