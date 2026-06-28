#!/usr/bin/env bash
# Godot headless client smoke test.
# Loads the Godot project, runs the headless smoke scene which logs in, creates
# a session, completes move/attack/pickup/equip, and verifies client state via
# the debug API. Skips gracefully (exit 0 with a notice) when the pinned Godot
# runtime is not installed, since CI runners may not yet provision it.
set -euo pipefail

GODOT="${GODOT:-godot}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLIENT_DIR="$ROOT/client"
# shellcheck source=quiet_helpers.sh
source "$ROOT/scripts/quiet_helpers.sh"
# shellcheck source=godot_ci_flags.sh
source "$ROOT/scripts/godot_ci_flags.sh"

export BASE_URL="${BASE_URL:-http://localhost:8888}"
export DEV_TOKEN="${DEV_TOKEN:-local-dev-token}"
export DEBUG_TOKEN="${DEBUG_TOKEN:-local-debug-token}"
export ARPG_EMAIL="${ARPG_EMAIL:-client-smoke+$(date -u +%Y%m%d%H%M%S)-$$@example.test}"

if ! command -v "$GODOT" >/dev/null 2>&1; then
  echo "[client-smoke] SKIP: Godot runtime '$GODOT' not found on PATH."
  echo "[client-smoke] Install Godot $(cat "$ROOT/.godot-version") and re-run, or set GODOT=/path/to/godot."
  exit 0
fi

echo "[client-smoke] Using Godot: $("$GODOT" --version 2>/dev/null | tail -1)"
echo "[client-smoke] BASE_URL=${BASE_URL}"

FAILED_GATES=()

run_godot_gate_with_timeout() {
  local timeout_s="$1"
  local log_file="$2"
  shift 2
  local pid waited=0
  "$@" >"$log_file" 2>&1 &
  pid=$!
  while kill -0 "$pid" >/dev/null 2>&1; do
    if (( waited >= timeout_s )); then
      kill "$pid" >/dev/null 2>&1 || true
      wait "$pid" >/dev/null 2>&1 || true
      return 124
    fi
    sleep 1
    waited=$((waited + 1))
  done
  wait "$pid"
}

require_running_server() {
  if curl -fsS "${BASE_URL}/readyz" >/dev/null 2>&1; then
    return 0
  fi

  echo "[client-smoke] FAIL: no reachable server at BASE_URL=${BASE_URL} (/readyz)"
  if [[ "${BASE_URL}" == "http://localhost:18081" ]] \
      && curl -fsS "http://localhost:8888/readyz" >/dev/null 2>&1; then
    echo "[client-smoke] Hint: a server is up on :8888 (make server default) but client-smoke uses TEST_BASE_URL :18081."
    echo "[client-smoke] Either: TEST_BASE_URL=http://localhost:8888 make client-smoke"
    echo "[client-smoke] Or:      make db-up && ARPG_ADDR=:18081 make server"
  else
    echo "[client-smoke] Start: make db-up && ARPG_ADDR=:18081 make server"
  fi
  exit 1
}

# Godot exits 0 even on a GDScript PARSE/load error when run via --script, so a
# bare exit-code check could pass silently on a broken gate. run_gate captures
# each gate's combined output and asserts its expected success sentinel appears,
# recording a clear failure if it does not.
run_gate() {
  local label="$1" sentinel="$2" script="$3"
  local gate_log gate_started elapsed timeout_s="${CLIENT_SMOKE_GATE_TIMEOUT_S:-180}"
  gate_log="$(mktemp -t arpg-client-gate.XXXXXX.log)"

  echo "[client-smoke] running $label"
  gate_started=$SECONDS
  set +e
  # shellcheck disable=SC2086
  run_godot_gate_with_timeout "$timeout_s" "$gate_log" "$GODOT" $GODOT_HEADLESS_FLAGS --path "$CLIENT_DIR" --script "$script"
  local status=$?
  set -e
  elapsed=$((SECONDS - gate_started))

  if [[ $status -eq 124 ]]; then
    echo "FAILED: $label (timed out after ${timeout_s}s)"
    show_log "$gate_log" "$label"
    FAILED_GATES+=("$label")
    rm -f "$gate_log"
    return 0
  fi

  if [[ $status -ne 0 ]]; then
    echo "FAILED: $label (exit $status, elapsed=${elapsed}s)"
    show_log "$gate_log" "$label"
    FAILED_GATES+=("$label")
    rm -f "$gate_log"
    return 0
  fi

  if ! grep -qF -- "$sentinel" "$gate_log"; then
    echo "FAILED: $label (missing sentinel: $sentinel, elapsed=${elapsed}s)"
    show_log "$gate_log" "$label"
    FAILED_GATES+=("$label")
    rm -f "$gate_log"
    return 0
  fi

  if is_quiet_mode; then
    echo "OK: $label (${elapsed}s)"
  else
    cat "$gate_log"
    echo "[client-smoke] OK: $label elapsed=${elapsed}s"
  fi

  rm -f "$gate_log"
}

