from tools.assets.rig_hero_glbs import REQUIRED_BONES, parse_glb
from tools.assets.rig_monster_glbs import BIPED_MONSTERS, ROOT, rig_monster_file, validate_target
from tools.assets.validate_assets import parse_glb_skin_joint_names


def test_rig_monster_targets_expose_required_biped_joints(tmp_path):
    for monster_id, (source_rel, _target_rel) in BIPED_MONSTERS.items():
        target = tmp_path / f"{monster_id}.glb"
        rig_monster_file(ROOT / source_rel, target)

        assert parse_glb_skin_joint_names(target) == set(REQUIRED_BONES)
        validate_target(target)

        parsed = parse_glb(target.read_bytes())
        assert parsed.gltf["skins"]
        assert parsed.gltf["asset"]["generator"] == "arpg-dev/tools/assets/rig_hero_glbs.py"
