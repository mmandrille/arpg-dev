# v335 — Movement input presenter extraction

**Status:** Complete  
**Codename:** movement-input-presenter

## What it proved

- `client/scripts/movement_input_presenter.gd` owns keyboard move intent dispatch, walk linger, force-stand stop intent, and bot-facing `intent_starts_motion` helpers.
- `main.gd` delegates movement-intent orchestration through the presenter while keeping thin wrappers for bot compatibility.
- `client/tests/test_movement_input_presenter.gd` covers intent classification and walk linger state.

## Verification

```bash
godot --headless --path client --script res://tests/test_movement_input_presenter.gd
make maintainability
```

## Deferred

- Further attack-move / sticky-target cluster extraction from `main.gd`.
