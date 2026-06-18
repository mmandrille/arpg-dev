# v267 Plan - Perf Debug Mode

Status: Complete
Goal: Add opt-in sampled backend/client perf logs for local lag diagnosis.
Architecture: Keep profiling lightweight and local. Backend sampling reads cheap simulation counts
from `game.Sim`; client sampling reads Godot `Performance` monitors and current client entity
state. `scripts/play.sh` passes `ARPG_PERF_DEBUG` through to both processes.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `server/internal/game/perf_debug.go` | Expose cheap active-level simulation counts |
| Add | `server/internal/realtime/perf_debug.go` | Own backend perf env parsing and structured log helper |
| Modify | `server/internal/realtime/session_loop.go` | Emit sampled `backend_perf` logs |
| Add | `client/scripts/perf_debug_sampler.gd` | Emit sampled `[client-perf]` logs |
| Modify | `client/scripts/main.gd` | Feed frame/entity state to the client sampler |
| Modify | `scripts/play.sh` | Pass `ARPG_PERF_DEBUG` to server and client |

## Tasks

- [x] Step 1: Add backend sim-count snapshot and sampled realtime perf logs.
- [x] Step 2: Add client frame/entity/render sampled perf logs.
- [x] Step 3: Wire `ARPG_PERF_DEBUG` through `make play`.
- [x] Step 4: Update docs and run focused verification.

## Verification

```bash
cd server && go test ./internal/game ./internal/realtime
make client-unit
make maintainability
ARPG_PERF_DEBUG=1 HEADLESS=1 make bot-visual scenario=74_map_transparency_setting
```
