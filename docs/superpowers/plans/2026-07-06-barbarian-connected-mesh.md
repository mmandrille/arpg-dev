# Barbarian Connected Mesh Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eliminate visible flat ring separations between body segments by adding cap suppression to `_prism_geom`, then rewriting `barbarian_glb()` with flush boundary conditions and a raised shoulder position.

**Architecture:** Two-task sequence: (1) add `cap_bot`/`cap_top` boolean params to `_prism_geom` with updated tests; (2) replace the 16-part barbarian parts list with a 17-part connected stack using the exact boundary Y values and radii from the spec.

**Tech Stack:** Python 3 stdlib only (`math`, `struct`, `json`). No new dependencies.

## Global Constraints

- `_prism_geom` signature change must be **backward-compatible** — all existing callers (sorcerer, paladin, rogue, ranger, monsters) must work unchanged with `cap_bot=True, cap_top=True` defaults.
- 17-bone skeleton joints list in `_full_humanoid_glb` is **frozen** — no changes to `name`, `parent`, or `local` translation of any joint.
- All other characters, monsters, and equipment are untouched.
- `make validate-assets` must pass after regeneration.
- `make test-py` must pass (all 5 existing tests + 2 new cap tests = 7 total).
- `gen_glb.py` current baseline: **642 lines**. If final count exceeds 642+25=667, update `.maintainability/file-size-baseline.tsv`.

---

### Task 1: Add `cap_bot` / `cap_top` params to `_prism_geom`

**Files:**
- Modify: `tools/assets/gen_glb.py:61-115` (`_prism_geom` function)
- Modify: `tools/test_gen_glb.py` (add 2 new tests; existing 5 must still pass)

**Interfaces:**
- Produces: `_prism_geom(n, r_bot, r_top, h, cap_bot=True, cap_top=True)` — same return type `(list, list, list)`
  - Vertex counts: both caps → `6n+2`; one cap → `5n+1`; no caps → `4n`
  - Index counts: both caps → `12n`; one cap → `9n`; no caps → `6n`

- [ ] **Step 1: Write two failing tests**

Add to `tools/test_gen_glb.py` (append after existing tests):

```python
def test_prism_geom_no_caps_vertex_count():
    pos, nrm, idx = _prism_geom(8, 0.1, 0.1, 0.2, cap_bot=False, cap_top=False)
    assert len(pos) == 4 * 8   # 32 — side faces only
    assert len(idx) == 6 * 8   # 48


def test_prism_geom_one_cap_vertex_count():
    pos, nrm, idx = _prism_geom(8, 0.1, 0.1, 0.2, cap_bot=True, cap_top=False)
    assert len(pos) == 5 * 8 + 1   # 41 — sides + bottom cap only
    assert len(idx) == 9 * 8       # 72
```

- [ ] **Step 2: Run tests — confirm 2 new tests fail**

```bash
cd /Users/mmandrille/git/arpg-dev && .venv/bin/pytest tools/test_gen_glb.py -v
```

Expected: 5 old tests PASS, 2 new tests FAIL with `TypeError: _prism_geom() got unexpected keyword argument 'cap_bot'`

- [ ] **Step 3: Modify `_prism_geom` to accept cap params**

In `tools/assets/gen_glb.py`, replace the function signature and the bottom/top cap blocks. The current function at line 61 reads:

```python
def _prism_geom(n: int, r_bot: float, r_top: float, h: float):
    """N-sided frustum centered on Y axis (y: -h/2 to +h/2), flat-shaded faces."""
    pos, nrm, idx = [], [], []
    hh = h / 2.0
    # Side faces: 4n vertices (4 per quad face)
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
    # Bottom cap: 1 center + n edge vertices
    cy = -hh
    cn = (0.0, -1.0, 0.0)
    c_bot = len(pos)
    pos.append((0.0, cy, 0.0))
    nrm.append(cn)
    bot_edges = []
    for i in range(n):
        a = 2 * math.pi * i / n
        bot_edges.append(len(pos))
        pos.append((r_bot * math.cos(a), cy, r_bot * math.sin(a)))
        nrm.append(cn)
    for i in range(n):
        p0 = bot_edges[i]
        p1 = bot_edges[(i + 1) % n]
        idx += [c_bot, p1, p0]  # CW from below = CCW viewed from -y
    # Top cap: 1 center + n edge vertices
    cy = hh
    cn = (0.0, 1.0, 0.0)
    c_top = len(pos)
    pos.append((0.0, cy, 0.0))
    nrm.append(cn)
    top_edges = []
    for i in range(n):
        a = 2 * math.pi * i / n
        top_edges.append(len(pos))
        pos.append((r_top * math.cos(a), cy, r_top * math.sin(a)))
        nrm.append(cn)
    for i in range(n):
        p0 = top_edges[i]
        p1 = top_edges[(i + 1) % n]
        idx += [c_top, p0, p1]  # CCW viewed from +y
    return pos, nrm, idx
```

