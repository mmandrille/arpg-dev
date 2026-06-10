# v48 — Co-op rewards and scaling

**Proves:** Co-op combat can reward nearby party members and scale monster challenge from shared
rules while keeping the server authoritative and protocol v6 unchanged.

- Shared combat rules and golden fixtures define full-XP proximity sharing plus logarithmic party
  scaling with defaults `10.0` radius, `0.25` per doubling, and `0.50` max bonus.
- Monster kill XP now goes once to every eligible alive, connected, same-level player within the
  configured radius; dead, disconnected, far, and different-level players are excluded.
- Private progression, skill progression, and XP events route by explicit owner, so each client sees
  only its own reward and persistence writes the rewarded character.
- Monster HP scales at spawn after rarity/template scaling; monster damage scales at attack
  resolution from the current alive connected same-level party count.
- Reconnect persistence now preserves legitimate tick-zero co-op joins by using an unset
  `joined_tick` sentinel instead of overwriting tick `0`.
- Protocol bot scenario `34_coop_rewards_and_scaling.json` proves nearby shared XP, different-level
  exclusion, replay, and fresh-session persistence; Go replay tests prove shared-XP reconstruction.

**Explicit non-goals:** no loot allocation changes, shared gold, explicit party bonus beyond
HP/damage scaling, monster population-count scaling, party UI, chat, friendly fire, PvP, respawn, or
client UI changes.
