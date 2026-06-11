# v66 Spec: Progress Backlog Hygiene

## Status

Approved for implementation in the current autoloop.

## Context

`PROGRESS.md` is the canonical baseline for future agents. After v64 and v65 shipped,
the lifecycle table was correct, but some discovery sections still listed those slices
as open or deferred. That makes the next autoloop pass noisy and can cause duplicate
slice proposals.

## Goals

- Mark v64 `mystery-seller-paid-reroll` and v65 `stash-search-and-sorting` as completed
  everywhere `PROGRESS.md` presents curated candidate status.
- Add concise recently-closed summaries for v64 and v65 so future agents can understand
  what shipped without opening the full specs first.
- Add the v64/v65 bot scenario IDs to the scenario catalog.
- Remove or narrow stale deferred backlog text that still treats paid rerolls and stash
  search/sort as unshipped work.
- Record the docs-only result in an as-built note.

## Non-goals

- No server, client, shared contract, bot, or test behavior changes.
- No new feature prioritization beyond correcting stale progress metadata.
- No engineering review; the next review remains due at v70.

## Acceptance Criteria

- `PROGRESS.md` no longer lists `mystery-seller-paid-reroll` or `stash-search-and-sorting`
  as open candidates.
- Deferred backlog language keeps still-open adjacent work, such as stash tabs/capacity,
  market delivery, daily mystery refresh, account-wide mystery stock, and final pricing.
- `docs/as-built/v66_progress-backlog-hygiene.md` summarizes the shipped cleanup.
- `make ci` remains green.
