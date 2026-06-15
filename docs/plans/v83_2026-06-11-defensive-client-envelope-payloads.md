# v83 Plan - Defensive Client Envelope Payloads

Status: Ready for implementation
Goal: Harden Godot client envelope payload access so malformed messages do not crash `_handle_message`.
Architecture: Keep the server protocol unchanged. The client message boundary converts missing or
non-dictionary payloads to an empty dictionary before dispatch, matching the defensive access style
already used by delta application.
Tech stack: Godot GDScript client and headless client unit tests.

## Baseline and shortcut decision

this is defensive parsing in existing GDScript, not UI, camera, inventory presentation, or art.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/main.gd` | Use defensive payload dictionary access in `_handle_message`. |
| Modify | `client/tests/test_delta_apply.gd` | Add malformed envelope unit coverage. |
| Add | `docs/as-built/v83_defensive-client-envelope-payloads.md` | Summarize shipped behavior. |
| Modify | `PROGRESS.md` | Mark v83 complete and close the review finding. |

## Maintenance ratchet

Target: source/test/tool files stay at or below 600 lines.

Hotspot / over-limit files touched:
- [x] `client/scripts/main.gd`
- [x] Other over-limit file from `.maintainability/file-size-baseline.tsv`: none expected

Decision:
- [x] Defer extraction with rationale: this is a narrow defensive edit at the message boundary.
  Extracting from `main.gd` belongs to the separate combat event presenter recommendation.

Verification:
```bash
make maintainability
```

## Task 1 - Defensive payload access

Files:
- Modify: `client/scripts/main.gd`

- [x] Step 1.1: Add a tiny helper or local guard that returns `{}` unless `env.get("payload")` is a dictionary.
- [x] Step 1.2: Route snapshot, delta, accepted, rejected, and error handling through the guarded payload.
- [x] Step 1.3: Preserve existing debug strings and pending-action cleanup behavior.

```bash
make client-unit
```

## Task 2 - Malformed envelope unit proof

Files:
- Modify: `client/tests/test_delta_apply.gd`

- [x] Step 2.1: Add a test that calls `_handle_message` with accepted/rejected/error envelopes missing payload.
- [x] Step 2.2: Add a test for non-dictionary payloads.
- [x] Step 2.3: Keep tests scene-tree-free on `MainScript.new()`.

```bash
make client-unit
```

## Task 3 - Lifecycle docs and CI

Files:
- Modify: `docs/specs/v83_spec-defensive-client-envelope-payloads.md`
- Modify: `docs/plans/v83_2026-06-11-defensive-client-envelope-payloads.md`
- Add: `docs/as-built/v83_defensive-client-envelope-payloads.md`
- Modify: `PROGRESS.md`

- [x] Step 3.1: Mark spec and plan complete.
- [x] Step 3.2: Add as-built and progress updates.
- [x] Step 3.3: Run final verification.

```bash
make maintainability
make ci
```

## Final verification

- [x] `make client-unit`
- [x] `make maintainability`
- [x] `make ci`

Deferred scope: combat event presenter extraction and broader `main.gd` decomposition.
