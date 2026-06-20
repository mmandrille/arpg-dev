# v297 Spec: Client-Side Windup / Recovery Presentation

Status: Complete
Date: 2026-06-19
Codename: `client-side-windup-recovery-presentation`
Baseline: v296 `attack-move-sticky-targeting`

## Purpose

Make basic attacks feel immediate while preserving server authority. The client should start the
local player swing/audio/recovery presentation when it sends an existing basic `action_intent`, then
avoid replaying that same local swing when the authoritative hit/miss result arrives.

This is the third Movement / Combat Fluidity autoloop slice. It builds on v295 input buffering and
v296 attack move/sticky targeting.

## Non-goals

- Do not change server combat legality, hit timing, damage, cooldown, death, loot, or protocol.
- Do not delay or predict damage numbers; authoritative events still own hit/miss/damage text and
  target reactions.
- Do not add hit stop, impact flash, camera shake, movement smoothing, command retarget grace, or
  melee lunge in this slice.
- Do not add new animation assets or external plugins.

## Acceptance Criteria

- Local player basic attack presentation starts when the client sends a legal basic monster
  `action_intent`.
- Recovery UI still starts immediately from the client-side basic attack cooldown estimate.
- When the matching authoritative local-player combat result arrives, damage/miss text and target
  reactions still play, but the local player swing and attack audio are not replayed a second time.
- Remote/player/monster attacks still play from authoritative events as before.
- v295 input buffering and v296 attack-move scenarios continue to pass.

## Scope And Likely Files

- Client presentation helper: a focused helper such as
  `client/scripts/combat_local_attack_presentation.gd`.
- Client input/event presentation: `client/scripts/main.gd`, kept within the file-size ratchet.
- Client tests: focused helper coverage in `client/tests/test_sustained_input.gd` or a small
  presentation test.
- Bot proof: reuse `77_input_buffering` and `78_attack_move_sticky_targeting`; add focused unit
  coverage for duplicate suppression.
- Docs: v297 plan, as-built, lifecycle, and `PROGRESS.md`.

## Test And Bot Proof

Focused checks:

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=78_attack_move_sticky_targeting
```

## Asset And Plugin Decision

- Adopt: existing player animation clips, attack audio cue, attack recovery UI, and authoritative
  combat event presentation.
- Borrow: v295/v296 bot scenarios for regression proof.
- Reject: new assets/plugins, server/protocol changes, and new impact VFX.

## ADR Alignment

- ADR-0001 D2/D3: server remains authoritative for simulation outcomes.
- ADR-0007: animation stays client-only and event-driven; this slice only prevents duplicate local
  presentation for a locally started attack.

## Open Questions And Risks

- No blocking questions. Use a minimal target-matching presentation helper; if the server rejects an
  intent, the client may still have played a short local swing, which is acceptable presentation
  prediction for this slice.
- `main.gd` is at its grandfathered allowance, so the implementation must offset any integration
  lines by simplifying touched presentation paths or moving state into helpers.