finish_gates() {
  local suite_label="$1"
  if [[ "${#FAILED_GATES[@]}" -eq 0 ]]; then
    return 0
  fi
  echo "[client-smoke] FAILED $suite_label: ${#FAILED_GATES[@]} gate(s) failed:"
  local gate
  for gate in "${FAILED_GATES[@]}"; do
    echo "  - $gate"
  done
  exit 1
}

# Import resources once so headless --script runs cleanly.
echo "[client-smoke] running Godot asset import"
import_started=$SECONDS
# shellcheck disable=SC2086
"$GODOT" $GODOT_HEADLESS_FLAGS --path "$CLIENT_DIR" --import >/dev/null 2>&1 || true
echo "OK: Godot asset import ($((SECONDS - import_started))s)"

# 1. GDScript golden-fixture test (server-independent; ADR D6 / acceptance #7).
run_gate "GDScript golden test" "[gdtest] PASS" res://tests/test_golden.gd

# 2. Item visual resolution test (server-independent; acceptance #14).
run_gate "GDScript item visual resolution test" "[gdtest] PASS" res://tests/test_item_visuals.gd
run_gate "GDScript projectile visual test" "[gdtest] PASS: test_projectile_visuals" res://tests/test_projectile_visuals.gd

# 2b. Rig gate: both GLBs import as skinned Skeleton3D (spec §10 fail-fast).
run_gate "GDScript rig gate" "[rig-gate] PASS" res://tools/inspect_rig.gd

# 2c. Animation controller + rigged scene test (server-independent).
run_gate "GDScript animation test" "[gdtest] PASS: animation controller + scenes" res://tests/test_animation.gd
run_gate "GDScript model viewer test" "[gdtest] PASS: test_model_viewer" res://tests/test_model_viewer.gd

# 2d. Bot scenario runner unit tests (server-independent).
run_gate "GDScript client bot unit test" "[gdtest] PASS: test_client_bot" res://tests/test_client_bot.gd
run_gate "GDScript bot entity distance test" "[gdtest] PASS: test_bot_entity_distance" res://tests/test_bot_entity_distance.gd
run_gate "GDScript bot facade unit test" "[gdtest] PASS: test_bot_facade" res://tests/test_bot_facade.gd

# 2e. Co-op local/remote player handling test (server-independent; v33).
run_gate "GDScript co-op client unit test" "[gdtest] PASS: test_coop_client" res://tests/test_coop_client.gd
run_gate "GDScript rogue presentation test" "[gdtest] PASS: test_rogue_presentation" res://tests/test_rogue_presentation.gd

# 2f. Waypoint panel scroll/layout test (server-independent; v19).
run_gate "GDScript waypoint panel test" "[gdtest] PASS: test_waypoint_panel" res://tests/test_waypoint_panel.gd
run_gate "GDScript quest/elite objective state test" "[gdtest] PASS: test_quest_elite_objective_state" res://tests/test_quest_elite_objective_state.gd
run_gate "GDScript quest journal panel test" "[gdtest] PASS: test_quest_journal_panel" res://tests/test_quest_journal_panel.gd
run_gate "GDScript elite objective tracker test" "[gdtest] PASS: test_elite_objective_tracker" res://tests/test_elite_objective_tracker.gd
run_gate "GDScript discovery minimap test" "[gdtest] PASS: test_discovery_minimap" res://tests/test_discovery_minimap.gd

# 2g. Sustained click hold state (server-independent; v27).
run_gate "GDScript sustained input test" "[gdtest] PASS: test_sustained_input" res://tests/test_sustained_input.gd
run_gate "GDScript path reject backoff test" "[gdtest] PASS: test_path_reject_backoff" res://tests/test_path_reject_backoff.gd

# 2h. Force-stand directional attack helpers (server-independent; v37).
run_gate "GDScript directional attack input test" "[gdtest] PASS: test_directional_attack_input" res://tests/test_directional_attack_input.gd
run_gate "GDScript text input focus guard test" "[gdtest] PASS: test_text_input_focus_guard" res://tests/test_text_input_focus_guard.gd

