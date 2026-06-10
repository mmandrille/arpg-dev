# v54 — Character select summaries

**Proves:** The player-facing character picker can show durable server-authored progression before
starting a session.

- `GET /v0/characters` now returns account-scoped summary fields for each character: `level`,
  `gold`, and `deepest_dungeon_depth`, alongside the existing name, id, dead state, and created date.
- Character listing remains read-only: missing durable progression rows are left-joined and
  coalesced to display defaults (`level: 1`, `gold: 0`, `deepest_dungeon_depth: 0`) instead of
  creating persistence as a side effect.
- Store and HTTP tests prove exact progression summaries, default rows, and account scoping.
- `CharacterSelectPanel` renders compact row summaries and exposes structured `character_rows`
  debug state without changing create, select, rename, delete, or dead-row behavior.
- Client bot scenario `27_character_select_summaries.json` starts from the main menu and proves
  the Choose Character panel exposes level/gold/depth/status summary data.
- The full autoloop idea menu from 2026-06-10 is recorded in curated candidates for future reuse.

---
