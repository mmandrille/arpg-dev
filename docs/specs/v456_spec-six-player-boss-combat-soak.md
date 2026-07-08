# v456 Spec — six player boss combat soak

## Goal

End-to-end proof that v453–v455 stack survives six-peer crowded combat intensity without WebSocket drops or persist-dominated tick overruns.

## Ship

- Extended scenario `six_player_boss_combat_soak` with six simulated co-op peers and staggered skill rotation.
- Custom bot runner `run_six_player_boss_combat_soak` with server perf assertions via `ARPG_BOT_SERVER_LOG`.
- Crowded lab world (`crowded_lightning_perf_probe`, 36 chase monsters) as boss-room density proxy.

## Non-goals

- Production matchmaking; six real Godot clients.
