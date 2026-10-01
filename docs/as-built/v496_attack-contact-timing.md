# v496 As-Built — Attack contact timing

- **Date:** 2026-09-30
- **Spec:** [`v496_spec-attack-contact-timing.md`](../specs/v496_spec-attack-contact-timing.md) · **Plan:** [`v496_2026-09-30-attack-contact-timing.md`](../plans/v496_2026-09-30-attack-contact-timing.md)
- **Scope:** client-side attack presentation and bot proof; no server damage, attack-speed, cooldown, or protocol change.

## Measured defects and fixes

The v493 baseline's lethal melee route sent one local click but started two lunges: the client consumed `monster_damaged` as the local swing's result, then treated the following `monster_killed` event as a second attack. `CombatLocalAttackPresentation` now leaves kill to death feedback only. The live 103 route reaches melee range and asserts exactly one lunge before floor retarget. Its focused sustained-input test covers the damage-plus-kill pair.

Matched 1280×720 Balanced [before](assets/v496/melee-before-balanced-matched.png) and [after](assets/v496/melee-after-balanced-matched.png) play-camera frames were captured at the same route checkpoint. The control run, with only the lethal-event guard removed, failed the one-lunge assertion as expected; the candidate passed. A still frame cannot show the duplicate motion on its own, so the route assertion supplies that behavioral proof.

A blocked `monster_damaged` event displayed `BLOCK` but incorrectly requested hit audio, target reaction, and hit particles. `GameplayFeedbackPresentation.has_hit_contact` now requires a confirmed hit for those effects. The live 109 route matches the block result and asserts zero added target impact feedback while retaining `BLOCK` text. Rapid buffered input is bounded and a floor retarget releases the pending combat command. The pre-fix reproduction and focused route details are in the [baseline note](../researchs/v496_attack-contact-baseline.md).

No new clip-contact offset was justified. The representative sword and bow hits both start locally in the input frame, and their authoritative results arrive at the existing clip contact position in the integrated v495+ client. Tuning remains in the existing shared presentation catalogs.

## Integrated contact timeline

The opt-in `v496_attack_contact_timeline` client-bot scenario equips a sword and then a bow against a durable live target in `combat_stat_lab`, seed `contacttimeline496`. It waits for the actual equipped item, records input/swing/result/feedback and sampled clip frames, and then retargets to the floor. Godot 4.7.2, Forward+ Metal, windowed 1280×720 and 60 Hz vsync were used in both tiers. The numbers below are one complete confirmed hit per weapon/tier; they are alignment observations, not latency percentiles.

| Tier and weapon | Input→local swing | Input→authoritative hit | Clip position at result | Result→reaction request |
|---|---:|---:|---:|---:|
| Balanced sword | 1.59 ms | 107.87 ms | 100.0 ms | 8.79 ms |
| Performance sword | 2.05 ms | 106.05 ms | 100.0 ms | 8.63 ms |
| Balanced bow | 0.06 ms | 408.40 ms | 399.9 ms | 0.73 ms |
| Performance bow | 0.06 ms | 408.71 ms | 399.8 ms | 0.69 ms |

The trace's sampled frame gaps had median/p95 of 16.60/19.41 ms in Balanced (108 samples) and 16.71/19.45 ms in Performance (109 samples). Hit text and target-side feedback occur only after the result row. The `audio_attack` call occurs with the local swing; `audio_damage` occurs after a confirmed result. These are scheduling records, not a listening test. The target-window [normal-speed Balanced recording](assets/v496/normal-speed-balanced.mov) and the [melee](assets/v496/normal-speed-melee.png) and [bow](assets/v496/normal-speed-bow.png) stills show the actual rendered sequence. The separate real-renderer [melee](assets/v496/melee-balanced.png) and [block](assets/v496/block-balanced.png) captures have more detail; Performance [melee](assets/v496/melee-performance.png) and [block](assets/v496/block-performance.png) captures preserve the same readable outcome cue.

The first sword attack returns to idle before the bow swap. The bow retarget overlaps an incoming player hit, so the initial trace labeled the hit clip as a return to locomotion. The trace now waits for an actual idle/walk/run clip; the focused trace test covers the corrected classification. That interruption does not constitute a second local bow swing. A normal-speed pre-fix video for the unchanged durable-target timing path was not captured; the before/after behavioral proof is the lethal lunge and block-event route, plus their focused tests.

## Verification and limits

`test_sustained_input`, `test_attack_contact_trace`, the 103/104/109 client scenarios, and windowed timeline runs in Balanced and Performance pass. The 109 scenario is the direct visual command for block feedback: `make bot-visual scenario=109_attack_block_contact`; the sword/bow timeline is `ARPG_ATTACK_TRACE=1 BOT_CLIENT_RENDER_QUALITY=balanced HEADLESS=0 make bot-visual scenario=v496_attack_contact_timeline` (replace tier with `performance`). The trace is opt-in and capped; normal gameplay has no recorder overlay. The integrated single `make ci` remains the final shared gate.

The window recording contains only the Godot window and no game audio. It proves normal-speed visible ordering in this lab viewpoint, while exact physical weapon/target contact and sound quality still benefit from human play review at wider camera/weapon combinations. No claim is made for all basic-attack families from this two-weapon sample.
