"""Cross-file consistency checks for the dungeon_obstacles golden fixture.

Extracted from validate_shared.py (mirroring validate_boss_patterns and the other
sibling validators) so the golden <-> dungeon_generation relationship can be unit
tested in isolation. The caller passes the already-loaded rule and golden dicts and
a ``report`` object exposing ``.ok(label)`` and ``.fail(label, detail)``; this module
imports nothing from validate_shared.
"""
from __future__ import annotations


def validate_dungeon_obstacle_goldens(report, dungeon_generation: dict, dungeon_obstacles_golden: dict) -> None:
    """Validate the dungeon_obstacles golden against dungeon_generation.

    Two layers:
      1. The v40 wall-layout contract (generated floor, matching floor_size, named
         shape families, positive wall floor, represented shape families).
      2. The golden floors (minimum water/hole counts and solid kinds) stay
         achievable under the dungeon_generation obstacle config, so a balance edit
         to generation cannot silently outrun the golden. The JSON schema checks
         presence/type; this checks the cross-file relationship a schema cannot.
    """
    floor_size = dungeon_generation.get("floor_size")
    obstacle_gen = dungeon_generation.get("obstacle_generation", {})
    obstacle_expected = dungeon_obstacles_golden.get("expected", {})
    obstacle_floor = obstacle_expected.get("floor_size", {})
    obstacle_shapes = set(obstacle_expected.get("shape_families", []))
    obstacle_walls = obstacle_expected.get("walls", [])
    generated_walls = [wall for wall in obstacle_walls if wall.get("source") == "generated"]
    if dungeon_obstacles_golden.get("level", 0) >= 0:
        report.fail("dungeon_obstacles golden", "level must be a generated dungeon floor")
    elif obstacle_floor != floor_size:
        report.fail("dungeon_obstacles golden", "floor_size must match dungeon_generation floor_size")
    elif len(obstacle_shapes) < 2:
        report.fail("dungeon_obstacles golden", "must name at least two shape families")
    elif obstacle_expected.get("minimum_generated_wall_count", 0) <= 0:
        report.fail("dungeon_obstacles golden", "minimum_generated_wall_count must be positive")
    elif len(generated_walls) > 0 and not obstacle_shapes.intersection({wall.get("shape_family") for wall in generated_walls}):
        report.fail("dungeon_obstacles golden", "generated wall shape_family must be represented in shape_families")
    else:
        report.ok("dungeon_obstacles golden declares v40 wall-layout contract")

    solid_weights = obstacle_gen.get("solid_kind_weights", {})
    consistent = True
    for kind, field in (("water", "minimum_water_count"), ("holes", "minimum_hole_count")):
        floor = obstacle_expected.get(field, 0)
        block = obstacle_gen.get(kind, {})
        cap = block.get("target_count", {}).get("max", 0)
        if floor > 0 and not block.get("enabled", False):
            report.fail("dungeon_obstacles golden vs generation", f"{field} > 0 but obstacle_generation.{kind}.enabled is false")
            consistent = False
        elif floor > cap:
            report.fail("dungeon_obstacles golden vs generation", f"{field} ({floor}) exceeds obstacle_generation.{kind}.target_count.max ({cap})")
            consistent = False
    unknown_solid = [k for k in obstacle_expected.get("solid_kinds", []) if solid_weights.get(k, 0) <= 0]
    if unknown_solid:
        report.fail("dungeon_obstacles golden vs generation", f"solid_kinds {unknown_solid} have no positive weight in obstacle_generation.solid_kind_weights")
        consistent = False
    if consistent:
        report.ok("dungeon_obstacles golden floors are achievable under dungeon_generation obstacle config")
