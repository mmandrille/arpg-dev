# `tools/assets/` — Python asset validation

Fast, engine-free validation of the asset pipeline (ADR-0006 D5, Python layer):

- `validate_assets.py` — schema-validate the manifest, verify `runtime_path` existence, verify
  `item_visuals.asset_id` resolution, verify character `required_nodes` cover referenced mount
  sockets, verify `provenance.sha256`, and best-effort GLB node-name inspection.
- `test_validate_assets.py` — pytest coverage (happy path + representative failures).

Run via `make validate-assets`. Authoritative runtime socket/visibility checks live in the Godot
headless smoke, not here.

## Presentation review helpers

- `make skill-logo-sheet` — render the current skill icon presentation metadata into a labeled SVG
  contact sheet. The generator lives at `tools/assets/skill_logo_sheet.py` and mirrors the current
  `client/scripts/skill_icon.gd` placeholder shapes.

Default output:

```bash
make skill-logo-sheet
# writes .artifacts/skill-logo-sheet.svg
```

Override the destination when needed:

```bash
make skill-logo-sheet OUT=/tmp/skill-logo-sheet.svg
```

Use this before improving skill logo metadata in `shared/assets/skill_presentations.v0.json`, so the
review image reflects the same placeholder shapes, colors, labels, and localized skill names the
client currently displays.
