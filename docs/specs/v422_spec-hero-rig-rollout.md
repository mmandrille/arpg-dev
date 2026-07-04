# v422 Spec: Hero Rig Rollout

Status: Approved
Date: 2026-07-03
Codename: hero-rig-rollout
Baseline: v421 gear-rig-corridor

## Purpose

Apply v421 mesh-height normalization and rig v2 to barbarian, rogue, sorcerer, and ranger so all
playable heroes share ~1.85m rig bounds and mount the new armor GLBs consistently.

## Non-goals

- No server/protocol changes.
- No new equipment GLB families (v421 assets reused).
- No monster/companion rig (v424).

## Acceptance criteria

- [ ] `HERO_TARGET_HEIGHTS` covers barbarian, rogue, sorcerer, ranger.
- [ ] Regenerated runtime hero GLBs pass `make validate-assets`.
- [ ] `test_animation.gd` passes walk/attack/off-hand bone checks for all five classes.
- [ ] Class presentation scales remain 1.0 (no 10x paladin-style hacks).

## Proof

```bash
make gen-assets
make validate-assets
godot --headless --path client --script res://tests/test_animation.gd
```
