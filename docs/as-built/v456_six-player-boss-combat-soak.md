# v456 As-Built — six-player-boss-combat-soak

Date: 2026-07-08  
Spec: [`docs/specs/v456_spec-six-player-boss-combat-soak.md`](../specs/v456_spec-six-player-boss-combat-soak.md)

## Shipped

- Extended scenario `six_player_boss_combat_soak`: 6 co-op peers, Volley/Magic Bolt/Lightning rotation, 8 cast cycles.
- `run_six_player_boss_combat_soak` + `assert_soak_server_perf` in `tools/bot/run.py`.
- `scripts/bot_local.sh` exports `ARPG_BOT_SERVER_LOG` for perf assertions.

## Verification

```bash
ARPG_PERF_DEBUG=1 make bot scenario=six_player_boss_combat_soak
```

## Notes

- Uses `crowded_lightning_perf_probe` (36 chasers) as compact boss-room density stand-in; authoritative boss-floor soak can follow with true boss template when needed.
