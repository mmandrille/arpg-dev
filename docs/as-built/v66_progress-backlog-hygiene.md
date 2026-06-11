# v66 As-built: Progress Backlog Hygiene

## What shipped

- `PROGRESS.md` now treats v64 `mystery-seller-paid-reroll` and v65
  `stash-search-and-sorting` consistently as completed slices across discovery sections.
- The scenario catalog includes the protocol/client bot proofs that shipped with those slices.
- Deferred backlog text now preserves still-open adjacent work without re-listing paid mystery
  rerolls or stash search/sort as unbuilt.

## Proof

- Docs-focused `rg` checks confirm stale open/deferred references were removed or narrowed.
- `git diff --check` passes.
- `make ci` passes.
