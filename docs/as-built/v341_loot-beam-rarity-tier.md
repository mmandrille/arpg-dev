# v341 — Loot beam rarity tier

Tiered rarity glow intensity, rare+ vertical pickup beams, and white highlight for currency labels when hovered/revealed (avoids unique-gold confusion).

Verification: `godot --headless --path client --script res://tests/test_loot_node_factory.gd`, `godot --headless --path client --script res://tests/test_loot_label_filter.gd`, `make client-unit`

Visual: `make bot-visual scenario=01_click_to_kill`
