# v378 As-Built — Recon Backpressure

## What shipped

- `shared/rules/main_config.v0.json` adds `client_perf.reconciliation_backpressure_threshold`.
- `reconciliation_backpressure.gd` clears client pending action targets when reconciliation delta exceeds the threshold.
- `main.gd` invokes backpressure after authoritative position reconcile; also clears charge channel visuals.

## Verification

```bash
godot --headless --path client --script res://tests/test_reconciliation_backpressure.gd
make client-unit
```
