from tools.assets.rig_quadruped_monster_glbs import (
    QUADRUPED_BONES,
    QUADRUPED_MONSTERS,
    ROOT,
    rig_quadruped_file,
    validate_target,
)
from tools.assets.validate_assets import parse_glb_skin_joint_names


def test_rig_quadruped_targets_expose_required_joints(tmp_path):
    for monster_id, (source_rel, _target_rel) in QUADRUPED_MONSTERS.items():
        target = tmp_path / f"{monster_id}.glb"
        rig_quadruped_file(ROOT / source_rel, target)

        assert parse_glb_skin_joint_names(target) == set(QUADRUPED_BONES)
        validate_target(target)
