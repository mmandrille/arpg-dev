# v535 Plan — Wider room entrances

- **Spec:** [`v535_spec-wider-entrances.md`](../specs/v535_spec-wider-entrances.md) · **Baseline:** `dab60eb5` + v532/v533 in tree. Standalone; v536 and v537 follow.
- **Review gate:** pass. Server-side data and a load-time validator, no protocol change. Determinism: no new randomness or map ranges. Replay/goldens move deliberately. Asset decision: none (no client asset). Ratchet: `rules.go` is grandfathered; edit is a field plus a 3-line call, new logic lives in the new `dungeon_passage_clearance.go`.
- **Correction recorded:** the spec's draft numbers (corridors 2.5/3.0) do not fit deep 10×8-room floors (see Task 3 and the as-built), so the widths are 2.25/2.4 and door gaps 2.4. Extra corrections: an approach-search fix (cell center) and re-pinning scenario 124 to seed `_03`.

## Tasks

- [x] **1. Floor first.** Add `PassageClearanceRules`, validator, schema entry, rules field; test (`dungeon_passage_clearance_test.go`) for the floor and for each rejected case.
- [x] **2. Retune data.** `passage_clearance.min_player_diameters` 2.5 (floor 2.25); `corridor_widths` [2.25, 2.4]; legacy `room_layout.corridor_width` 2.25; `doors.gap_width` 2.5.
- [x] **3. Generation budget.** Bisect which width starves deep floors (corridors, not door gaps; cliff between 2.0 and 2.25 on the depth-6 profile with 10×8 rooms). Resolve in data: remove that profile's `room_size_min` override so it uses the default 12×10. Confirm `TestRoomCorridorLayout_SeedSweepAlwaysGenerates` passes.
- [x] **4. Fallout.** Run `go test ./internal/game -count=1`; fix failing tests by deriving expectations from rules (Test Locking Policy); `make regen-golden` only for genuinely layout-owned goldens; re-run `make validate-shared`, `make lint-determinism`, `go vet`.
- [x] **5. Proof.** Run `124_room_doors` and one threshold-crossing bot scenario; one real-camera capture of a generated entrance before/after (same seed).
- [x] **6. As-built, CODEMAP row, lifecycle, PROGRESS.**

## Final gate

Focused only; combined `make ci` after v532–v537.
