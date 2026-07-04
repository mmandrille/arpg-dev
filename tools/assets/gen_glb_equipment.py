"""Deterministic equipment and weapon GLB generators (extracted from gen_glb)."""
from __future__ import annotations


def _build_glb(color: tuple[float, float, float, float], parts: list[dict]) -> bytes:
    from tools.assets.gen_glb import _build_glb as build_static_glb

    return build_static_glb(color, parts, [])

def rusty_sword_glb() -> bytes:
    """Low-poly rusty one-handed sword, grip at origin, blade pointing +Y."""
    color = (0.45, 0.3, 0.18, 1.0)  # rusty brown
    parts = [
        {"name": "grip", "translation": [0.0, -0.08, 0.0], "scale": [0.05, 0.2, 0.05]},
        {"name": "guard", "translation": [0.0, 0.04, 0.0], "scale": [0.26, 0.05, 0.07]},
        {"name": "blade", "translation": [0.0, 0.5, 0.0], "scale": [0.07, 0.9, 0.02]},
    ]
    return _build_glb(color, parts)


def long_sword_glb() -> bytes:
    """Steel longsword with longer blade than rusty sword."""
    color = (0.62, 0.64, 0.68, 1.0)
    parts = [
        {"name": "grip", "translation": [0.0, -0.10, 0.0], "scale": [0.05, 0.24, 0.05]},
        {"name": "guard", "translation": [0.0, 0.06, 0.0], "scale": [0.30, 0.06, 0.08]},
        {"name": "blade", "translation": [0.0, 0.62, 0.0], "scale": [0.08, 1.15, 0.025]},
        {"name": "pommel", "translation": [0.0, -0.22, 0.0], "scale": [0.10, 0.08, 0.10]},
    ]
    return _build_glb(color, parts)


def rapier_glb() -> bytes:
    """Thin thrusting sword."""
    color = (0.70, 0.72, 0.78, 1.0)
    parts = [
        {"name": "grip", "translation": [0.0, -0.06, 0.0], "scale": [0.04, 0.18, 0.04]},
        {"name": "guard", "translation": [0.0, 0.04, 0.0], "scale": [0.22, 0.04, 0.05]},
        {"name": "blade", "translation": [0.0, 0.48, 0.0], "scale": [0.035, 0.95, 0.012]},
    ]
    return _build_glb(color, parts)


def equipment_shield_kite_glb() -> bytes:
    """Kite shield for paladin-style off-hand."""
    color = (0.50, 0.52, 0.58, 1.0)
    parts = [
        {"name": "kite_face", "translation": [0.0, 0.0, 0.05], "scale": [0.42, 0.72, 0.08]},
        {"name": "kite_boss", "translation": [0.0, 0.08, 0.10], "scale": [0.14, 0.14, 0.08]},
        {"name": "kite_grip", "translation": [0.0, -0.12, -0.06], "scale": [0.10, 0.48, 0.06]},
    ]
    return _build_glb(color, parts)


def equipment_shield_tower_glb() -> bytes:
    """Large tower shield."""
    color = (0.44, 0.40, 0.36, 1.0)
    parts = [
        {"name": "tower_face", "translation": [0.0, 0.0, 0.05], "scale": [0.58, 0.88, 0.08]},
        {"name": "tower_boss", "translation": [0.0, 0.12, 0.10], "scale": [0.18, 0.18, 0.08]},
        {"name": "tower_grip", "translation": [0.0, -0.18, -0.06], "scale": [0.12, 0.56, 0.06]},
    ]
    return _build_glb(color, parts)


def training_bow_glb() -> bytes:
    """Low-poly bow, grip at origin, bow/string standing along local Y."""
    color = (0.38, 0.24, 0.12, 1.0)
    parts = [
        {"name": "grip", "translation": [0.0, 0.0, 0.0], "scale": [0.08, 0.24, 0.08]},
        {"name": "upper_limb_inner", "translation": [-0.08, 0.38, 0.0], "scale": [0.05, 0.55, 0.05]},
        {"name": "lower_limb_inner", "translation": [-0.08, -0.38, 0.0], "scale": [0.05, 0.55, 0.05]},
        {"name": "upper_limb_tip", "translation": [-0.19, 0.78, 0.0], "scale": [0.05, 0.36, 0.05]},
        {"name": "lower_limb_tip", "translation": [-0.19, -0.78, 0.0], "scale": [0.05, 0.36, 0.05]},
        {"name": "string", "translation": [0.16, 0.0, 0.0], "scale": [0.018, 1.45, 0.018]},
    ]
    return _build_glb(color, parts)


