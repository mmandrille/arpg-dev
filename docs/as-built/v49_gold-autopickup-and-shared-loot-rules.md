# v49 — Gold auto-pickup and shared loot rules

**Proves:** Floor loot remains shared for everyone while gold gains a server-owned quality-of-life
auto-pickup path.

- Gold floor entities are still shared world entities with no owner/reservation/personal-loot field.
- After connected player movement resolves, the sim scans levels, gold entity ids, and eligible
  player ids in stable order; the lowest eligible same-level alive connected player in loot range
  wins the gold.
- Gold pickup emits public `entity_remove` plus winner-private `gold_update`,
  `character_progression_update`, and `gold_picked_up`; realtime filtering and persistence route by
  explicit owner even when passive pickup has no actor input.
- Manual in-range gold pickup and out-of-range auto-navigation compatibility are preserved; passive
  pickup cannot duplicate a pending clicked gold entity.
- Non-gold item loot does not auto-pick up and still requires explicit `action_intent`.
- Protocol v6 needed no schema bump: existing deltas/events cover passive gold pickup, and
  `gold_picked_up.correlation_id` remains optional.
- Protocol bot scenario `35_gold_autopickup_shared_loot.json` proves shared co-op floor gold,
  auto-pickup without click, winner-private wallet updates, replay/fresh-session persistence, and a
  deterministic click-required non-gold item proof.

**Explicit non-goals:** no personal loot, hidden loot, duplicated per-player drops, loot
reservations, shared/split gold, item auto-pickup, loot allocation UI, drop-rate rebalance, or client
UI/art/audio changes.
