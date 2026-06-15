from __future__ import annotations


def validate_boss_patterns(
    report,
    boss_patterns: dict,
    boss_pattern_golden: dict,
    boss_floor: dict,
    boss_floor_golden: dict,
    boss_templates: dict,
) -> None:
    min_telegraph_ticks = int(boss_patterns["minimum_telegraph_ticks"])
    for pattern_id, pattern in boss_patterns["patterns"].items():
        pattern_failed = False
        previous_telegraph = None
        for index, phase in enumerate(pattern["phases"]):
            if phase["duration_ticks"] <= 0:
                report.fail("boss pattern duration", f"{pattern_id}[{index}] must be positive")
                pattern_failed = True
                break
            if phase["kind"] == "telegraph":
                if phase["duration_ticks"] < min_telegraph_ticks:
                    report.fail("boss pattern telegraph duration", f"{pattern_id}[{index}] below {min_telegraph_ticks}")
                    pattern_failed = True
                    break
                previous_telegraph = phase
                continue
            if "damage" not in phase:
                continue
            if previous_telegraph is None:
                report.fail("boss pattern telegraph guarantee", f"{pattern_id}[{index}] damages without prior telegraph")
                pattern_failed = True
                break
            damage = phase["damage"]
            if damage["max"] < damage["min"]:
                report.fail("boss pattern damage", f"{pattern_id}[{index}] max must be >= min")
                pattern_failed = True
                break
            if phase.get("shape") != previous_telegraph.get("hit_shape"):
                report.fail("boss pattern hit predicate", f"{pattern_id}[{index}] active shape must match telegraph hit_shape")
                pattern_failed = True
                break
            if phase.get("radius") != previous_telegraph.get("radius"):
                report.fail("boss pattern hit predicate", f"{pattern_id}[{index}] active radius must match telegraph radius")
                pattern_failed = True
                break
        if not pattern_failed:
            report.ok(f"boss pattern {pattern_id} satisfies telegraph guarantee")

    if boss_floor_golden["level"] != boss_floor["first_level"]:
        report.fail("boss_floor golden", "level must match boss_floor.first_level")
    elif boss_floor_golden["floor_size"] != boss_floor["floor_size"]:
        report.fail("boss_floor golden", "floor_size must match boss_floor rules")
    else:
        expected = boss_floor_golden["expected"]
        template_id = expected["boss"]["template_id"]
        template = boss_templates["bosses"].get(template_id)
        if template is None:
            report.fail("boss_floor golden", f"unknown boss template {template_id}")
        elif expected["locked_reason"] != boss_floor["locked_exit_reason"]:
            report.fail("boss_floor golden", "locked reason must match boss_floor rules")
        elif expected["stairs_down_initial_state"] != "locked" or expected["teleporter_initial_state"] != "absent":
            report.fail("boss_floor golden", "initial exit states must be locked/absent")
        elif expected["boss"]["base_monster_def_id"] != template["base_monster_def_id"]:
            report.fail("boss_floor golden", "boss base_monster_def_id mismatch")
        elif expected["boss"]["visual_model"] not in template["visual"].get("model_pool", [template["visual"]["model"]]) or expected["boss"]["visual_scale"] != template["visual"]["scale"]:
            report.fail("boss_floor golden", "boss visual metadata mismatch")
        else:
            report.ok("boss_floor golden matches boss-floor rules")

    pattern = boss_patterns["patterns"].get(boss_pattern_golden["pattern_id"])
    if pattern is None:
        report.fail("boss_pattern_timeline golden", f"unknown pattern {boss_pattern_golden['pattern_id']}")
    elif boss_pattern_golden["minimum_telegraph_ticks"] != boss_patterns["minimum_telegraph_ticks"]:
        report.fail("boss_pattern_timeline golden", "minimum telegraph ticks mismatch")
    elif boss_pattern_golden["cooldown_ticks"] != pattern["cooldown_ticks"]:
        report.fail("boss_pattern_timeline golden", "cooldown ticks mismatch")
    elif len(boss_pattern_golden["timeline"]) != len(pattern["phases"]):
        report.fail("boss_pattern_timeline golden", "phase count mismatch")
    else:
        failed_pattern_golden = False
        cursor = 0
        for expected_phase, rule_phase in zip(boss_pattern_golden["timeline"], pattern["phases"]):
            duration = rule_phase["duration_ticks"]
            if expected_phase["kind"] != rule_phase["kind"]:
                report.fail("boss_pattern_timeline golden", f"phase {expected_phase['phase_index']} kind mismatch")
                failed_pattern_golden = True
                break
            if expected_phase["start_tick"] != cursor or expected_phase["end_tick"] != cursor + duration - 1:
                report.fail("boss_pattern_timeline golden", f"phase {expected_phase['phase_index']} tick boundary mismatch")
                failed_pattern_golden = True
                break
            if expected_phase["duration_ticks"] != duration:
                report.fail("boss_pattern_timeline golden", f"phase {expected_phase['phase_index']} duration mismatch")
                failed_pattern_golden = True
                break
            for key in ("telegraph_type", "hit_shape", "shape", "radius", "damage"):
                if key in expected_phase and expected_phase[key] != rule_phase.get(key):
                    report.fail("boss_pattern_timeline golden", f"phase {expected_phase['phase_index']} {key} mismatch")
                    failed_pattern_golden = True
                    break
            if failed_pattern_golden:
                break
            cursor += duration
        if not failed_pattern_golden:
            dodge = boss_pattern_golden["dodge_case"]
            if not dodge["player_starts_in_contact"] or dodge["break_contact_before_tick"] >= boss_pattern_golden["timeline"][1]["start_tick"]:
                report.fail("boss_pattern_timeline golden", "dodge case must break contact before active starts")
            elif dodge["expected_damage"] != 0:
                report.fail("boss_pattern_timeline golden", "dodge case expected damage must be 0")
            else:
                report.ok("boss_pattern_timeline golden matches boss pattern rules")
