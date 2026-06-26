# v343 — Player damage vignette

Screen-edge vignette pulses on local `player_damaged` events, scaled by damage vs max HP. Wired through `CombatEventPresentationScript.bind_camera(..., player_id)` and `PlayerDamageVignette.attach()`.

Verification: `godot --headless --path client --script res://tests/test_look_and_feel_polish.gd`
