# v294 Plan: Full-CI Residual Stabilization

## Status

Complete after final `make ci` proof.

## Tasks

1. Stabilize the elite-minion follow test.
   - Isolate the synthetic lab from generated dungeon walls.
   - Place the leader/minion on reachable geometry and assert movement toward the same follow goal
     used by production AI.
   - Keep no-passive-aggro and leader-assist assertions intact.

2. Make mercenary protocol proofs variant-tolerant.
   - Remove stale `mercenary_guard` locks where the selected offer may now vary.
   - Continue asserting hired companion presence and death-loss behavior.

3. Fix click-driven client bot event matching.
   - Treat click-target selectors as target selection criteria, not pending event filters.
   - Allow `monster_def_id` expectations to match source, target, or generic event fields.

4. Refresh stale client scenario contracts.
   - Make broad boss health/phase UI scenarios boss-template agnostic.
   - Make mercenary roster/combat stats scenarios selected-offer tolerant.
   - Keep the boss telegraph decal scenario pinned to a deterministic Cave Warden seed for
     shape-specific coverage.

5. Stabilize replay-tail proof for the second boss template.
   - Drain one tick after the boss kill so recorded output includes the same summoned-add tail
     event that deterministic replay derives.
   - Improve replay event-count mismatch diagnostics to report the first extra or missing event.

6. Verify and document.
   - Run focused Go, protocol bot, and failed client bot scenarios.
   - Run full client bot and Godot bot unit coverage.
   - Run final `make ci`.
   - Update `PROGRESS.md`, lifecycle, and as-built docs, then commit.

## Bot Scenarios

- `make bot scenario=mercenary_hiring_board`
- `make bot scenario=mercenary_death_loss`
- `make bot scenario=second_boss_template`
- `make bot-client scenario=26_boss_health_bar_ui.json`
- `make bot-client scenario=28_boss_phase_readability.json`
- `make bot-client scenario=47_mercenary_roster_ui.json`
- `make bot-client scenario=64_mercenary_combat_stats.json`
- `make bot-client scenario=66_boss_telegraph_decals.json`

For visual verification of the retained Cave Warden decal proof:

```sh
make bot-visual scenario=66_boss_telegraph_decals.json
```
