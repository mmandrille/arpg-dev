# v2 — Equip and see it

**Proves:** ADR-0001 D7 Tier A + ADR-0006 asset pipeline contract.

- Shared `item_visuals.v0.json` + `assets.v0.json` → Godot mount on `right_hand_socket`
- Deterministic `gen_glb.py` runtime assets; `make validate-assets`
- Equipped `rusty_sword` visible on character; server authority unchanged
- Resume restores equipped weapon from persisted inventory

**Scope limit:** only `rusty_sword` has a visual mapping; other items deferred.
