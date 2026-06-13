# v138 As-Built: CODEMAP & Reduction Ratchet

Date: 2026-06-13
Status: Complete

## What Changed

- Added `docs/CODEMAP.md`, a domain-to-files index for loading focused context before broad grep or
  coordinator reads.
- Added `tools/validate_codemap.py` and `tools/test_validate_codemap.py`; `make validate-shared`
  now validates that every CODEMAP path exists.
- Upgraded `scripts/check-file-size-ratchet.sh` with `ROOT`/`BASELINE` test overrides, a lower-bound
  ratchet that forces baselines down after file shrinkage, and a grandfathered-file trend line.
- Added `tools/test_file_size_ratchet.py` for upper-bound, lower-bound, and within-allowance cases.
- Made `make ci` depend on `maintainability`, so the ratchet is now a CI gate.
- Updated `CLAUDE.md`, `AGENTS.md`, and `skills/plan/SKILL.md` with CODEMAP and reduction-ratchet
  policy, including touch-to-shrink and new-domain guidance.

## Proof

- `.venv/bin/python -m pytest tools/test_file_size_ratchet.py -q`
- `.venv/bin/python -m pytest tools/test_validate_codemap.py -q`
- `make validate-shared`
- `make maintainability`
- `make ci`

## Notes

- No gameplay, protocol, replay, persistence, or client presentation behavior changed.
- The baseline TSV was refreshed to exact current counts so the new lower-bound rule starts from the
  current repository state.
- Downstream coordinator splits and market expiration freshness remain deferred roadmap slices.
