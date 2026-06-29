# v377 As-Built — Delta Frame Coalesce

## What shipped

- `main.gd` queues incoming `state_delta` payloads per frame and merges them in `_flush_pending_deltas()` before apply.
- `delta_frame_coalesce.gd` owns merge semantics: concatenate events/changes; last `performance` payload wins.
- Reduces duplicate entity refresh work when the server fans out multiple deltas in one client frame.

## Verification

```bash
godot --headless --path client --script res://tests/test_delta_frame_coalesce.gd
make client-unit
```
