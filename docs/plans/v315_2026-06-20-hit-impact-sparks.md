# v315 Plan - Hit Impact Sparks

Status: Complete
Goal: Add small code-native impact spark meshes for existing authoritative damage events.
Architecture: `ImpactSparks` owns mesh/material creation. `main.gd` only routes existing combat
events to the helper after combat text, preserving server authority and avoiding protocol changes.
Tech stack: Godot 4 GDScript helper, focused unit test, existing combat presentation test.

## Baseline and Shortcut Decision

Builds on v314. Asset/plugin decision: adopt code-native mesh/material effects; reject imported VFX
assets, particle plugins, and shader packages.

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `client/scripts/impact_sparks.gd` | Build impact spark nodes from event metadata |
| Modify | `client/scripts/main.gd` | Spawn sparks on damage/kill combat-text paths |
| Create | `client/tests/test_impact_sparks.gd` | Prove helper colors, naming, and child meshes |
| Modify | `client/tests/test_rogue_presentation.gd` | Prove event integration spawns sparks |
| Modify | `scripts/client_smoke.sh` | Register focused unit test |
| Create | `docs/as-built/v315_hit-impact-sparks.md` | Completion proof |

## Maintenance Ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [ ] `client/scripts/main.gd`

Decision:
- [x] Keep `main.gd` edits to narrow event routing and put VFX construction in a new focused helper.

Verification:
```bash
make maintainability
```

## Tasks

- [x] Add the `ImpactSparks` helper.
- [x] Route existing damage/kill event presentation through the helper.
- [x] Add focused and integration coverage.
- [x] Register the focused unit test.
- [x] Update lifecycle docs and as-built proof.

## Verification

- [x] `godot --headless --path client --script res://tests/test_impact_sparks.gd`
- [x] `make client-unit`
- [x] `make maintainability`

Final full `make ci` remains deferred to the enclosing `$autoloop` batch gate.
