# v109 As-Built: Permanent Death Corpse Recovery

**Status:** Complete on `main`

## What shipped

- Permanent character death now leaves a server-owned hero corpse with persisted item recovery data.
- Corpse recovery is exposed through protocol v8 snapshots/deltas/events and a `corpse_withdraw_item_intent`.
- Realtime session attach and resume include active session corpses so surviving players can recover items through the normal authoritative session loop.
- The Godot client renders hero corpses as interactables, supports corpse item withdrawal through the stash-style panel, and gives corpse targets loot-label style hover/reveal affordances.

## Verification

- The slice landed with focused Go, replay, protocol, and Godot client test coverage in the v109 commit series.
- The latest repository baseline after the v109 follow-up commits is synced with `origin/main`.

## Scope limits

- No dedicated v109 spec or plan file was committed with the external-agent implementation.
- Respawn/checkpoint design, corpse expiration rules, and broader death-economy tuning remain future work.
