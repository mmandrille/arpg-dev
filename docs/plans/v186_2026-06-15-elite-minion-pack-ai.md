# v186 Plan: Elite Minion Pack AI

Status: Complete
Date: 2026-06-15
Spec: `docs/specs/v186_spec-elite-minion-pack-ai.md`

## Adoption Checklist

- Decision: reject new Godot plugin or asset dependency.
- Reason: this slice is server AI behavior and protocol/bot verification only.
- Borrow/adopt: reuse the existing companion follow distances and monster movement/pathing primitives.

## Tasks

- [x] Add helpers to identify elite minions and resolve their living pack leader.
- [x] Route idle elite minion movement toward the leader instead of player passive aggro.
- [x] Copy the leader's chase target to minions when the leader is engaged.
- [x] Prevent idle elite minions from opportunistic player attacks when the leader is not engaged.
- [x] Preserve hit/group aggro so attacked packs still assist deterministically.
- [x] Add Go coverage for idle follow, no passive aggro, and leader-driven assist.
- [x] Add protocol bot proof covering elite objective compatibility and leader-driven combat.
- [x] Update docs/as-built and `PROGRESS.md`.
- [x] Run `make ci`.

## Bot Proof

Scenario: `tools/bot/scenarios/77_elite_minion_pack_ai.json`

Expected flow:

1. Descend to the deterministic elite objective dungeon floor.
2. Assert both an elite leader and non-leader minion are present.
3. Attack an elite leader until combat starts.
4. Verify the pack can damage the player through leader-driven engagement.
5. Verify the elite objective remains locked before the leader is cleared.

Visual verification command:

```bash
make bot-visual scenario=77_elite_minion_pack_ai.json
```
