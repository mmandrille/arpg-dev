# v157 Spec — Skill-bar secondary bindings

Date: 2026-06-14
Status: Draft

## Goal

Give each character a second skill binding row without changing existing F-key and right-click behavior. Primary bindings remain F1-F8. Secondary bindings are Shift+F1 through Shift+F8 and share the same authoritative skill binding contract.

## Player-facing behavior

- F1-F8 continue to select the primary bound skill and assign the hovered skill when the skills panel is open.
- Shift+F1 through Shift+F8 use binding slots 8-15.
- The skills panel debug state reports primary labels as `F1`...`F8` and secondary labels as `S-F1`...`S-F8`.
- The selected right-click skill remains the active skill shown on the skill bar.

## Contract

- `set_skill_bindings_intent.payload.function_keys` expands from 8 to 16 fixed entries.
- `session_snapshot.skill_bindings.function_keys` and `state_delta.skill_bindings_update.skill_bindings.function_keys` expose 16 fixed entries.
- Slots 0-7 are primary, slots 8-15 are secondary.
- Older callers that submit fewer entries are normalized with empty trailing bindings.

## Out of scope

- New skill balance, new skills, or cooldown changes.
- A redesigned multi-slot skill bar.
- New mouse-button bindings beyond the existing right-click selected skill.

## Verification

- Shared schema validation accepts 16-entry examples.
- Go sim test proves primary and secondary bindings normalize, snapshot, and delta.
- Client unit coverage proves Shift+F selection uses secondary slots.
- Protocol bot scenario proves the server accepts and snapshots secondary bindings.
