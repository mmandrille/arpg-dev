# v385 As-Built — gameplay feel polish

Autoloop batch: level-up celebration, door scale, consumable feedback, loot-label pickup, and
monster impact artifact toggle.

## What shipped

### Level-up (item 11)
- **No XP floaters** — unchanged; no new floating XP text added.
- **Server:** `level_up.restore_hp_and_mana` in `character_progression.v0.json` (default `true`).
  `restorePlayerResourcesOnLevelUp` in `level_up_resources.go` sets HP/mana to max and emits
  `player_healed` / `player_mana_restored` with `reason: "level_up"`.
- **Client:** radial glow burst on `character_leveled` via `level_up_burst.gd`; level-up restores
  skip consumable floaters/effects.

### Doors (item 14)
- `door_presentation.gd`: wall-scale height (~2.85), framed wooden design, open burst on unlock.

### Consumables (item 15)
- `consumable_use_effect.gd`: brief flash + torus ring on potion/mana use (not on level-up restore).

### Loot labels (item 16)
- `loot_label_hover.gd`: screen-space pick when Alt reveals labels; click label text to highlight
  and pick up.

### Enemy artifacts off (item 17)
- `combat_feel_presentation.v0.json` → `enemy_impact_feedback.enabled: false` disables monster
  impact sparks, outcome punch, and model lean/flash/death reactions. Player reactions preserved.

### Extractions (maintainability)
- `gameplay_feedback_presentation.gd`, `interactable_state_presentation.gd`, `level_up_resources.go`,
  `level_up_test.go`; baseline updates in `.maintainability/file-size-baseline.tsv`.

## Verification

```bash
make maintainability
make validate-shared
make client-unit
cd server && go test ./internal/game/... -run TestExperienceGainAndLevelUpFromMonsterKill
make ci
```

Manual: `make play` — Alt+click loot labels, open doors, level-up burst after kill.

CI pack: demoted flaky `market_purchase_ui` to extended; tightened `client_combat_feedback`
(crit/block/toggle contract, high-VIT debug setup); updated `model_reaction_polish` for disabled
monster artifacts.
