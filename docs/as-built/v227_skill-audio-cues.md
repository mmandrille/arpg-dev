# v227 As-Built - Skill Audio Cues

Date: 2026-06-16

## What shipped

- Added generated skill cue families for projectile, buff, protection, movement, heal, revival, and
  generic fallback casts.
- Recorded `last_skill_id` in `ClientAudioController` debug state so muted/headless runs still prove
  the selected skill cue.
- Extended reusable client-bot `assert_audio_state` checks to cover `last_cue` and `last_skill_id`.
- Updated the skill-points client scenario to prove Rage produces the `skill_buff` cue after the
  authoritative `skill_cast` event.

## Proof

```bash
make client-unit
make bot-client scenario=19_skill_points_and_magic_bolt.json HEADLESS=1
make maintainability
```

All focused checks passed on 2026-06-16 during `$autoloop`. The enclosing batch-level `make ci`
passed after the selected v226-v232 feature queue completed.

Manual visual/audio proof, if desired:

```bash
make bot-visual scenario=19_skill_points_and_magic_bolt.json
```

## Scope limits

- No server, protocol, shared skill-rule, mana, cooldown, damage, animation, or progression changes
  shipped.
- No production audio assets, external audio packs/plugins, positional audio, per-rank sound
  variants, or class-authored sound libraries shipped.
