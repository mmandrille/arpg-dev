# v186 Spec: Elite Minion Pack AI

Status: Complete
Date: 2026-06-15

## Goal

Reuse the server-owned follower/assist behavior from companions so elite pack minions act through their elite leader instead of behaving like standalone passive-aggro monsters.

## Requirements

- Elite minions follow their living elite leader when the pack is not engaged.
- Elite minions assist the leader's current target once the leader is chasing.
- Elite minions do not independently passive-aggro nearby players while their leader is idle.
- Pack behavior remains deterministic and server-authoritative.
- Existing elite aura, objective, HUD, and pack metadata behavior continues to work.
- Bot proof shows an elite leader can drive minion assist behavior through combat.

## Non-Goals

- No new elite UI.
- No new elite command skill or player-facing pet command surface.
- No client rendering changes beyond existing monster/aura/objective presentation.
- No persistence or cross-level elite pack state changes.
