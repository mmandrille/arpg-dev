# v217 Spec: Paladin Charge Channeling Protocol

Status: Draft
Date: 2026-06-16

## Goal

Make Paladin Charge a held/channeling skill instead of a one-shot dash. While the player holds the skill input, the server moves the paladin in the current pointer direction, spends mana over time, and applies the existing Charge line impact to monsters crossed by the channel path. Releasing the input, running out of mana, or being blocked ends the channel.

## Player Experience

- Right-click Charge starts charging in the pointer direction.
- Moving the pointer while still holding right click updates the charge direction, so the paladin can trace a curved/circular path through enemies.
- Releasing right click stops the charge.
- The `paladin_class_foundation` bot scenario demonstrates one longer held Charge path that turns through multiple enemies, not two separate Charge casts.
- Visual verification command: `make bot-visual scenario=paladin_class_foundation`.

## Rules

- Channel tuning stays data-driven in `shared/rules/skills.v0.json`.
- Charge remains cooldown-free. Channeling skills do not start cooldowns; mana consumption is the limiter.
- Charge retains server-owned collision, movement, damage, stun, and push resolution.
- Mana drain is owned by server ticks and is emitted through authoritative entity updates.
- A monster should not be repeatedly damaged/stunned/pushed by every tick of the same Charge channel.

## Protocol

Add `channel_skill_intent` with payload:

- `skill_id`: active skill id.
- `phase`: `start`, `update`, or `stop`.
- `direction`: required for `start` and `update`; omitted for `stop`.

The server emits additive events:

- `skill_channel_started`
- `skill_channel_updated`
- `skill_channel_ended`

`skill_channel_ended.reason` should explain normal release or forced termination such as `insufficient_mana` or `blocked`.

## Acceptance

- Shared schema validation accepts `channel_skill_intent` and the channel events.
- Server tests prove Charge starts, moves over multiple ticks, drains data-driven mana, turns direction via update, impacts monsters on the path, and does not create cooldowns.
- Input decode tests cover `start`, `update`, and `stop`.
- Client tests cover channel payloads and right-click channel send helpers.
- Bot tooling can run a channel path action.
- `make bot scenario=paladin_class_foundation` uses one longer curved Charge channel and observes channel start/end plus monster displacement/stun.

## Non-Goals

- Full bespoke Charge animation art.
- Multi-player simultaneous channels.
- Backward compatibility for one-shot Paladin Charge behavior.
