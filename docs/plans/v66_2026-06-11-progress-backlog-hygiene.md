# v66 Plan: Progress Backlog Hygiene

## Spec

[`docs/specs/v66_spec-progress-backlog-hygiene.md`](../specs/v66_spec-progress-backlog-hygiene.md)

## File Map

- `PROGRESS.md` — correct stale scenario catalog, recently-closed, deferred backlog, curated
  candidate, lifecycle, and current-status entries.
- `docs/as-built/v66_progress-backlog-hygiene.md` — record what the docs-only slice proved.

## Tasks

- [x] Patch the scenario catalog with v64/v65 scenario IDs.
- [x] Add recently-closed summaries for v64 and v65.
- [x] Correct deferred backlog and curated autoloop candidate status.
- [x] Add v66 lifecycle/current-status/as-built documentation.
- [x] Run docs-focused verification and `make ci`.

## Verification

```bash
rg -n "mystery-seller-paid-reroll|stash-search-and-sorting|paid mystery rerolls|sorting/search|mystery_seller_paid_reroll|stash_search_and_sorting|v66" PROGRESS.md docs/specs docs/plans docs/as-built
git diff --check
make ci
```
