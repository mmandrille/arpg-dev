# v198 Plan: Mercenary Foundation

## Spec

`docs/specs/v198_spec-mercenary-foundation.md`

## Tasks

1. Add `mercenary_guard` to monster rules with companion-safe stats, no drops, and no XP.
2. Add matching monster visual metadata and text catalog entries.
3. Add `mercenary_foundation_lab` with one owned mercenary companion and one hostile target.
4. Add a focused Go test for mercenary companion identity and authored stats.
5. Add protocol bot scenario `86_mercenary_foundation.json` for follow/assist combat proof.
6. Update as-built docs and `PROGRESS.md`, then run verification and commit.

## Verification

- `make maintainability`
- `make validate-shared`
- `cd server && go test ./internal/game -run 'Mercenary|Companion' -count=1`
- `make bot scenario=86_mercenary_foundation.json`
- `make ci`

## Notes

- This is an archetype and simulation foundation only. Hiring, persistence, mercenary equipment, and roster UI are deliberately deferred.
- No Godot plugin adoption is needed because the existing biped dummy presentation is reused.
