# v116 Plan — Elite Aura Radius Preview

Status: Complete
Goal: Render a display-only command-aura radius preview for visible elite leaders with currently
buffed followers.
Architecture: Reuse the existing server-owned `elite_command` follower effect as the only authority
for active aura state. The client computes which visible leader to decorate from entity metadata and
loads the radius from shared dungeon-generation rules. Bot assertions use debug state instead of
pixel matching.
Tech stack: Godot client, shared JSON loader, Python/Godot client bot, docs.

## Baseline and shortcut decision

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/player_status_effect_markers.gd` | Add reusable elite aura radius preview marker/debug helper. |
| Create | `client/scripts/elite_aura_preview_sync.gd` | Keep leader/follower preview sync out of the large coordinator. |
| Modify | `client/scripts/main.gd` | Sync leader radius previews from visible server-authored follower effect ids. |
| Modify | `client/scripts/bot_scenario_runner.gd` | Assert elite aura preview debug fields. |
| Create | `tools/bot/scenarios/client/37_elite_aura_radius_preview.json` | Focused client bot proof. |
| Create | `docs/as-built/v116_elite-aura-radius-preview.md` | As-built proof summary. |
| Modify | `PROGRESS.md` | Mark v116 complete and carry deferred scope. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] `client/scripts/bot_scenario_runner.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [x] Defer extraction with rationale: this slice only wires a narrow presentation sync and
  assertion into already-grandfathered client files; extracting during the visual proof would add
  more risk than the small delta.

Verification:
```bash
make maintainability
```

## Task 1 — Client Aura Preview

Files:
- Modify: `client/scripts/player_status_effect_markers.gd`
- Create: `client/scripts/elite_aura_preview_sync.gd`
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Add a code-native radius preview node with active-count/radius debug helpers.
- [x] Step 1.2: Load the shared elite aura radius and sync previews to visible pack leaders with
  server-marked followers.
```bash
make client-unit
```

## Task 2 — Bot Scenario

Files:
- Modify: `client/scripts/bot_scenario_runner.gd`
- Create: `tools/bot/scenarios/client/37_elite_aura_radius_preview.json`

- [x] Step 2.1: Add bot assertions for elite command radius preview count/radius.
- [x] Step 2.2: Add a generated dungeon scenario that verifies marker plus radius preview.
```bash
make bot-client scenario=37_elite_aura_radius_preview
```

## Task 3 — Lifecycle Docs and CI

Files:
- Modify: `PROGRESS.md`
- Create: `docs/as-built/v116_elite-aura-radius-preview.md`

- [x] Step 3.1: Record completed slice, test proof, and deferred production VFX/audio/nameplates.
```bash
make maintainability
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make bot-client scenario=37_elite_aura_radius_preview`
- [x] `make maintainability`
- [x] `make ci`
