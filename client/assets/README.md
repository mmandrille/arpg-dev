# `client/assets/` — runtime `.glb` assets (Godot-imported)

Committed runtime glTF-binary models that Godot imports. The Godot project root **is** `client/`,
so a manifest `runtime_path` of `client/assets/foo/bar.glb` (repo-root-relative) is loaded in-engine
as `res://assets/foo/bar.glb` — strip `client/`, prepend `res://` (ADR-0006 D6).

```
client/assets/
  characters/base_humanoid/base_humanoid.glb
  equipment/weapons/rusty_sword/rusty_sword.glb
```

These files are the source of truth the asset manifest (`assets/manifests/assets.v0.json`) points
at and the `make validate-assets` command checks (existence + `provenance.sha256`).
