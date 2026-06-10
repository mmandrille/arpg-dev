# v27 — Hold click controls

**Proves:** Diablo-style sustained left-click input can live entirely in the Godot client by repeating
existing intents at the current send cadence, without protocol or server changes.

- Hold LMB on a live monster locks a sticky target and repeats `action_intent` at `SEND_INTERVAL`
  until the monster dies, the player dies, LMB releases, or the target becomes invalid.
- Hold LMB on floor repeats `move_to_intent` toward the mouse ground point when cursor movement
  exceeds a 0.25 xz epsilon.
- Loot, doors, stairs, teleporters, and chest clicks stay one-shot; open chests are non-actionable
  and do not spam intents.
- Out-of-range hold-attack still uses v11 auto-approach; WASD cancel of auto-nav is unchanged.
- `SustainedClickInput` helper + `test_sustained_input.gd` cover hold start/stop/epsilon logic;
  bot hold+drag scenario remains deferred.

**Explicit non-goals:** no server swing cooldown, no hold-move walk animation, no controls remapping
UI, no new bot drag scenario.
