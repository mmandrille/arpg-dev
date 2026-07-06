# Barbarian Low-Poly Mesh Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the barbarian's block-cube body segments with low-poly frustum/prism shapes that match the reference humanoid silhouette while keeping the 17-bone skeleton frozen.

**Architecture:** Add `_prism_geom(n, r_bot, r_top, h)` to `gen_glb.py`, extend `_build_skinned_glb` to accept pre-baked custom geometry as a 5th part element, then rewrite `barbarian_glb()` to use frustum parts at the existing mesh-space centers.

**Tech Stack:** Python 3 stdlib only (`math`, `struct`, `json`). No new dependencies.

## Global Constraints

- 17-bone skeleton joints list in `_full_humanoid_glb` is frozen — no changes to `name`, `parent`, or `local` translation of any joint.
- All mesh-part centers match existing box centers from the spec exactly (tolerance ±0.001).
- All other characters (sorcerer, paladin, rogue, ranger), monsters, and equipment are untouched.
- `make validate-assets` must pass after regeneration.
- `make test-py` must pass.
- `gen_glb.py` will exceed 600 lines after this work (~621). Add it to `.maintainability/file-size-baseline.tsv` with the actual post-implementation line count.

---

### Task 1: Add `_prism_geom` + extend `_build_skinned_glb`

**Files:**
- Modify: `tools/assets/gen_glb.py` (add `import math`, add `_prism_geom`, refactor part loop in `_build_skinned_glb`)
- Create: `tools/test_gen_glb.py`

**Interfaces:**
- Produces: `_prism_geom(n: int, r_bot: float, r_top: float, h: float) -> tuple[list, list, list]`
  - Returns `(positions, normals, indices)` — same contract as `_cube_geometry()`
  - Frustum centered on Y axis: bottom cap at `y = -h/2`, top cap at `y = +h/2`
  - Side faces have flat outward normals (low-poly shading)
  - Vertex count: `6*n + 2`; index count: `12*n`
- Produces: `_build_skinned_glb` now also accepts 5-element parts:
  `(joint_idx, (tx,ty,tz), None, color_or_None, (positions, normals, indices))`

- [ ] **Step 1: Write failing tests**

Create `tools/test_gen_glb.py`:

```python
import math
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from tools.assets.gen_glb import _prism_geom, _build_skinned_glb, barbarian_glb


def test_prism_geom_vertex_count():
    pos, nrm, idx = _prism_geom(8, 0.1, 0.1, 0.2)
    assert len(pos) == 6 * 8 + 2  # 50
    assert len(nrm) == len(pos)
    assert len(idx) == 12 * 8     # 96


def test_prism_geom_indices_in_range():
    pos, nrm, idx = _prism_geom(6, 0.05, 0.08, 0.4)
    assert all(0 <= i < len(pos) for i in idx)


def test_prism_geom_normals_unit_length():
    pos, nrm, idx = _prism_geom(8, 0.13, 0.13, 0.28)
    for n in nrm:
        length = math.sqrt(sum(v * v for v in n))
        assert abs(length - 1.0) < 1e-5, f"normal not unit: {n} length={length}"


def test_prism_geom_side_normals_outward():
    # Side face normals for a regular prism must point away from Y axis
    pos, nrm, idx = _prism_geom(8, 0.1, 0.1, 0.2)
    # First 8*4=32 vertices are side face verts; their normals have y near 0
    # and xz component pointing outward (dot with position xz > 0)
    for i in range(8 * 4):
        px, py, pz = pos[i]
        nx, ny, nz = nrm[i]
        dot_xz = px * nx + pz * nz
        assert dot_xz > 0.0, f"side normal points inward at vert {i}: pos={pos[i]}, nrm={nrm[i]}"


def test_barbarian_glb_is_valid_gltf():
    data = barbarian_glb()
    assert data[:4] == b"glTF"
    assert len(data) > 10_000
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /path/to/arpg-dev && .venv/bin/pytest tools/test_gen_glb.py -v
```

