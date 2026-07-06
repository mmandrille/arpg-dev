import math
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from tools.assets.geom_primitives import _prism_between, _prism_ellipse_geom, _prism_geom
from tools.assets.gen_glb import _build_skinned_glb, barbarian_glb


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


def test_prism_geom_no_caps_vertex_count():
    pos, nrm, idx = _prism_geom(8, 0.1, 0.1, 0.2, cap_bot=False, cap_top=False)
    assert len(pos) == 4 * 8   # 32 — side faces only
    assert len(idx) == 6 * 8   # 48


def test_prism_geom_one_cap_vertex_count():
    pos, nrm, idx = _prism_geom(8, 0.1, 0.1, 0.2, cap_bot=True, cap_top=False)
    assert len(pos) == 5 * 8 + 1   # 41 — sides + bottom cap only
    assert len(idx) == 9 * 8       # 72


def test_prism_ellipse_matches_circle():
    pos_c, nrm_c, idx_c = _prism_geom(8, 0.1, 0.1, 0.2)
    pos_e, nrm_e, idx_e = _prism_ellipse_geom(8, 0.1, 0.1, 0.1, 0.1, 0.2)
    assert len(pos_c) == len(pos_e)
    assert len(idx_c) == len(idx_e)


def test_prism_between_endpoints():
    p0 = (0.0, 0.0, 0.0)
    p1 = (0.0, 1.0, 0.0)
    pos, nrm, idx = _prism_between(p0, p1, 6, 0.1, 0.08)
    ys = [p[1] for p in pos]
    assert min(ys) >= -0.001
    assert max(ys) <= 1.001
    assert all(0 <= i < len(pos) for i in idx)