# 2i. Shop panel render and intent payload test (server-independent; v41).
run_gate "GDScript shop panel test" "[gdtest] PASS: test_shop_panel" res://tests/test_shop_panel.gd
run_gate "GDScript shop tooltip stability test" "[gdtest] PASS: test_shop_tooltip_stability" res://tests/test_shop_tooltip_stability.gd
run_gate "GDScript blacksmith panel test" "[gdtest] PASS: test_blacksmith_panel" res://tests/test_blacksmith_panel.gd
run_gate "GDScript set collection panel test" "[gdtest] PASS: test_set_collection_panel" res://tests/test_set_collection_panel.gd
run_gate "GDScript mercenary panel test" "[gdtest] PASS: test_mercenary_panel" res://tests/test_mercenary_panel.gd

# 2j. Town service bridge routing test (server-independent; v127).
run_gate "GDScript town service bridge test" "[gdtest] PASS: test_town_service_bridge" res://tests/test_town_service_bridge.gd
run_gate "GDScript inventory transfer router test" "[gdtest] PASS: test_inventory_transfer_router" res://tests/test_inventory_transfer_router.gd
run_gate "GDScript inventory panel test" "[gdtest] PASS: test_inventory_panel" res://tests/test_inventory_panel.gd

# 2k. Stash panel render and intent payload test (server-independent; v50).
run_gate "GDScript stash panel test" "[gdtest] PASS: test_stash_panel" res://tests/test_stash_panel.gd

# 2l. Skill point panel and skill bar tests (server-independent; v44).
run_gate "GDScript text catalog test" "[gdtest] PASS: test_text_catalog" res://tests/test_text_catalog.gd
run_gate "GDScript enemy health bar settings test" "[gdtest] PASS: test_enemy_health_bar_settings" res://tests/test_enemy_health_bar_settings.gd
run_gate "GDScript audio settings test" "[gdtest] PASS: test_audio_settings" res://tests/test_audio_settings.gd
run_gate "GDScript client audio controller test" "[gdtest] PASS: test_client_audio_controller" res://tests/test_client_audio_controller.gd
run_gate "GDScript character bar test" "[gdtest] PASS: test_character_bar" res://tests/test_character_bar.gd
run_gate "GDScript character stats panel test" "[gdtest] PASS: test_character_stats_panel" res://tests/test_character_stats_panel.gd
run_gate "GDScript skill rules loader test" "[gdtest] PASS: test_skill_rules_loader" res://tests/test_skill_rules_loader.gd
run_gate "GDScript skills panel test" "[gdtest] PASS: test_skills_panel" res://tests/test_skills_panel.gd
run_gate "GDScript skill bar test" "[gdtest] PASS: test_skill_bar" res://tests/test_skill_bar.gd
run_gate "GDScript status effects bar test" "[gdtest] PASS: test_status_effects_bar" res://tests/test_status_effects_bar.gd
run_gate "GDScript status effect presentation test" "[gdtest] PASS: test_status_effect_presentation" res://tests/test_status_effect_presentation.gd
run_gate "GDScript aura soft lights test" "[gdtest] PASS: test_aura_soft_lights" res://tests/test_aura_soft_lights.gd

# 2l. Boss health bar render/state test (server-independent; v53).
run_gate "GDScript boss health bar test" "[gdtest] PASS: test_boss_health_bar" res://tests/test_boss_health_bar.gd

# 2m. Delta and snapshot state-mutation unit tests (server-independent; v53).
run_gate "GDScript net client test" "[gdtest] PASS: test_net_client" res://tests/test_net_client.gd
run_gate "GDScript delta apply test" "[gdtest] PASS: test_delta_apply" res://tests/test_delta_apply.gd

# 2n. Loot label rarity filter unit tests (server-independent; v153).
run_gate "GDScript loot label filter test" "[gdtest] PASS: test_loot_label_filter" res://tests/test_loot_label_filter.gd
run_gate "GDScript loot filter ground item test" "[gdtest] PASS: test_loot_filter_ground_items" res://tests/test_loot_filter_ground_items.gd
run_gate "GDScript loot node factory test" "[gdtest] PASS: test_loot_node_factory" res://tests/test_loot_node_factory.gd
run_gate "GDScript impact sparks test" "[gdtest] PASS: test_impact_sparks" res://tests/test_impact_sparks.gd
run_gate "GDScript combat outcome punch test" "[gdtest] PASS: test_combat_outcome_punch" res://tests/test_combat_outcome_punch.gd
run_gate "GDScript skill rank intensity test" "[gdtest] PASS: test_skill_rank_intensity" res://tests/test_skill_rank_intensity.gd
run_gate "GDScript look-and-feel polish test" "[gdtest] PASS: test_look_and_feel_polish" res://tests/test_look_and_feel_polish.gd

