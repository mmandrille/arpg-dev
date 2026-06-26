# --- Shared contracts ---------------------------------------------------------
.PHONY: validate-shared validate-assets gen-assets gen-anims
validate-shared: tools ## Validate all shared JSON (protocol, rules, golden) against schemas
	$(PY) tools/validate_shared.py
	$(PY) tools/validate_codemap.py

validate-assets: tools ## Validate the asset manifest, runtime .glb paths, and GLB nodes
	$(PY) tools/assets/validate_assets.py

gen-assets: tools ## Regenerate committed runtime .glb files (deterministic source-of-truth)
	$(PY) tools/assets/gen_glb.py
	$(PY) tools/assets/rig_hero_glbs.py
	$(PY) tools/assets/rig_monster_glbs.py
	$(PY) tools/assets/rig_quadruped_monster_glbs.py

gen-anims: ## Regenerate committed AnimationLibrary .tres clips (requires Godot)
	$(GODOT) --headless --rendering-method gl_compatibility --path client --import >/dev/null 2>&1 || true
	$(GODOT) --headless --rendering-method gl_compatibility --path client --script res://tools/build_animations.gd
