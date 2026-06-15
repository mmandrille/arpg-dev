# v200 As-built: Progress Doc Compaction

## What shipped

- Slimmed `PROGRESS.md` from ~1,592 lines to a ~180-line agent dashboard (current status, ADRs,
  open gaps, checklist).
- Extracted reference archives under `docs/progress/`:
  - `slice-lifecycle.md` — full lifecycle table
  - `slice-codename-index.md` — vN codename lookup
  - `scenario-catalog.md` — bot/smoke scenario inventory
  - `shipped-changelog-archive.md` — preserved historical "Recently closed" prose
- Added `scripts/check-progress-dashboard.sh` to `make maintainability` (250-line cap, no
  "Recently closed" section in `PROGRESS.md`).
- Updated agent skills (`next`, `execute`, `plan`, `finish`, `spec`) for tiered PROGRESS reading.

## Proof

- `./scripts/check-progress-dashboard.sh`
- `make maintainability`

## Follow-up

- Run `$review` at the v200 engineering-review milestone before the next feature batch.
- On `/finish`, update `docs/progress/slice-lifecycle.md` — not inline changelog prose in
  `PROGRESS.md`.
