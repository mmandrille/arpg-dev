# v8 — Equipped weapon damage

**Proves:** Equipped item rules can change authoritative combat outcomes without protocol, replay, or client UI changes.

- `rusty_sword` declares `damage: {min: 3, max: 5}` in `shared/rules/items.v0.json`.
- Server attack damage resolves the equipped weapon at hit time; missing/no-damage equipment falls back to `combat.player_damage`.
- Go and GDScript golden tests consume `shared/golden/equipped_weapon_damage.json`.
- `tools/validate_shared.py` rejects damage on non-weapon or non-equippable items and checks golden/rules drift.
- `gear_before_combat` now asserts `training_dummy_reward` dies in one acknowledged equipped attack.
- Replay, reconnect resume, `/state`, and Godot smoke stay green through `make ci`.

**Explicit non-goals:** no additive stat system, armor, healing, client damage preview, or inventory
UI/plugin adoption. Attack range was deferred in v8 and closed by v10.