def starter_staff_glb() -> bytes:
    """Low-poly two-handed staff, hand grip at origin, shaft pointing +Y."""
    color = (0.28, 0.19, 0.36, 1.0)
    parts = [
        {"name": "lower_cap", "translation": [0.0, -0.72, 0.0], "scale": [0.09, 0.12, 0.09]},
        {"name": "shaft_lower", "translation": [0.0, -0.34, 0.0], "scale": [0.055, 0.78, 0.055]},
        {"name": "grip_wrap", "translation": [0.0, 0.0, 0.0], "scale": [0.075, 0.24, 0.075]},
        {"name": "shaft_upper", "translation": [0.0, 0.45, 0.0], "scale": [0.055, 0.9, 0.055]},
        {"name": "head_cross", "translation": [0.0, 0.96, 0.0], "scale": [0.34, 0.055, 0.08]},
        {"name": "crystal_core", "translation": [0.0, 1.13, 0.0], "scale": [0.18, 0.22, 0.18]},
        {"name": "crystal_tip", "translation": [0.0, 1.32, 0.0], "scale": [0.10, 0.16, 0.10]},
    ]
    return _build_glb(color, parts)


def starter_axe_glb() -> bytes:
    """Low-poly two-handed axe, grip at origin, haft pointing +Y."""
    color = (0.34, 0.25, 0.18, 1.0)
    parts = [
        {"name": "butt_cap", "translation": [0.0, -0.54, 0.0], "scale": [0.10, 0.10, 0.10]},
        {"name": "haft_lower", "translation": [0.0, -0.20, 0.0], "scale": [0.07, 0.68, 0.07]},
        {"name": "grip_wrap", "translation": [0.0, 0.08, 0.0], "scale": [0.09, 0.28, 0.09]},
        {"name": "haft_upper", "translation": [0.0, 0.50, 0.0], "scale": [0.07, 0.72, 0.07]},
        {"name": "head_socket", "translation": [0.0, 0.88, 0.0], "scale": [0.15, 0.16, 0.12]},
        {"name": "upper_blade", "translation": [0.0, 0.88, 0.24], "scale": [0.055, 0.30, 0.32]},
        {"name": "lower_blade", "translation": [0.0, 0.88, -0.24], "scale": [0.055, 0.30, 0.32]},
        {"name": "top_spike", "translation": [0.0, 1.12, 0.0], "scale": [0.10, 0.22, 0.08]},
    ]
    return _build_glb(color, parts)


def equipment_helm_glb() -> bytes:
    """Medium helm centered on head socket origin."""
    color = (0.52, 0.54, 0.58, 1.0)
    parts = [
        {"name": "helmet_cap", "translation": [0.0, 0.12, 0.0], "scale": [0.62, 0.56, 0.62]},
        {"name": "helmet_brow", "translation": [0.0, -0.06, -0.16], "scale": [1.0, 0.12, 0.62]},
    ]
    return _build_glb(color, parts)


def equipment_chest_glb() -> bytes:
    """Chest armor with pauldrons, origin at torso center."""
    color = (0.48, 0.50, 0.55, 1.0)
    parts = [
        {"name": "chest_plate", "translation": [0.0, 0.0, 0.14], "scale": [0.86, 1.0, 0.28]},
        {"name": "left_pauldron", "translation": [-0.58, 0.34, 0.0], "scale": [0.32, 0.18, 0.34]},
        {"name": "right_pauldron", "translation": [0.58, 0.34, 0.0], "scale": [0.32, 0.18, 0.34]},
    ]
    return _build_glb(color, parts)


