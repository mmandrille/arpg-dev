# v213 Spec: Bishop Account Revival

Date: 2026-06-16

## Goal

Make the town Bishop a stronger account recovery service:

- Respec is free.
- The Bishop panel has a separate button that revives every dead character on the current account.

## Player-Facing Requirements

- Opening the Bishop still restores the active hero's health and mana.
- The respec action no longer requires or deducts gold.
- The Bishop panel shows respec as a free action.
- A separate Bishop action sends `bishop_revive_all_intent`.
- Revive-all clears the durable dead/death-level state for all dead characters owned by the actor account.
- Revive-all must not revive characters owned by other accounts.

## Protocol Requirements

- Add `bishop_revive_all_intent` with `bishop_entity_id`.
- Add `bishop_revive_all` event with Bishop service metadata.
- Existing Bishop range, target, and live-player checks apply before the action is accepted.

## Adopt / Borrow / Reject

- Adopt: existing Bishop service/panel flow and durable character death model.
- Borrow: existing account-scoped store update patterns.
- Reject: new assets, new NPCs, or a separate HTTP endpoint; the feature belongs in the current Bishop service.

## Out of Scope

- Item corpse recovery changes.
- Per-character revive selection.
- Paid revive pricing.
- New visual effects or audio.
