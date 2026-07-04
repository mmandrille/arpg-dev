# v440 As Built — Attack Animations Melee 2H

Date: 2026-07-04

## Shipped

- `shared/assets/attack_presentation.v0.json` + `AttackPresentationLoader`.
- `attack_2h` animation clip; combat presentation reads equipped weapon metadata.
- Off-hand mount suppressed when main-hand weapon is two-handed.

## Verification

```bash
make client-unit
```