Replace the entire function with:

```python
def _prism_geom(n: int, r_bot: float, r_top: float, h: float,
                cap_bot: bool = True, cap_top: bool = True):
    """N-sided frustum centered on Y axis (y: -h/2 to +h/2), flat-shaded faces.

    cap_bot/cap_top=False suppresses the disc at that end — use at internal
    joints between adjacent frustums to remove visible ring separations.
    Vertex counts: both caps 6n+2, one cap 5n+1, no caps 4n.
    """
    pos, nrm, idx = [], [], []
    hh = h / 2.0
    # Side faces: 4n vertices (4 per quad face)
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
    if cap_bot:
        cy = -hh
        cn = (0.0, -1.0, 0.0)
        c_bot = len(pos)
        pos.append((0.0, cy, 0.0))
        nrm.append(cn)
        bot_edges = []
        for i in range(n):
            a = 2 * math.pi * i / n
            bot_edges.append(len(pos))
            pos.append((r_bot * math.cos(a), cy, r_bot * math.sin(a)))
            nrm.append(cn)
        for i in range(n):
            idx += [c_bot, bot_edges[(i + 1) % n], bot_edges[i]]
    if cap_top:
        cy = hh
        cn = (0.0, 1.0, 0.0)
        c_top = len(pos)
        pos.append((0.0, cy, 0.0))
        nrm.append(cn)
        top_edges = []
        for i in range(n):
            a = 2 * math.pi * i / n
            top_edges.append(len(pos))
            pos.append((r_top * math.cos(a), cy, r_top * math.sin(a)))
            nrm.append(cn)
        for i in range(n):
            idx += [c_top, top_edges[i], top_edges[(i + 1) % n]]
    return pos, nrm, idx
```

- [ ] **Step 4: Run all 7 tests — all must pass**

```bash
.venv/bin/pytest tools/test_gen_glb.py -v
```

Expected: 7/7 PASS. If any of the original 5 fail, the default `cap_bot=True, cap_top=True` is broken — re-check that the `if cap_bot:` guard does NOT execute when `cap_bot=True`.

- [ ] **Step 5: Commit**

```bash
git add tools/assets/gen_glb.py tools/test_gen_glb.py
git commit -m "feat: add cap_bot/cap_top params to _prism_geom — suppress joint disc faces"
```

---

### Task 2: Rewrite `barbarian_glb()` with connected stack

**Files:**
- Modify: `tools/assets/gen_glb.py:452-481` (`barbarian_glb` function body)
- Regenerate: `client/assets/characters/barbarian/barbarian.glb`
- Maybe update: `.maintainability/file-size-baseline.tsv`

**Interfaces:**
- Consumes: `_prism_geom(..., cap_bot=..., cap_top=...)` from Task 1
- Consumes: `_full_humanoid_glb(skin, parts=parts)` — unchanged signature

- [ ] **Step 1: Run existing tests to confirm green baseline**

```bash
cd /Users/mmandrille/git/arpg-dev && .venv/bin/pytest tools/test_gen_glb.py -v
```

Expected: 7/7 PASS (5 original + 2 cap tests from Task 1).

- [ ] **Step 2: Replace `barbarian_glb()` body**

In `tools/assets/gen_glb.py`, replace the entire `barbarian_glb` function (lines 452–481) with:

