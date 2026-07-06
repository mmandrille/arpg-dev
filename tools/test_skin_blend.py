import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from tools.assets.gen_glb import barbarian_glb
from tools.assets.skin_blend import blend_lateral, blend_segment, blend_segment_triple


def test_blend_segment_end_weights():
    fn = blend_segment((0.0, 0.0, 0.0), (0.0, 1.0, 0.0), ja=2, jb=1, split=0.4)
    j0, w0 = fn(0.0, 0.0, 0.0)
    j2, w2 = fn(0.0, 1.0, 0.0)
    assert j0[1] == 1 and abs(w0[1] - 1.0) < 1e-6
    assert j2[0] == 2 and abs(w2[0] - 1.0) < 1e-6


def test_blend_segment_midpoint_is_mixed():
    fn = blend_segment((0.0, 0.0, 0.0), (0.0, 1.0, 0.0), ja=2, jb=1, split=0.5)
    _, w = fn(0.0, 0.2, 0.0)
    assert w[0] > 0.0 and w[1] > 0.0
    assert abs(sum(w) - 1.0) < 1e-6


def test_blend_segment_triple_mid_is_pure_b():
    fn = blend_segment_triple((0.0, 0.0, 0.0), (0.0, 1.0, 0.0), ja=1, jb=2, jc=3, zone_a=0.3, zone_c=0.7)
    _, w = fn(0.0, 0.5, 0.0)
    assert abs(w[1] - 1.0) < 1e-6


def test_blend_segment_triple_endpoints_mix():
    fn = blend_segment_triple((0.0, 0.0, 0.0), (0.0, 1.0, 0.0), ja=1, jb=2, jc=3, zone_a=0.4, zone_c=0.7)
    _, w0 = fn(0.0, 0.0, 0.0)
    _, w1 = fn(0.0, 1.0, 0.0)
    assert abs(w0[1] - 1.0) < 1e-6
    _, wm = fn(0.0, 0.85, 0.0)
    assert wm[1] > 0.0 and wm[2] > 0.0
    _, w1 = fn(0.0, 1.0, 0.0)
    assert abs(w1[2] - 1.0) < 1e-6


def test_barbarian_glb_has_triple_weights():
    data = barbarian_glb()
    import json
    import struct
    jlen = struct.unpack_from("<I", data, 12)[0]
    blen = struct.unpack_from("<I", data, 20 + jlen)[0]
    bin_buf = data[20 + jlen + 8:20 + jlen + 8 + blen]
    gltf = json.loads(data[20:20 + jlen])
    w_view = gltf["bufferViews"][gltf["accessors"][4]["bufferView"]]
    off = w_view["byteOffset"]
    count = gltf["accessors"][4]["count"]
    triple = 0
    for i in range(count):
        w0, w1, w2, w3 = struct.unpack_from("<ffff", bin_buf, off + i * 16)
        if w2 > 0.01:
            triple += 1
    assert triple > 50, f"expected triple-weight verts, got {triple}"


def test_barbarian_glb_has_dual_weights():
    data = barbarian_glb()
    jlen = struct.unpack_from("<I", data, 12)[0]
    blen = struct.unpack_from("<I", data, 20 + jlen)[0]
    bin_start = 20 + jlen + 8
    bin_buf = data[bin_start:bin_start + blen]

    import json
    gltf = json.loads(data[20:20 + jlen])
    w_acc = gltf["accessors"][4]
    w_view = gltf["bufferViews"][w_acc["bufferView"]]
    off = w_view["byteOffset"]
    count = w_acc["count"]
    dual = 0
    for i in range(count):
        w0, w1, w2, w3 = struct.unpack_from("<ffff", bin_buf, off + i * 16)
        if w1 > 0.01:
            dual += 1
    assert dual > 100, f"expected many dual-weight verts, got {dual}"