Expected: `ImportError` or `AttributeError: module has no attribute '_prism_geom'`

- [ ] **Step 3: Add `import math` at the top of `gen_glb.py`**

In `tools/assets/gen_glb.py`, add after the stdlib imports block (after `from pathlib import Path`):

```python
import math
```

- [ ] **Step 4: Add `_prism_geom` after `_cube_geometry`**

Insert after the `_pad` function (after line ~58), before `_build_glb`:

```python
def _prism_geom(n: int, r_bot: float, r_top: float, h: float):
    """N-sided frustum centered on Y axis (y: -h/2 to +h/2), flat-shaded faces."""
    pos, nrm, idx = [], [], []
    hh = h / 2.0
    for i in range(n):
        a0 = 2 * math.pi * i / n
        a1 = 2 * math.pi * (i + 1) / n
        am = (a0 + a1) / 2.0
        bl = (r_bot * math.cos(a0), -hh, r_bot * math.sin(a0))
        br = (r_bot * math.cos(a1), -hh, r_bot * math.sin(a1))
        tr = (r_top * math.cos(a1),  hh, r_top * math.sin(a1))
        tl = (r_top * math.cos(a0),  hh, r_top * math.sin(a0))
        dr = (r_top - r_bot) / h if h else 0.0
        raw = (math.cos(am), -dr, math.sin(am))
        nl = math.sqrt(sum(v * v for v in raw))
        fn = tuple(v / nl for v in raw)
        b = len(pos)
        for v in (bl, tl, tr, br):  # CCW from outside → outward normal
            pos.append(v)
            nrm.append(fn)
        idx += [b, b + 1, b + 2, b, b + 2, b + 3]
    for cy, r, yn in ((-hh, r_bot, -1.0), (hh, r_top, 1.0)):
        cn = (0.0, yn, 0.0)
        for i in range(n):
            a0 = 2 * math.pi * i / n
            a1 = 2 * math.pi * (i + 1) / n
            ci = len(pos)
            pos.append((0.0, cy, 0.0))
            nrm.append(cn)
            p0 = len(pos)
            pos.append((r * math.cos(a0), cy, r * math.sin(a0)))
            nrm.append(cn)
            p1 = len(pos)
            pos.append((r * math.cos(a1), cy, r * math.sin(a1)))
            nrm.append(cn)
            if yn < 0:
                idx += [ci, p1, p0]  # CW from below = CCW viewed from -y
            else:
                idx += [ci, p0, p1]  # CCW viewed from +y
    return pos, nrm, idx
```

- [ ] **Step 5: Extend the part loop in `_build_skinned_glb`**

Find this block in `_build_skinned_glb` (around line 183):

```python
    for part in parts:
        if len(part) == 4:
            joint_idx, (tx, ty, tz), (sx, sy, sz), part_color = part
        else:
            joint_idx, (tx, ty, tz), (sx, sy, sz) = part
            part_color = color
        base = len(positions)
        for (px, py, pz), n in zip(cube_pos, cube_nrm):
            positions.append((px * sx + tx, py * sy + ty, pz * sz + tz))
            normals.append(n)
            colors0.append(part_color)
            joints0.append((joint_idx, 0, 0, 0))
            weights0.append((1.0, 0.0, 0.0, 0.0))
        for i in cube_idx:
            indices.append(base + i)
```

Replace with:

