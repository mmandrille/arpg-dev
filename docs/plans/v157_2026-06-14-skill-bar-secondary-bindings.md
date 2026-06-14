# v157 Plan — Skill-bar secondary bindings

Date: 2026-06-14
Spec: `docs/specs/v157_spec-skill-bar-secondary-bindings.md`

## Adoption checklist

- Existing Godot plugins/assets: reject. This is input/protocol/UI state, not a new UI package or art need.
- Existing local systems: adopt `set_skill_bindings_intent`, skill panel binding display, and current bot scenario runner.

## Tasks

1. Protocol and persistence contract
   - Expand skill binding arrays from 8 to 16 in shared schemas and examples.
   - Add a DB migration widening `character_skill_bindings` and `session_start_skill_bindings` slot checks to 0-15.
   - Update store normalization and session-start loading to use 16 slots.

2. Server sim
   - Raise `skillFunctionKeyCount` to 16.
   - Keep validation for every non-empty skill id.
   - Extend `TestSetSkillBindingsIntentUpdatesSnapshotAndDelta` for a secondary slot.

3. Client
   - Raise `SKILL_FUNCTION_KEY_COUNT` to 16.
   - Route Shift+F1 through Shift+F8 to slots 8-15.
   - Keep unmodified F1-F8 behavior for slots 0-7.
   - Update skill panel assigned labels for secondary slots.

4. Bot and tests
   - Add protocol bot actions/assertions for skill binding slots.
   - Add a scenario `67_skill_secondary_bindings.json`.
   - Update client skill panel/unit coverage if the local tests expose function-key routing.

5. Close-out
   - Write as-built notes.
   - Run targeted checks, then `make ci`.