# 2o. World-detail render + combat-feel unit tests (server-independent; v295-v308).
# These existed on disk but were never registered, so they ran zero times in CI.
run_gate "GDScript ground/wall factories test" "[gdtest] PASS: test_factories" res://tests/test_factories.gd
run_gate "GDScript dungeon depth lighting test" "[gdtest] PASS: test_dungeon_depth_lighting" res://tests/test_dungeon_depth_lighting.gd
run_gate "GDScript town night lighting test" "[gdtest] PASS: test_town_night_lighting" res://tests/test_town_night_lighting.gd
run_gate "GDScript fog-of-war overlay test" "[gdtest] PASS: test_fog_of_war_overlay" res://tests/test_fog_of_war_overlay.gd
run_gate "GDScript fog LOS shadow cache test" "[gdtest] PASS: test_fog_los_shadow_cache" res://tests/test_fog_los_shadow_cache.gd
run_gate "GDScript hero light source test" "[gdtest] PASS: test_hero_light_source" res://tests/test_hero_light_source.gd
run_gate "GDScript movement visual smoothing test" "[gdtest] PASS: test_movement_visual_smoothing" res://tests/test_movement_visual_smoothing.gd
run_gate "GDScript entity tick smoothing test" "[gdtest] PASS: test_entity_tick_smoothing" res://tests/test_entity_tick_smoothing.gd
run_gate "GDScript combat feel presentation loader test" "[gdtest] PASS: test_combat_feel_presentation_loader" res://tests/test_combat_feel_presentation_loader.gd
run_gate "GDScript loot tick smoothing test" "[gdtest] PASS: test_loot_tick_smoothing" res://tests/test_loot_tick_smoothing.gd
run_gate "GDScript interactable tick smoothing test" "[gdtest] PASS: test_interactable_tick_smoothing" res://tests/test_interactable_tick_smoothing.gd
run_gate "GDScript attack animation scaling test" "[gdtest] PASS: test_attack_animation_scaling" res://tests/test_attack_animation_scaling.gd
run_gate "GDScript projectile tick smoothing test" "[gdtest] PASS: test_projectile_tick_smoothing" res://tests/test_projectile_tick_smoothing.gd
run_gate "GDScript mobility skill presentation test" "[gdtest] PASS: test_mobility_skill_presentation" res://tests/test_mobility_skill_presentation.gd
run_gate "GDScript dungeon torch placement test" "[gdtest] PASS: test_dungeon_torch_placement" res://tests/test_dungeon_torch_placement.gd
run_gate "GDScript player movement feel test" "[gdtest] PASS: test_player_movement_feel" res://tests/test_player_movement_feel.gd
run_gate "GDScript melee lunge presentation test" "[gdtest] PASS: test_melee_lunge_presentation" res://tests/test_melee_lunge_presentation.gd
run_gate "GDScript command retarget grace test" "[gdtest] PASS: test_command_retarget_grace" res://tests/test_command_retarget_grace.gd
run_gate "GDScript movement input presenter test" "[gdtest] PASS: test_movement_input_presenter" res://tests/test_movement_input_presenter.gd

# 2p. Camera mode presentation loader unit tests (server-independent; v329).
run_gate "GDScript camera mode settings test" "[gdtest] PASS: test_camera_mode_settings" res://tests/test_camera_mode_settings.gd
run_gate "GDScript window display mode settings test" "[gdtest] PASS: test_window_display_mode_settings" res://tests/test_window_display_mode_settings.gd
run_gate "GDScript crosshair target system test" "[gdtest] PASS: test_crosshair_target_system" res://tests/test_crosshair_target_system.gd
run_gate "GDScript panel intent input test" "[gdtest] PASS: test_panel_intent_input" res://tests/test_panel_intent_input.gd

if [[ "${CLIENT_UNIT_ONLY:-}" == "1" ]]; then
  finish_gates "client unit"
  echo "[client-unit] PASS"
  exit 0
fi

require_running_server

# 3. Headless slice smoke against the running server.
run_gate "headless slice smoke" "[smoke] PASS" res://scripts/smoke.gd
finish_gates "client smoke"
echo "[client-smoke] PASS"
