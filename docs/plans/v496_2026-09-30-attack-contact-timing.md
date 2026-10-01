# v496 Plan — Attack contact timing

Status: Complete; combined `make ci` passed in 10m33s.
Goal: Align local basic-attack motion and authoritative outcomes in live melee and ranged play.
Architecture: The Godot client starts a presentation-only swing on input; existing server events remain the sole source of hit, miss, block, and damage. Presentation timing lives in schema-backed `shared/assets/` data. The final timing and visual evidence was repeated after v495 integration.
Tech stack: Godot/GDScript, shared JSON and schemas, existing live-client bot.

## Baseline and shortcut decision

- Reuse v464 local input and retarget flow, v465 target outcome punch, v475 KayKit attack clips, and v492 particles. The existing 103 client scenario proves movement dispatch but starts outside melee range, so it does not prove swing recovery. The existing 104 scenario covers a bow hit and kill but not miss/block or frame timing.
- **Borrow** the in-repo KayKit clips, audio cue, and VFX catalog. **Reject** external assets and plugins.
- Record real-client timestamps for input, local clip start, clip contact, result arrival, target feedback, and locomotion return for a melee and bow attack, including miss or block. Save normal-speed play-camera frame sequences and note renderer, quality tier, sample count, and limitations. Repeat on v495 before final timing or performance conclusions.

## File map

| Action | Path | Responsibility |
|--------|------|----------------|
| Modify | `client/scripts/combat_local_attack_presentation.gd` | Keep a lethal hit's kill event from starting a second swing |
| Modify | `client/tests/test_sustained_input.gd` | Prove one local swing for a damage-plus-kill event pair |
| Modify | `tools/bot/scenarios/client/103_combat_input_flow_polish.json` | Reach a real melee swing before retarget and check the lunge count |
| Modify | `tools/bot/scenarios/client/104_combat_impact_confirmation.json` | Replace a race-prone pickup-event wait with inventory-state proof |
| Modify | `client/scripts/bot_assertion_handlers.gd` | Assert a maximum local lunge count in the live bot |
| Modify | `client/scripts/bot_step_catalog.gd`, `client/scripts/bot_wait_handlers.gd` | Wait for equipped state without missing an earlier equip event |
| Modify | `client/scripts/main.gd`, `client/scripts/gameplay_feedback_presentation.gd` | Keep blocked monster damage out of hit audio, target reaction, and hit VFX; expose contact and input state to the bot |
| Create | `client/scripts/bot_combat_contact_assertions.gd` | Assert event-specific contact feedback and bounded buffered input |
| Modify | `client/scripts/bot_scenario_runner.gd`, `client/tests/test_client_bot.gd` | Add target reaction upper bound and exercise focused bot assertions |
| Create | `tools/bot/scenarios/client/109_attack_block_contact.json` | Prove blocked contact and rapid retarget in the combat-stat lab |
| Create | `docs/researchs/v496_attack-contact-baseline.md` | Record current partial live baseline and proof limits |
| Create | `docs/as-built/v496_attack-contact-timing.md` | Record measured before/after and acceptance limits when slice finishes |
| Modify | `PROGRESS.md`, `docs/progress/slice-lifecycle.md` | Close lifecycle only after v495 and final gates |

## Maintenance ratchet

Target: new source/test/tool files stay at or below 600 lines; touched grandfathered files do not grow beyond their allowance.

`client/scripts/bot_assertion_handlers.gd` and `client/scripts/main.gd` are grandfathered above 600 lines. Focused checks live in `bot_combat_contact_assertions.gd`, keeping both coordinators within their 25-line allowances. No server or protocol file is touched.

Verification: `make maintainability`.

## Task 1 — Capture current live baseline

Files: `docs/researchs/v496_attack-contact-baseline.md`, `.artifacts/` captures.

- [x] Run existing 103 and 104 live-client scenarios, and measure the melee/bow clip lengths and event order.
- [x] Capture melee and ranged input-to-clip, contact, result, effect, and locomotion timings, including a block; mark unavailable measurements explicitly.

Verify: `HEADLESS=1 make bot-visual scenario=103_combat_input_flow_polish` and `HEADLESS=1 make bot-visual scenario=104_combat_impact_confirmation`.

## Task 2 — Independent duplicate-start correction

Files: `client/scripts/combat_local_attack_presentation.gd`, `client/tests/test_sustained_input.gd`.

