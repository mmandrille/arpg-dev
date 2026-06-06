# Agent entrypoint

Read these **before** specs, plans, or code:

1. [`CLAUDE.md`](CLAUDE.md) — commands, architecture, invariants, SDD process.
2. [`docs/PROGRESS.md`](docs/PROGRESS.md) — current slice baseline and open gaps.
3. [`docs/researchs/godot-plugins-and-shortcuts.md`](docs/researchs/godot-plugins-and-shortcuts.md) — **check for existing Godot plugins, demos, and asset packs** before building new client UI, inventory presentation, isometric/camera tooling, or placeholder art from scratch.

When starting client-side work, run the adoption checklist in the plugins doc and record *adopt / borrow / reject* in the slice plan.

## Development priority

While the game is still in active development, do **not** preserve backward compatibility just for its own sake. Prefer the cleanest, healthiest implementation and update contracts, fixtures, tests, tools, and docs together.
