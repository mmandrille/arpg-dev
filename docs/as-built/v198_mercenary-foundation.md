# v198 As-built: Mercenary Foundation

Date: 2026-06-15
Status: Complete - `make ci` green

## What shipped

- Added `mercenary_guard` as a no-drop, no-XP companion-safe monster archetype.
- Added mercenary visual metadata using the existing biped dummy presentation.
- Added English and Spanish text catalog entries for the mercenary name.
- Added `mercenary_foundation_lab` with an owned mercenary companion and a hostile target.
- Added a focused server test proving the mercenary appears as an owned companion with authored HP
  and attack stats.
- Added protocol bot scenario `86_mercenary_foundation.json`, proving the mercenary follows and
  damages a target through existing companion AI.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'Mercenary|Companion' -count=1`
- `make bot scenario=86_mercenary_foundation.json`
- `make ci`

## Deferred

- Hiring, persistence, roster UI, equipment, commands, death recovery, and leveling remain future
  mercenary slices.