```python
    for part in parts:
        if len(part) == 5:
            joint_idx, (tx, ty, tz), _, part_color, (geom_pos, geom_nrm, geom_idx) = part
            if part_color is None:
                part_color = color
        elif len(part) == 4:
            joint_idx, (tx, ty, tz), (sx, sy, sz), part_color = part
            geom_pos = [(px * sx, py * sy, pz * sz) for px, py, pz in cube_pos]
            geom_nrm, geom_idx = cube_nrm, cube_idx
        else:
            joint_idx, (tx, ty, tz), (sx, sy, sz) = part
            part_color = color
            geom_pos = [(px * sx, py * sy, pz * sz) for px, py, pz in cube_pos]
            geom_nrm, geom_idx = cube_nrm, cube_idx
        base = len(positions)
        for (px, py, pz), n in zip(geom_pos, geom_nrm):
            positions.append((px + tx, py + ty, pz + tz))
            normals.append(n)
            colors0.append(part_color)
            joints0.append((joint_idx, 0, 0, 0))
            weights0.append((1.0, 0.0, 0.0, 0.0))
        for i in geom_idx:
            indices.append(base + i)
```

- [ ] **Step 6: Run tests — expect `test_barbarian_glb_is_valid_gltf` to fail (function unchanged), others pass**

```bash
.venv/bin/pytest tools/test_gen_glb.py -v
```

Expected: `test_prism_geom_*` all PASS, `test_barbarian_glb_is_valid_gltf` PASS (barbarian_glb still works, just uses boxes). All 5 tests green.

If `test_barbarian_glb_is_valid_gltf` somehow fails, debug before continuing.

- [ ] **Step 7: Commit**

```bash
git add tools/assets/gen_glb.py tools/test_gen_glb.py
git commit -m "feat: add _prism_geom frustum primitive + 5-element part support in _build_skinned_glb"
```

---

### Task 2: Rewrite `barbarian_glb` with frustum parts + ratchet update

**Files:**
- Modify: `tools/assets/gen_glb.py` (`_full_humanoid_glb` signature + `barbarian_glb` body)
- Modify: `.maintainability/file-size-baseline.tsv`

**Interfaces:**
- Consumes: `_prism_geom` from Task 1
- Consumes: `_full_humanoid_glb(color, parts=None, extra_parts=None)` — new `parts` param (Task 2 adds it)

- [ ] **Step 1: Add `parts` parameter to `_full_humanoid_glb`**

Find the function signature and the final `return` line:

```python
def _full_humanoid_glb(color, extra_parts=None) -> bytes:
```

Change to:

```python
def _full_humanoid_glb(color, parts=None, extra_parts=None) -> bytes:
```

Find the block at the end of `_full_humanoid_glb` that looks like:

```python
    if extra_parts:
        parts = list(parts) + list(extra_parts)
    return _build_skinned_glb(color, joints, parts)
```

Replace with (the local variable was already named `parts` inside the function; rename it to `_default_parts` to avoid shadowing the parameter):

```python
    if parts is None:
        parts = _default_parts
    if extra_parts:
        parts = list(parts) + list(extra_parts)
    return _build_skinned_glb(color, joints, parts)
```

And rename the local `parts = [...]` list inside `_full_humanoid_glb` to `_default_parts = [...]`.

- [ ] **Step 2: Rewrite `barbarian_glb`**

Replace the current 3-line function:

```python
def barbarian_glb() -> bytes:
    """17-bone humanoid base — each mesh segment weighted to its bone."""
    return _full_humanoid_glb((0.66, 0.36, 0.25, 1.0))
```

With:

