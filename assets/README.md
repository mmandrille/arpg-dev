# `assets/` — asset sources, exports, and manifests

This tree is the **authoring/provenance** side of the pipeline (ADR-0006).

```
assets/
  characters/          base character source/export notes (e.g. base_humanoid)
  equipment/weapons/   weapon source/export notes (e.g. rusty_sword)
  manifests/           asset manifest JSON: asset_id -> runtime .glb path (+ provenance)
```

## v2 policy (ADR-0006 D2)

- **Runtime `.glb` files live under `client/assets/...`**, not here. They are the committed,
  imported-by-Godot artifacts the manifest points at via `runtime_path`.
- This tree commits **no heavyweight binary sources** (`.blend`, etc.) for v2. The manifest's
  `source_path` documents the *intended* source layout; the actual runtime asset's origin is
  recorded as `provenance` (source/origin URL, `license`, `sha256`) in the manifest entry and the
  spec's as-built notes.
- The per-asset `README.md` under each source dir records where that asset came from and how to
  regenerate/re-source it.

The proof of this slice is the **manifest → import → mount** contract, not the authorship tool.