```python
def barbarian_glb() -> bytes:
    """17-bone low-poly humanoid — connected frustum stack, no internal cap rings."""
    skin = (0.66, 0.36, 0.25, 1.0)
    parts = [
        # --- Torso stack: shared boundaries, no internal caps ---
        # Hips (bone 1): y [0.889, 1.050], r=0.17 both ends
        ( 1, ( 0.000,  0.970,  0.159), None, skin,
          _prism_geom(8, 0.17, 0.17, 0.161, cap_bot=True,  cap_top=False)),
        # Lower torso (bone 1): y [1.050, 1.340], r 0.17→0.19
        ( 1, ( 0.000,  1.195,  0.159), None, skin,
          _prism_geom(8, 0.17, 0.19, 0.290, cap_bot=False, cap_top=False)),
        # Upper torso (bone 2): y [1.340, 1.550], r 0.19→0.23
        ( 2, ( 0.000,  1.445,  0.159), None, skin,
          _prism_geom(8, 0.19, 0.23, 0.210, cap_bot=False, cap_top=False)),
        # Neck (bone 3): y [1.550, 1.700], r 0.23→0.16
        ( 3, ( 0.000,  1.625,  0.159), None, skin,
          _prism_geom(8, 0.23, 0.16, 0.150, cap_bot=False, cap_top=False)),
        # Head (bone 4): y [1.700, 1.902], r 0.16→0.13
        ( 4, ( 0.000,  1.801,  0.159), None, skin,
          _prism_geom(8, 0.16, 0.13, 0.202, cap_bot=False, cap_top=True)),
        # --- Right arm: elbow joint connected, shoulder + wrist capped ---
        # Upper arm (bone 8): y [1.131, 1.550], elbow at bottom (no cap)
        ( 8, ( 0.316,  1.341,  0.035), None, skin,
          _prism_geom(6, 0.042, 0.060, 0.419, cap_bot=False, cap_top=True)),
        # Forearm (bone 9): y [0.877, 1.131], elbow at top (no cap)
        ( 9, ( 0.252,  1.004,  0.044), None, skin,
          _prism_geom(6, 0.036, 0.042, 0.254, cap_bot=True,  cap_top=False)),
        (10, ( 0.223,  0.877,  0.048), (0.09, 0.10, 0.07)),  # box — too small for prism
        # --- Left arm (mirrored X) ---
        ( 5, (-0.316,  1.341,  0.035), None, skin,
          _prism_geom(6, 0.042, 0.060, 0.419, cap_bot=False, cap_top=True)),
        ( 6, (-0.252,  1.004,  0.044), None, skin,
          _prism_geom(6, 0.036, 0.042, 0.254, cap_bot=True,  cap_top=False)),
        ( 7, (-0.223,  0.877,  0.048), (0.09, 0.10, 0.07)),  # box — too small for prism
        # --- Right leg: knee joint connected, hip + ankle capped ---
        # Thigh (bone 14): y [0.493, 0.889], knee at bottom (no cap)
        (14, ( 0.155,  0.691,  0.159), None, skin,
          _prism_geom(6, 0.062, 0.075, 0.396, cap_bot=False, cap_top=True)),
        # Shin (bone 15): y [0.039, 0.493], knee at top (no cap)
        (15, ( 0.194,  0.266,  0.132), None, skin,
          _prism_geom(6, 0.048, 0.062, 0.454, cap_bot=True,  cap_top=False)),
        (16, ( 0.232,  0.039,  0.105), (0.13, 0.07, 0.20)),  # box — too small for prism
        # --- Left leg (mirrored X) ---
        (11, (-0.155,  0.691,  0.159), None, skin,
          _prism_geom(6, 0.062, 0.075, 0.396, cap_bot=False, cap_top=True)),
        (12, (-0.194,  0.266,  0.132), None, skin,
          _prism_geom(6, 0.048, 0.062, 0.454, cap_bot=True,  cap_top=False)),
        (13, (-0.232,  0.039,  0.105), (0.13, 0.07, 0.20)),  # box — too small for prism
    ]
    return _full_humanoid_glb(skin, parts=parts)
```

- [ ] **Step 3: Run tests — all 7 must still pass**

```bash
.venv/bin/pytest tools/test_gen_glb.py -v
```

Expected: 7/7 PASS. `test_barbarian_glb_is_valid_gltf` exercises the new parts list and must return valid glTF bytes.

- [ ] **Step 4: Regenerate assets**

```bash
make gen-assets
```

Expected: prints `wrote client/assets/characters/barbarian/barbarian.glb (XXXXX bytes)`. New file will be larger than 44 KB (more geometry due to neck part + connected mesh).

- [ ] **Step 5: Validate assets**

```bash
make validate-assets
```

Expected: no errors. If SHA256 mismatch error appears, the manifest was not auto-updated — check that `make gen-assets` ran successfully and re-run.

- [ ] **Step 6: Run CI**

```bash
make ci
```

Expected: all steps pass. If maintainability fails: check `wc -l tools/assets/gen_glb.py`. If the count exceeds 667 (current baseline 642 + 25 allowance), update the entry in `.maintainability/file-size-baseline.tsv`:

```
tools/assets/gen_glb.py	<actual_count>	# connected barbarian mesh: cap_bot/cap_top in _prism_geom; extraction deferred — binary packing pipeline is tightly coupled
```

- [ ] **Step 7: Visual check**

```bash
make model model=character_barbarian_v0
```

In the Godot preview confirm:
- No horizontal ring lines between hips/spine/chest/neck/head — the torso reads as one connected shape
- No ring line between upper arm and forearm at the elbow
- No ring line between thigh and shin at the knee
- Arm tops are higher than before (shoulder at chest level y≈1.55)
- Bones still sit inside the mesh segments

- [ ] **Step 8: Commit**

```bash
git add tools/assets/gen_glb.py client/assets/characters/barbarian/barbarian.glb
# add .maintainability/file-size-baseline.tsv only if it changed
git commit -m "feat: barbarian connected mesh — cap suppression at torso/arm/leg joints, raised shoulders"
```