```python
def barbarian_glb() -> bytes:
    """17-bone low-poly humanoid — frustum segments weighted to each bone."""
    skin = (0.66, 0.36, 0.25, 1.0)
    parts = [
        # head (bone 4)
        ( 4, ( 0.000,  1.762,  0.159), None, skin, _prism_geom(8, 0.13, 0.13, 0.28)),
        # upper torso (bone 2): wide at shoulders, narrow at waist
        ( 2, ( 0.000,  1.507,  0.159), None, skin, _prism_geom(8, 0.19, 0.23, 0.33)),
        # lower torso (bone 1)
        ( 1, ( 0.000,  1.236,  0.159), None, skin, _prism_geom(8, 0.17, 0.19, 0.32)),
        # hips (bone 1)
        ( 1, ( 0.000,  0.980,  0.159), None, skin, _prism_geom(8, 0.15, 0.17, 0.18)),
        # right arm: upper (bone 8), forearm (bone 9), hand box (bone 10)
        ( 8, ( 0.316,  1.285,  0.035), None, skin, _prism_geom(6, 0.045, 0.055, 0.32)),
        ( 9, ( 0.252,  1.004,  0.044), None, skin, _prism_geom(6, 0.036, 0.045, 0.26)),
        (10, ( 0.223,  0.877,  0.048), (0.09, 0.10, 0.07)),
        # left arm: upper (bone 5), forearm (bone 6), hand box (bone 7)
        ( 5, (-0.316,  1.285,  0.035), None, skin, _prism_geom(6, 0.045, 0.055, 0.32)),
        ( 6, (-0.252,  1.004,  0.044), None, skin, _prism_geom(6, 0.036, 0.045, 0.26)),
        ( 7, (-0.223,  0.877,  0.048), (0.09, 0.10, 0.07)),
        # right leg: thigh (bone 14), shin (bone 15), foot box (bone 16)
        (14, ( 0.155,  0.690,  0.159), None, skin, _prism_geom(6, 0.062, 0.075, 0.40)),
        (15, ( 0.194,  0.266,  0.132), None, skin, _prism_geom(6, 0.048, 0.062, 0.46)),
        (16, ( 0.232,  0.039,  0.105), (0.13, 0.07, 0.20)),
        # left leg: thigh (bone 11), shin (bone 12), foot box (bone 13)
        (11, (-0.155,  0.690,  0.159), None, skin, _prism_geom(6, 0.062, 0.075, 0.40)),
        (12, (-0.194,  0.266,  0.132), None, skin, _prism_geom(6, 0.048, 0.062, 0.46)),
        (13, (-0.232,  0.039,  0.105), (0.13, 0.07, 0.20)),
    ]
    return _full_humanoid_glb(skin, parts=parts)
```

- [ ] **Step 3: Run tests**

```bash
.venv/bin/pytest tools/test_gen_glb.py -v
```

Expected: all 5 tests PASS.

- [ ] **Step 4: Regenerate assets**

```bash
make gen-assets
```

Expected output includes `wrote client/assets/characters/barbarian/barbarian.glb (XXXXX bytes)`. The new file will be larger than before (frustum geometry has more vertices than boxes).

- [ ] **Step 5: Validate assets**

```bash
make validate-assets
```

Expected: no errors. If it fails, read the error — it's likely a node-name mismatch in the manifest, which would mean the GLB structure changed unexpectedly.

- [ ] **Step 6: Check line count and update maintainability baseline**

```bash
wc -l tools/assets/gen_glb.py
```

If the line count exceeds 600, add an entry to `.maintainability/file-size-baseline.tsv`. The entry format is `path<TAB>linecount<TAB># reason`:

```
tools/assets/gen_glb.py	<actual_count>	# frustum geometry, skinned GLB builder, and character part definitions are tightly coupled to the binary packing pipeline; extraction would require threading buffer state across module boundaries
```

Replace `<actual_count>` with the real number from `wc -l`.

- [ ] **Step 7: Run fast CI**

```bash
make ci
```

Expected: all steps pass including `make maintainability`. If maintainability fails because gen_glb.py is over 600 and not in the baseline, add it per Step 6.

- [ ] **Step 8: Visual check — open barbarian model in Godot**

```bash
make model model=character_barbarian_v0
```

Verify in the Godot preview:
- Body parts have faceted low-poly look (not flat boxes)
- Bones sit inside the mesh segments
- Arms hang at the sides (A-pose / T-pose rest position looks correct)
- No part is obviously misaligned or floating away from its bone

- [ ] **Step 9: Commit**

```bash
git add tools/assets/gen_glb.py client/assets/characters/barbarian/barbarian.glb .maintainability/file-size-baseline.tsv
git commit -m "feat: barbarian low-poly frustum mesh — replace block cubes with n-sided prisms"
```
