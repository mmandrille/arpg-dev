# v441 As Built — Attack Animations Ranged

Date: 2026-07-04

## Shipped

- `attack_ranged` (bow hold + string pull) and `attack_staff` (two-hand cast) clips.
- `AttackPresentationLoader.clip_for_weapon()` routes ranged weapons to new clips.

## Verification

```bash
make client-unit
```

Visual: `make bot-visual scenario=<ranger-scenario>` or showme ranger/sorcerer attack pose.