def equipment_gloves_glb() -> bytes:
    color = (0.42, 0.40, 0.38, 1.0)
    parts = [
        {"name": "left_glove", "translation": [-0.36, 0.0, 0.0], "scale": [0.42, 0.42, 0.36]},
        {"name": "right_glove", "translation": [0.36, 0.0, 0.0], "scale": [0.42, 0.42, 0.36]},
    ]
    return _build_glb(color, parts)


def equipment_boots_glb() -> bytes:
    color = (0.38, 0.34, 0.30, 1.0)
    parts = [
        {"name": "left_boot", "translation": [-0.52, 0.0, -0.08], "scale": [0.48, 0.62, 0.78]},
        {"name": "right_boot", "translation": [0.52, 0.0, -0.08], "scale": [0.48, 0.62, 0.78]},
    ]
    return _build_glb(color, parts)


def equipment_belt_glb() -> bytes:
    color = (0.36, 0.28, 0.20, 1.0)
    parts = [
        {"name": "belt_band", "translation": [0.0, 0.0, 0.0], "scale": [1.05, 0.24, 0.34]},
        {"name": "belt_buckle", "translation": [0.0, 0.0, -0.04], "scale": [0.24, 0.28, 0.40]},
    ]
    return _build_glb(color, parts)


def equipment_amulet_glb() -> bytes:
    color = (0.72, 0.62, 0.22, 1.0)
    parts = [
        {"name": "amulet_chain", "translation": [0.0, 0.0, 0.0], "scale": [0.34, 0.04, 0.34]},
        {"name": "amulet_gem", "translation": [0.0, -0.32, 0.0], "scale": [0.20, 0.24, 0.12]},
    ]
    return _build_glb(color, parts)


def equipment_ring_glb() -> bytes:
    color = (0.78, 0.72, 0.48, 1.0)
    parts = [
        {"name": "ring_band", "translation": [0.0, 0.0, 0.0], "scale": [0.32, 0.06, 0.32]},
        {"name": "ring_stone", "translation": [0.0, -0.30, 0.0], "scale": [0.14, 0.12, 0.10]},
    ]
    return _build_glb(color, parts)


def equipment_shield_glb() -> bytes:
    """Round shield for off-hand socket, face pointing +Z."""
    color = (0.46, 0.48, 0.52, 1.0)
    parts = [
        {"name": "round_shield_face", "translation": [0.0, 0.0, 0.04], "scale": [0.48, 0.48, 0.08]},
        {"name": "round_shield_boss", "translation": [0.0, 0.0, 0.09], "scale": [0.16, 0.16, 0.10]},
        {"name": "round_shield_grip", "translation": [0.0, 0.0, -0.07], "scale": [0.12, 0.62, 0.07]},
    ]
    return _build_glb(color, parts)


EQUIPMENT_TARGETS = {
    "client/assets/equipment/weapons/rusty_sword/rusty_sword.glb": rusty_sword_glb,
    "client/assets/equipment/weapons/long_sword/long_sword.glb": long_sword_glb,
    "client/assets/equipment/weapons/rapier/rapier.glb": rapier_glb,
    "client/assets/equipment/weapons/training_bow/training_bow.glb": training_bow_glb,
    "client/assets/equipment/weapons/starter_staff/starter_staff.glb": starter_staff_glb,
    "client/assets/equipment/weapons/starter_axe/starter_axe.glb": starter_axe_glb,
    "client/assets/equipment/armor/helm/helm.glb": equipment_helm_glb,
    "client/assets/equipment/armor/chest/chest.glb": equipment_chest_glb,
    "client/assets/equipment/armor/gloves/gloves.glb": equipment_gloves_glb,
    "client/assets/equipment/armor/boots/boots.glb": equipment_boots_glb,
    "client/assets/equipment/armor/belt/belt.glb": equipment_belt_glb,
    "client/assets/equipment/armor/amulet/amulet.glb": equipment_amulet_glb,
    "client/assets/equipment/armor/ring/ring.glb": equipment_ring_glb,
    "client/assets/equipment/armor/shield/shield.glb": equipment_shield_glb,
    "client/assets/equipment/armor/shield/kite_shield.glb": equipment_shield_kite_glb,
    "client/assets/equipment/armor/shield/tower_shield.glb": equipment_shield_tower_glb,
}
