"""Re-skin static hero GLBs onto the frozen 17-bone canonical humanoid skeleton."""
from __future__ import annotations

import struct

from tools.assets.canonical_skeleton import (
    joint_globals_from_mesh,
    vertex_weights,
)
from tools.assets.rig_hero_glbs import (
    HERO_TARGET_HEIGHTS,
    REQUIRED_BONES,
    _append_accessor,
    _append_buffer_view,
    _bounds,
    _inverse_translation_matrix,
    _joint_nodes,
    _normalize_mesh_height,
    parse_glb,
    read_position_accessor,
    write_glb,
)


def rig_glb_canonical_bytes(data: bytes, *, hero_id: str = "", target_height: float | None = None) -> bytes:
    """Skin a static mesh with canonical joint globals and segment distance weights."""
    parsed = parse_glb(data)
    gltf = parsed.gltf
    if gltf.get("skins"):
        raise ValueError("source GLB is already skinned")

    gltf.pop("animations", None)
    bin_buf = bytearray(parsed.bin_blob)

    positions_by_accessor: dict[int, list[tuple[float, float, float]]] = {}
    primitives: list[dict] = []
    for mesh in gltf.get("meshes", []):
        for primitive in mesh.get("primitives", []):
            attrs = primitive.get("attributes", {})
            if "POSITION" not in attrs:
                continue

            position_accessor = int(attrs["POSITION"])
            positions_by_accessor.setdefault(
                position_accessor,
                read_position_accessor(gltf, bytes(bin_buf), position_accessor),
            )
            primitives.append(primitive)

    target_height = target_height if target_height is not None else HERO_TARGET_HEIGHTS.get(hero_id)
    if target_height is not None:
        _normalize_mesh_height(gltf, bin_buf, positions_by_accessor, target_height)

    mins, maxs = _bounds(positions_by_accessor)
    all_positions = [p for values in positions_by_accessor.values() for p in values]
    globals_ = joint_globals_from_mesh(all_positions, mins, maxs)

    first_joint_node = len(gltf.setdefault("nodes", []))
    gltf["nodes"].extend(_joint_nodes(globals_, first_joint_node))
    joint_indices = list(range(first_joint_node, first_joint_node + len(REQUIRED_BONES)))

    ibm_payload = bytearray()
    for pos in globals_:
        ibm_payload.extend(struct.pack("<16f", *_inverse_translation_matrix(pos)))

    ibm_view = _append_buffer_view(gltf, bin_buf, bytes(ibm_payload))
    ibm_accessor = _append_accessor(
        gltf,
        {"bufferView": ibm_view, "componentType": 5126, "count": len(REQUIRED_BONES), "type": "MAT4"},
    )

    skin_index = len(gltf.setdefault("skins", []))
    gltf["skins"].append({
        "joints": joint_indices,
        "inverseBindMatrices": ibm_accessor,
        "skeleton": first_joint_node,
    })

    max_joint = 0
    max_weight = 0.0
    for primitive in primitives:
        positions = positions_by_accessor[int(primitive["attributes"]["POSITION"])]
        joints_payload = bytearray()
        weights_payload = bytearray()
        for pos in positions:
            j, w = vertex_weights(pos, globals_)
            max_joint = max(max_joint, max(j))
            max_weight = max(max_weight, max(w))
            joints_payload.extend(struct.pack("<HHHH", *j))
            weights_payload.extend(struct.pack("<ffff", *w))

        joints_view = _append_buffer_view(gltf, bin_buf, bytes(joints_payload), target=34962)
        weights_view = _append_buffer_view(gltf, bin_buf, bytes(weights_payload), target=34962)
        joints_accessor = _append_accessor(
            gltf,
            {
                "bufferView": joints_view,
                "componentType": 5123,
                "count": len(positions),
                "type": "VEC4",
                "min": [0, 0, 0, 0],
                "max": [max_joint, 0, 0, 0],
            },
        )
        weights_accessor = _append_accessor(
            gltf,
            {
                "bufferView": weights_view,
                "componentType": 5126,
                "count": len(positions),
                "type": "VEC4",
                "min": [0.0, 0.0, 0.0, 0.0],
                "max": [max_weight, max_weight, max_weight, max_weight],
            },
        )
        primitive["attributes"]["JOINTS_0"] = joints_accessor
        primitive["attributes"]["WEIGHTS_0"] = weights_accessor

    for node in gltf.get("nodes", [])[:first_joint_node]:
        if "mesh" in node:
            node["skin"] = skin_index

    scene_index = int(gltf.get("scene", 0))
    scenes = gltf.setdefault("scenes", [{"nodes": []}])
    scene_nodes = scenes[scene_index].setdefault("nodes", [])
    if first_joint_node not in scene_nodes:
        scene_nodes.append(first_joint_node)

    gltf["buffers"][0]["byteLength"] = len(bin_buf)
    gltf["asset"]["generator"] = "arpg-dev/tools/assets/rig_canonical_hero.py"
    return write_glb(gltf, bytes(bin_buf))
