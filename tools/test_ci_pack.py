from tools.bot.ci_pack import (
    CI_SELECTOR,
    client_pack_ids,
    protocol_pack_ids,
    validate_ci_pack,
)
from tools.bot.run import load_scenarios, select_scenarios


def test_ci_pack_validation_passes():
    validate_ci_pack()


def test_ci_pack_sizes_are_curated_subset():
    protocol = load_scenarios()
    selected = select_scenarios(protocol, CI_SELECTOR)
    assert len(selected) == len(protocol_pack_ids())
    assert len(selected) < len(protocol)


def test_ci_selector_excludes_extended_only_scenarios():
    protocol = load_scenarios()
    selected_ids = {scenario.id for scenario in select_scenarios(protocol, CI_SELECTOR)}
    assert "teleporter_lab" not in selected_ids
    assert "vertical_slice" in selected_ids


def test_ci_pack_client_paths_resolve():
    from tools.bot.ci_pack import resolve_client_pack_paths

    paths = resolve_client_pack_paths()
    assert len(paths) == len(client_pack_ids())
