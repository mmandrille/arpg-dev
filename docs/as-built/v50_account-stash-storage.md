# v50 — Account stash storage

**Proves:** Characters on the same account now have a durable town stash for item and gold storage,
with replay-safe session-start snapshots and actor-private realtime payloads.

- Shared rules add a `town_stash` interactable on `dungeon_levels` town level `0`; protocol v7 adds
  stash item/gold transfer intents, stash snapshot fields, stash change ops, and stash events.
- Account-owned stash tables store items and gold separately from character-scoped inventory rows,
  while session-start stash snapshot tables freeze the replay source of truth.
- The Go sim opens stash through existing `action_intent`, validates item/gold deposit and withdraw,
  rejects equipped/hotbar-assigned/full-capacity/insufficient-funds cases, and emits paired
  inventory/gold plus stash changes without optimistic client ownership.
- Realtime persistence groups each stash transfer through atomic store methods and filters stash
  snapshots, changes, and events to the owning account/player.
- Godot adds a server-driven `StashPanel` beside the existing inventory flow, including stash/bag
  item grids, stash and character gold display, fixed one-gold transfer controls, and client-bot
  hooks.
- Protocol bot scenario `36_account_stash_storage.json` proves dungeon loot/gold acquisition,
  town stash open, item and gold deposit/withdraw, replay, reconnect, `/state`, and fresh-session
  persistence; client bot scenario `23_account_stash_panel.json` proves the live Godot panel flow.

**Explicit non-goals:** no remote stash access outside town, stash sorting/filtering/search/tabs,
capacity upgrades, item stacks, direct equip/use/sell/upgrade from stash, player-market delivery,
crafting/material tabs, arbitrary numeric gold entry, production stash art/audio, or real-time
cross-session push for another already-open session on the same account.
