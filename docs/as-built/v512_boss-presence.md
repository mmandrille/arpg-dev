# v512 As-built — Boss Presence

- **Status:** Complete (focused verification; integrated and combined `make ci` green, 12m00s)

- **Spec:** [`v512_spec-boss-presence.md`](../specs/v512_spec-boss-presence.md)
- **Plan:** [`v512_2026-10-01-boss-presence.md`](../plans/v512_2026-10-01-boss-presence.md)
- **Base:** `425b9ae4`. Client-only; no server, protocol, golden, or `boss_templates.v0.json` change.

## What shipped

- `shared/assets/boss_presentation.v0.json` (+ schema): per `boss_template_id` display name, epithet,
  kit `visual_key`, scale, tint, attachments, headgear, ground aura, banner timing. Validated by
  `tools/validate_boss_presentation.py` (wired into `validate_shared.py`).
- Cave Warden: Skeleton Warrior rig, axe + large shield, gold crown, violet aura, scale 2.0.
  Crypt Matron: Skeleton Mage rig (baked-in hat), staff, green aura, scale 1.55 (ratio 1.29).
  The server's random `model_pool` (wolf/bat/dummy) is now ignored by the client for catalogued bosses.
- `BossPresentationLoader` (override seam for model, scale, tint), `BossPresentationMounter`
  (bone attachments + data-built horns/crown), `BossArenaPresence` (themed ring, inner disc, pulse,
  soft light; Performance tier keeps the ring only; meshes now rebuild only on change),
  `BossIntroBanner` (once per boss entity, full health, no active phase; re-armed on level reset).
- Bot: `wait_boss_intro_banner`, `has_boss_arena_presence` / `has_boss_headgear` entity flags,
  scenarios 111/112.
- Assets (CC0, staged KayKit Skeletons 1.0): `skeleton_mage.glb`, `skeleton_staff.glb`,
  `skeleton_shield_large_a.glb` + manifest entries with sha256 provenance.

## Proof (all in `/Users/mmandrille/git/arpg-dev-batch/v512-boss-presence`)

- pytest: `tools/test_validate_boss_presentation.py` (8), whole `tools` suite 299 passed;
  `validate_shared.py` 2241 ok; `validate_assets.py` 463 ok; `make maintainability` passed.
- Godot headless: `test_boss_presentation_loader`, `test_boss_intro_banner`, `test_boss_arena_presence`,
  `test_kit_monsters` (auto-covers the two new scenes), `test_boss_health_bar`, `test_boss_lane_marker`,
  `test_factories`, `test_client_bot`, `test_combat_vfx`, `test_aura_soft_lights`.
- Bot: client `boss_intro_banner_cave_warden` / `_crypt_matron` (headless ~5.4 s, non-headless ~7.5 s,
  default 10 s budget, no navigation), regression `boss_portrait_panel`, `boss_lane_telegraphs_visual`,
  protocol `second_boss_template`, `boss_floor_gate`.
- Captures (ignored, `.artifacts/bot-captures/v512_{cave_warden,crypt_matron}_{banner,iso,performance}.png`,
  1920x1080 Metal Forward+): banner mid-fade, isometric with aura, Performance tier.

## Limits

- Draw cost, one boss, Balanced, measured in an isolated scene (ten settled frames): legacy skeleton
  boss 2 draw calls / 708 primitives; Warden 20 / 8686; Matron 12 / 7067 (includes aura ring + disc).
  No matched in-dungeon frame-time comparison was taken; weaker-GPU cost is unverified.
- Boss entity may arrive only inside the server interest radius, so the banner fires on first sight.
- Real-camera chest-view/first-person boss framing is not captured (player spawns overlapping the boss
  in `boss_floor_gate_lab`); close-ups were verified in a standalone model render only.
- The stale `model_pool` and `visual` block in `boss_templates.v0.json` are untouched (follow-up).
