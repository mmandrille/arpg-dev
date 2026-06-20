# v297 As-Built: Client-Side Windup / Recovery Presentation

Date: 2026-06-19
Spec: [`docs/specs/v297_spec-client-side-windup-recovery-presentation.md`](../specs/v297_spec-client-side-windup-recovery-presentation.md)
Plan: [`docs/plans/v297_2026-06-19-client-side-windup-recovery-presentation.md`](../plans/v297_2026-06-19-client-side-windup-recovery-presentation.md)

## What Shipped

- Added `CombatLocalAttackPresentation`, a client-only tracker for the locally started basic attack
  target.
- Started local player basic attack audio/animation presentation when the client sends an existing
  monster `action_intent` from normal attack dispatch, sticky attack follow-up, autoplay, or direct
  bot monster click.
- Consumed matching local-player `monster_damaged`, `monster_killed`, `attack_missed`, and
  `attack_blocked` authoritative results so the local swing/audio is not replayed a second time.
- Preserved authoritative result presentation: damage/miss text, target hit/death reactions, damage
  audio, kill audio, server combat legality, cooldowns, damage, death, loot, and protocol remain
  unchanged.
- Kept remote players and monsters event-driven, because only local-player source events with the
  tracked target are consumed.
- Moved the playback helper into the focused presentation tracker so `main.gd` stayed under its
  grandfathered file-size allowance.

## Proof

Focused verification:

```bash
godot --headless --path client --script res://tests/test_sustained_input.gd
godot --headless --path client --script res://tests/test_client_bot.gd
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=77_input_buffering HEADLESS=1 ./scripts/bot_client_local.sh
GODOT=godot ARPG_ADDR=:18082 BASE_URL=http://localhost:18082 SCENARIO=78_attack_move_sticky_targeting HEADLESS=1 ./scripts/bot_client_local.sh
make maintainability
```

Result: green on 2026-06-19. The local bot script printed the same post-pass
`cleanup_account.py` missing-`httpx` warning in this environment, but both scenarios returned
success.

## Manual Visual Command

```bash
make bot-visual scenario=78_attack_move_sticky_targeting
```

## Deferred

- Hit stop / impact flash remains the next selected Movement / Combat Fluidity slice.
- Movement acceleration smoothing, command retarget grace, and melee lunge remain later selected
  slices in this feature queue.
- Server-authoritative hit timing, damage prediction, animation-speed scaling, and production
  attack assets remain out of scope.
