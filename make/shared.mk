# --- Shared contracts ---------------------------------------------------------
.PHONY: validate-shared validate-assets gen-assets
validate-shared: tools ## Validate all shared JSON (protocol, rules, golden) against schemas
	$(PY) tools/validate_shared.py

validate-assets: tools ## Validate the asset manifest, runtime .glb paths, and GLB nodes
	$(PY) tools/assets/validate_assets.py

gen-assets: tools ## Regenerate committed runtime .glb files (deterministic source-of-truth)
	$(PY) tools/assets/gen_glb.py
