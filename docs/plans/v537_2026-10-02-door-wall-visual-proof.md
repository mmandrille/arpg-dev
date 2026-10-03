# v537 Plan — Door, wall and entrance visual proof

- **Spec:** [`v537_spec-door-wall-visual-proof.md`](../specs/v537_spec-door-wall-visual-proof.md) · **Baseline:** `dab60eb5` + v532/v533/v535/v536. Evidence-only; runs last.
- **Review gate:** pass. No code, protocol, rules or golden change; scenarios must satisfy the movement contract (no incidental navigation) and the `ci_tier`/pack validators. Asset/plugin decision: none (uses the in-repo `capture_frame` step).

## Tasks

- [x] **1. Scenarios.** Add client scenarios for the entrance/wall frames at default and zoomed-in views (lab world with a generated room-threshold door) and the deep props floor (`dungeon_dressing_deep_lab`). Register them in the movement audit.
- [x] **2. Capture.** Run each windowed (`HEADLESS=0`); keep a warning-clean pass; record Godot warnings and A/B any leak warning against the baseline worktree.
- [x] **3. Review frames.** View each PNG, keep at least four distinct ones, note what they show.
- [x] **4. Go test runtime.** Time the `game` package isolated; decide and apply the test-timeout handling (CI `go test` timeout and/or slow-test trims); record it.
- [x] **5. As-built, lifecycle, PROGRESS (set complete pending combined CI).**

## Final gate

Focused only; the combined `make ci` runs after this slice.
