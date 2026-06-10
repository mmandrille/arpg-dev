# v36 — Inventory paper-doll capacity

**Proves:** Inventory capacity can be server-owned, item-derived, and rendered as a fixed
paper-doll bag grid without making the Godot UI authoritative.

- Shared item template rules now allow `inventory_rows` and include deterministic
  `cave_pack_belt`, with `shared/golden/inventory_capacity.json` pinning base rows `3`,
  5-column base capacity `15`, +1 row capacity `20`, hotbar/equipped exemptions, and rejection
  reasons.
- Session snapshots expose `inventory_rows` and `inventory_capacity`; relevant equipped/hotbar
  deltas publish the same fields so client, bot, reconnect, and replay observe identical capacity.
- Go sim derives capacity from equipped items, counts only bag entries that are not equipped and not
  assigned to hotbar, rejects full pickups with `inventory_full`, and rejects capacity-shrinking
  unequip/unassign paths before mutation with `capacity_would_overflow`.
- Godot replaces the two-column equipment list with a named paper-doll layout around a
  `character_paper_doll` placeholder, renders a 5-column bag with exactly `inventory_capacity`
  visible cells, and keeps the bag drop target outside grid math.
- Inventory debug state now reports paper-doll slot ids/positions, preview status, capacity rows,
  bag columns, available slot count, and empty-slot style markers for headless assertions.
- Protocol bot scenario `25_inventory_capacity_and_paper_doll.json` proves base capacity, full-bag
  rejection, +1 row equip to capacity 20, five more bag entries, reconnect, and replay.
- Client bot scenario `13_inventory_paper_doll.json` proves base 15-cell grid, all paper-doll slot
  ids, the preview node, belt equip, and expanded 20-cell grid.

**Explicit non-goals:** no stash, vendors, crafting, item sorting/filtering, comparison UI,
multi-cell item footprints, passive skill sources for inventory rows, production paper-doll art,
or full model-backed SubViewport preview.
