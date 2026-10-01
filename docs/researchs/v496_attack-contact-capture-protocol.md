# v496 normal-speed contact capture protocol

Status: instrumented and unit-checked in the v496 checkout. The real renderer capture and timing table remain open until the coordinating task releases the host.

## Fixture and recorder

`tools/bot/scenarios/client/v496_attack_contact_timeline.json` uses the existing combat-stat lab: sword and bow loot, a durable armored target, and one floor retarget. It waits for the actual item definition in the equipped slot before changing weapons. Use the existing scenario 109 separately for an authoritative block; 103 and 104 remain regressions.

Set `ARPG_ATTACK_TRACE=1` only for the measurement run. The client bot then shows a small frame/time/clip overlay and prints bounded `[attack-trace]` JSON rows on exit. It is otherwise inactive. The trace records input, intent dispatch, local swing, result processing, text request, reaction/VFX call, audio call, and return to locomotion. `frame_sample` rows include the current animation position, visible combat text, lunge offset, and inter-frame gap. A text or VFX call is not proof the effect was visible; mark the actual contact and outcome frame in the recording.

Record the Godot window with a normal-speed OS screen recorder. Do not use Godot `--write-movie`: the pre-v495 attempt produced 2 seconds of footage in 71 seconds, so its timing was not representative. On macOS, start the built-in screen recording (`Shift-Command-5`) before launching the bot, and stop it after the bot passes. Keep the overlay in frame. The video supplies the real frame sequence; the trace aligns its visible contact frame to the monotonic client timestamps. In-process PNG reads during the swing would perturb the timing being measured.

From the integrated checkout, for each quality tier:

```sh
mkdir -p .artifacts/v496
ARPG_ATTACK_TRACE=1 BOT_CLIENT_RENDER_QUALITY=balanced HEADLESS=0 BOT_STEP_DELAY=0.25 make bot-visual scenario=v496_attack_contact_timeline | tee .artifacts/v496/contact-balanced.log
ARPG_ATTACK_TRACE=1 BOT_CLIENT_RENDER_QUALITY=performance HEADLESS=0 BOT_STEP_DELAY=0.25 make bot-visual scenario=v496_attack_contact_timeline | tee .artifacts/v496/contact-performance.log
```

The bot launcher normally deletes its temporary Godot log. With tracing enabled, it echoes just the `[attack-trace]` rows before that deletion so `tee` retains them. The `fixture` row identifies the world, seed, quality, renderer, and the `trace_start` row identifies viewport and Godot version. Run scenario 109 with the same trace flag for a block result and `BLOCK` text.

## Measurement table to complete after capture

For each family and tier, select one complete attack attempt by target ID and frame number. Record input→first attack-clip frame, input→visually identified contact frame, input→`result_received`, result→first visible combat text/reaction frame, and input→`locomotion_return`. Note sample count, viewport, renderer, clip position at visible contact, median/p95 frame gap, whether audio was heard, and whether the recording stayed at normal play speed. Treat `feedback_text_requested` and audio rows as scheduling evidence; use the recorded frames and listening check for visible/audible proof. Keep a miss or block row from scenario 109 with zero confirmed-hit reaction.

If the trace shows a reproducible visual offset, add a schema-backed contact offset in `shared/assets/` and verify both weapon families. If timing already aligns, retain the current data and document it as the control. Final v496 acceptance and performance claims still require the coordinating v495-based capture.