- [x] For a lethal basic hit, consume the `monster_damaged` outcome without replaying the locally started swing; treat its subsequent `monster_killed` event as kill feedback only.
- [x] Prove one local swing for the damage-plus-kill pair, and preserve server-started fallback for nonlethal damage, miss, and block.
- [x] Keep authoritative confirmed-hit reaction, combat text, kill cue, and v492 particles unchanged.

Verify: `godot --headless --path client --script res://tests/test_sustained_input.gd`; `make maintainability`.

## Task 3 — Live bot and visual proof

Files: `tools/bot/scenarios/client/103_combat_input_flow_polish.json`, `tools/bot/scenarios/client/104_combat_impact_confirmation.json`, `client/scripts/bot_assertion_handlers.gd`, `client/scripts/bot_step_catalog.gd`, `client/scripts/bot_wait_handlers.gd`.

- [x] Ensure the scenario reaches a real melee swing and asserts exactly one local lunge after its lethal result, before floor retarget.
- [x] Make the bow scenario's pickup and equip waits robust to events arriving before their wait steps, then rerun the bow hit/kill regression.
- [x] Reproduce a blocked `monster_damaged` event incorrectly playing target hit feedback; gate hit audio/reaction/VFX on confirmed contact while retaining `BLOCK` text. Focused scenario 109 proves the block event has zero added reaction feedback.
- [x] Finish scenario 109's rapid buffered-click bound and floor-retarget order proof.
- [x] Save matched before/after play-camera frames and inspect Balanced and Performance outcome cues; record the normal-speed Balanced sequence.

Verify: `HEADLESS=1 make bot-visual scenario=103_combat_input_flow_polish`; `HEADLESS=1 make bot-visual scenario=104_combat_impact_confirmation`; `BOT_STEP_DELAY=0 HEADLESS=1 make bot-visual scenario=109_attack_block_contact`.

## Task 4 — v495-dependent timing pass and closeout

Files: `shared/assets/attack_presentation.v0.json` and schema if measured clip offsets need tuning; focused client tests/scenario; `docs/as-built/v496_attack-contact-timing.md`, `PROGRESS.md`, `docs/progress/slice-lifecycle.md`.

- [x] After v495 lands, repeat melee/ranged live timing and compare input, clip contact, authoritative outcome, effect, and locomotion.
- [x] Correct the reproducible lethal-event replay and blocked-hit classification; keep aligned clip timing unchanged.
- [x] Assert hit/block classification, rapid input bounds, and no duplicate start in the live bot. Inspect real-renderer play-camera frames for both quality tiers; miss remains a focused-test case.
- [x] Complete as-built evidence; lifecycle closeout belongs to the combined integration. The review/refactor gate is deferred until this parallel batch finishes.

Verify: `make validate-shared`; focused client tests; `make client-unit`; `HEADLESS=1 make bot-visual scenario=103_combat_input_flow_polish`; `HEADLESS=1 make bot-visual scenario=104_combat_impact_confirmation`; `make maintainability`. The coordinating task runs one `make ci` after combining slices on main.

## Final verification

- [x] `make validate-shared`
- [x] `make client-unit` on the isolated v493 checkout
- [x] `make maintainability` on the isolated v493 checkout
- [x] Named live-client scenarios above on the isolated v493 checkout
- [x] Real-renderer Balanced and Performance capture review
- [x] Coordinating task ran the final `make ci` on the combined changes after integration

The coordinating task completed the final combined gate, lifecycle closeout, and transfer to main. The engineering review and refactor remain next-work items. No protocol or combat-rule changes were needed.

## Opt-in trace and normal-speed capture follow-up

The v496 trace delta adds `client/scripts/attack_contact_trace.gd`,
`client/tests/test_attack_contact_trace.gd`, and
`tools/bot/scenarios/client/v496_attack_contact_timeline.json`. The scenario
uses the existing combat-stat lab to record a sword and bow hit against its
durable target. The trace records client monotonic time and process frames for
input, local swing, authoritative result, feedback calls, and locomotion, plus
bounded frame samples and a visible time/clip overlay.

Use an external screen recorder for the normal-speed frame sequence. The exact
Balanced/Performance commands and evidence table are in
`docs/researchs/v496_attack-contact-capture-protocol.md`. The unit/static
checks do not close capture acceptance; run the windowed scenario and inspect
the actual frames after the host is released.
