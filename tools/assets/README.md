# `tools/assets/` — Python asset validation

Fast, engine-free validation of the asset pipeline (ADR-0006 D5, Python layer):

- `validate_assets.py` — schema-validate the manifest, verify `runtime_path` existence, verify
  `item_visuals.asset_id` resolution, verify character `required_nodes` cover referenced mount
  sockets, verify `provenance.sha256`, and best-effort GLB node-name inspection.
- `test_validate_assets.py` — pytest coverage (happy path + representative failures).

Run via `make validate-assets`. Authoritative runtime socket/visibility checks live in the Godot
headless smoke, not here.
