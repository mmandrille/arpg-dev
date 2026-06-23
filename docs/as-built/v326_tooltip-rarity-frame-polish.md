# v326 As Built - Tooltip Rarity Frame Polish

Date: 2026-06-23

## What Shipped

- Item tooltips use rarity-colored borders from `ClientConstants.LOOT_LABEL_RARITY_COLORS`.
- Magic/rare/unique items get thicker border width for quicker loot scanning.

## Proof

```bash
godot --headless --path client --script res://tests/test_look_and_feel_polish.gd
```
