# v170 Plan — Validate Shared Catalog Split

Status: Complete
Goal: Move `main_config` gameplay validation out of `validate_shared.py`.
Architecture: `validate_shared.py` remains the orchestration entrypoint. A focused helper owns
`main_config` gameplay bounds and dungeon monster drop-source resolution checks while preserving
the existing report interface and validation labels.
Tech stack: Python shared validation and maintainability ratchet.

## Baseline and shortcut decision

Builds on v169. No Godot plugin decision is required; this is Python tooling-only maintenance.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Add | `tools/validate_main_config.py` | Main-config gameplay validation helper |
| Modify | `tools/validate_shared.py` | Delegate main-config gameplay checks |
| Modify | `tools/test_validate_shared.py` | Focused helper regression |
| Modify | `.maintainability/file-size-baseline.tsv` | Lower `validate_shared.py` baseline |
| Modify | `PROGRESS.md` | Slice lifecycle and summary |
| Add | `docs/as-built/v170_validate-shared-catalog-split.md` | As-built proof |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `tools/validate_shared.py`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none
- [x] Did every touched grandfathered file stay at or below its baseline?

Decision:
- [x] Extract only the `main_config` gameplay validation domain.

Verification:
```bash
make maintainability
```

## Task 1 — Extract main-config validation

Files:
- Add: `tools/validate_main_config.py`
- Modify: `tools/validate_shared.py`

- [x] Step 1.1: Move `main_config` gameplay validation into a helper using the existing report API.
- [x] Step 1.2: Keep `validate_shared.py` responsible for loading catalogs and preserving later local values.
```bash
make validate-shared
```

## Task 2 — Regression and lifecycle docs

Files:
- Modify: `tools/test_validate_shared.py`
- Modify: `.maintainability/file-size-baseline.tsv`
- Modify: `PROGRESS.md`
- Add: `docs/as-built/v170_validate-shared-catalog-split.md`

- [x] Step 2.1: Add a focused helper regression for invalid drop-rate config.
- [x] Step 2.2: Lower the `validate_shared.py` baseline.
- [x] Step 2.3: Update lifecycle docs and write the as-built note.
```bash
.venv/bin/pytest tools/test_validate_shared.py -q
make maintainability
make ci
```

## Final verification

- [x] `.venv/bin/pytest tools/test_validate_shared.py -q`
- [x] `make validate-shared`
- [x] `make maintainability`
- [x] `make ci`
