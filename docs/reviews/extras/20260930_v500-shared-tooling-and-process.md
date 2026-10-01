# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v500**

**Date:** 2026-09-30
**Scope:** shared/protocol, shared/rules, shared/golden, tools/, Makefile, scripts/, docs/specs,
docs/plans, docs/adr, SDD process — v487-v500 since the v486 review (14 slices)
**Baseline:** main @ 33c4dc35 (clean worktree)
**Stats:** protocol schemas 36 files v0-v8 (unchanged count); rules 22 instances + 22 schemas;
presentation catalogs 25 + 25 schemas (8 touched this window: `dungeon_kit_presentation`,
`equipment_display`, `item_presentations`, `render_presentation`, `town_presentation`,
`vfx_presentation`, `combat_feel_presentation`, `fog_presentation`); goldens 36 + 36 schemas (one
value edit — `shop_appraisals.json` — no new case files); `tools/bot/run.py` 4,540 lines;
`tools/validate_shared.py` 3,188 lines; maintainability ratchet 35 grandfathered files / 66,055
lines (`make maintainability`, 2026-09-30).
**Overview:** [`../20260930_v500-overview.md`](../20260930_v500-overview.md)

---

## Summary

This window (v487-v500, 14 slices in one day via the new `$autoloop` coordinated-batch process) was
almost entirely client-presentation work — kit ground loot, stairs, coins/potions, a town look pass,
combat VFX, town floor/ground detail, dungeon room dressing, frame-pacing and lighting fixes, live
targeting corrections, and town terrain/nature landmarks — plus one protocol/HTTP slice (v488 live
stash sync) and the scheduled `$refactor` that opened the window (`5aeb7f01`..`21c631a4` then
`86188689`..`6aba0b84`..`6c25333c`..`b4a5493b`). Every one of the 14 slices in this window has a
spec, a plan, and an as-built with no number collisions — the single worst v486 process finding is
fully resolved for this batch. The v494-v500 slices ran through the new autoloop batch workflow
(`.claude/skills/autoloop/`); it integrated cleanly with one combined `make ci` pass (11m13s per
PROGRESS.md, confirmed by `b0355228`) and a follow-up doc-only closeout commit.

The scheduled `$refactor` landed real fixes against the v486 findings: three server-authoritative
goldens now have Go consumers (`6c25333c`), the determinism lint now covers all of `game/`
(`5aeb7f01`), client gates now fail on GDScript runtime errors (`1958456b`), the client-chosen-tick
hole is bounded (`80465ab6`), a self-referencing `state_delta.v8` schema bug was fixed
(`b4a5493b`), and stale doc findings were repaired (`d3b5c234`). What did **not** move: `envelope.v8`
is still stale (30 of 43 real intents — narrower than v486's 24-of-45 gap, but still wrong and still
unvalidated), the protocol-versioning invariant in `CLAUDE.md` still claims a bump policy that is not
practiced, `execute_step` and `cross_checks()` are untouched monoliths, the ratchet's "touch-to-shrink"
rule is still being ignored (21 of 35 grandfathered files are now over baseline, up from 20), the
`origin` remote still carries a plaintext token, and the installed Godot (4.7.2) still drifts from the
unenforced `.godot-version` pin (4.6.3). This review is itself 0 slices late per PROGRESS.md's own
tracking (it explicitly flagged the review as "due now at v500").

---

## 1. Architecture

**[Strength — new]** The `$autoloop` coordinated-batch workflow
(`.claude/skills/autoloop/references/batch-workflow.md`) is a well-specified process artifact. It
names a single coordinator role, forbids slice sessions from running `make ci`/`/finish`/push,
requires a handoff manifest (base commit, full file list, focused-check outcomes, conflict risk),
and gives the coordinator an explicit ledger (`proposed → dispatched → ready → integrated → verified
→ closed`) plus an explicit prohibition on claiming completion while any slice is blocked or
missing from `main`. It is consistent with how v494-v500 actually ran: PROGRESS.md records "isolated
worktrees and integrated together; the final combined `make ci` passed in 11m13s **after an
integration fix**" — i.e. the workflow's "diagnose, fix the integration or slice defect, and rerun"
step was exercised for real, not just specified. One gap: the workflow document has no guidance on
what happens when two slice worktrees touch the *same* shared JSON catalog in incompatible ways
(this batch touched `town_presentation.v0.json` from at least v491, v493, v494, v500 — see §3); it
relies entirely on the coordinator's manual three-way merge judgment in step 2 of "Coordinator loop",
with no reconciliation checklist specific to shared schema files.

**[Med]** `envelope.v8.schema.json`'s intent enum is still wrong, though narrower than at v486. It
lists 30 `*_intent` values; `server/internal/inputdecode` decodes 43 distinct intent strings — 13
are still missing (v486 found 21 of 45 missing). `messages.v8.schema.json` carries the fuller,
current list (48 entries including non-intent event types). `tools/validate_shared.py:309-337` still
only asserts the v0-v8 envelope files exist; nothing validates the v8 enum against
`inputdecode`'s real set, and `docs/CODEMAP.md:24,27` still cites `envelope.v8.schema.json` as a
protocol-contract file alongside `messages.v8`. Recommendation 8 from v486 ("update `envelope.v8` to
the decoded intent list, or delete it and point CODEMAP at `messages.v8`") is unresolved.

**[Med]** The versioning-by-filename pattern continues. `state_delta.v8.schema.json` was edited again
in this window (`b4a5493b`, fixing a `$defs.equipped` self-reference — a real bug, correctly fixed —
and widening `equipped` to an explicit ten-slot object) with no version bump and no `protocol_version`
negotiation anywhere in the stack (still nothing in `server/internal`, `client/scripts`,
`tools/bot` — confirmed by the same grep that found nothing at v486). `CLAUDE.md`'s Key Invariants
section still states "Changes to `shared/protocol/` require a schema version bump" — this is the
same untrue claim flagged at v486 (recommendation 8c), still not rewritten to match the
additive-in-place practice that is actually used.

**[Strength]** Dual-consumer golden discipline held for the one golden value actually touched.
`shop_appraisals.json`'s `summary_lines` entry changed ("Kind: consumable, Restores 3 HP" →
"Level 1, Restores 3 HP"); `tools/validate_shared.py:2268-2341` and
`server/internal/game/golden_shop_appraisals_test.go:61-90` (new in this window, from the `$refactor`
commit `6c25333c`) both read the same file and neither needed updating beyond the data change — no
drift. No new rules/golden instance files were added in v487-v500; the eight presentation-catalog
JSON/schema pairs that grew (`dungeon_kit_presentation`, `equipment_display`, `item_presentations`,
`render_presentation`, `town_presentation`, `vfx_presentation`, `combat_feel_presentation`,
`fog_presentation`) are all client-presentation data per ADR-0001 D2, correctly kept out of
`shared/rules`.

## 2. Technical

**[Resolved since v486]** The three server-authoritative goldens with no Go consumer are fixed.
`6c25333c` ("test: add Go consumers for three server-owned goldens") added
`server/internal/game/golden_use_consumable_test.go`, `golden_shop_appraisals_test.go`, and extended
`golden_skill_progression_test.go` to load `use_consumable.json`, `shop_appraisals.json`, and
`skill_points_and_magic_bolt.json` through `loadGolden`. This closes v486's High finding #7/recommendation
#2 for structure; the leveled-potion coverage gap (v460 #2 / v486 §2 High: "still level-1 red_potion
only, zero `item_level` fields") was **not** addressed — `use_consumable.json` still has no
multi-level case, so the new Go test proves the *existing* (level-1-only) contract, not the richer
one both reviews called for.

**[Resolved since v486]** The determinism lint now covers all of `game/`, not just `sim.go`/
`handlers.go`. `5aeb7f01` ("widen determinism lint to all of game/ with a per-file baseline") plus
`18bb0697` (tie-break skill aim targets and unique test-chest IDs by sorted order) close the
`skill_aim_nav.go:21` order-dependent tie-break that v486 flagged as a live bug the old lint couldn't
see.

**[Resolved since v486]** Client gates no longer pass silently on a GDScript runtime error.
`1958456b` ("fail client gates on GDScript runtime/parse errors") and `86d3b834` ("retire the dead
eye-view weapon assertion; fix scenario 105") close both the "blind gate" and the
`eye_view_weapon_presentation` dead-assertion findings from v486's Top-10 #4.

**[Resolved since v486]** The unbounded client-chosen tick hole (v486's #1 blocker) is fixed:
`80465ab6` ("bound client-chosen input ticks to the current tick plus a small lead").

**[New, Low]** `b3d32d23`..`545bb59e` (v489 kit stairs, v490 kit coins/potions) and the v494-v500
batch are all additive presentation data; none of them touches `server/internal/game` formula code,
so this window adds no new golden-coverage debt beyond the pre-existing leveled-potion gap above.

**[Not addressed]** `run.py execute_step` and `validate_shared.cross_checks()` are untouched by this
window's code changes: `execute_step` is still a single branch chain (currently 4,540-line file,
~14 lines over its 4,526 baseline — see §3) and `cross_checks()` is still one function at
`validate_shared.py:203` (file now 3,188 lines, 6 over its 3,182 baseline). Agent rule 6 ("run.py
split freeze") is still being read as "never touch" rather than as the registry refactor the rule
actually describes.

## 3. Maintainability

**[Strength]** The ratchet's raw total is still trending down: 35 grandfathered files / 66,055 lines
now vs 35/66,266 reported at v486 (and 36/67,491 at v460) — a real, if small (-211 line), reduction
confirmed by `make maintainability` run directly on this checkout, not just inherited baseline
arithmetic.

**[Med — regressed slightly]** Touch-to-shrink (CLAUDE.md Maintainability Ratchet rule 4) is still
not being followed, and the count of files over their own baseline went up, not down: **21 of 35**
grandfathered files are above baseline now (v486: 20/35), for roughly 259 lines of unclaimed slack
(v486: 263 — essentially flat). The largest overages:
- `client/scripts/blacksmith_panel.gd` +22 (771 → 793)
- `client/scripts/market_panel.gd` +22 (1,059 → 1,081)
- `tools/bot/test_protocol.py` +21 (1,434 → 1,455)
- `client/scripts/fog_of_war_overlay.gd` +20 (611 → 631)
- `client/scripts/bot_assertion_handlers.gd` +19 (655 → 674)
- `tools/bot/run.py` +14 (4,526 → 4,540)
- `tools/validate_shared.py` +6 (3,182 → 3,188)

None of these 21 files is over its +25 hard ceiling yet, so the ratchet gate (`make maintainability`)
still passes CI, but the slack is being spent broadly across the codebase rather than being paid down
per rule 4's "a slice that edits a grandfathered file should leave it at or below its current
baseline."

**[New finding]** Shared presentation catalogs are a live multi-slice conflict surface for the batch
process. `town_presentation.v0.json`/`.schema.json` grew the most of any shared file in this window
(+373/+541 lines respectively) and was touched by v491 (town look pass), v493 (town floor/ground
detail), v494 (dungeon room dressing touches shared presentation infra), and v500 (town terrain and
nature landmarks) — four different slice sessions in the same batch editing the same catalog. The
batch integrated cleanly (one fix needed per PROGRESS.md), but the autoloop workflow doc (§1) has no
explicit reconciliation procedure for this specific, now-demonstrated, shared-JSON contention pattern
beyond generic "resolve overlapping files against the current integrated state."

**[Not addressed]** `.godot-version` is still unenforced. The pin is `4.6.3-stable`
(`.godot-version`), `client/project.godot:15` still declares `"4.6"`, and the installed host Godot is
now confirmed **4.7.2.stable.official** — a full minor-version drift, same gap v486 flagged. The pin
is still only echoed in error-message strings in `scripts/{bot_client,bot_visual,client_smoke,play,
play_remote}.sh`; nothing diffs the running `godot --version` against the file and fails.

**[Not addressed, Security]** `git remote -v` on this checkout still shows a GitHub personal access
token embedded in plaintext in the `origin` URL (same owner-blocked/accepted-risk item PROGRESS.md
already records). Confirmed present, not re-quoted here. v486's recommendation (`git remote set-url`
to a bare HTTPS URL plus a credential helper, and rotate the token) is still open and is explicitly an
owner action, not something this review attempts to fix.

**[Not addressed]** There is still no `.github/` directory and no remote CI run on `main` — "green"
remains self-reported (`make ci` run by the coordinator session), same as v486.

## 4. Documentation

**[Strength]** The scheduled `$refactor` (commit `d3b5c234`, "repair stale v486 review doc findings")
closed several v486 Documentation findings directly: struck the "Protocol schema drift (v485)" gap
from PROGRESS.md, removed the obsolete "Adventurers 2.0" note, fixed the ADR-0018 rig-item status,
corrected the `kaykit-asset-inventory.md` mesh-name prefixes (`Knight_Helmet`, `Barbarian_BearHat`,
`Mage_Hat`), and dropped the incorrect "(optional)" label on CI step 11 in `scripts/ci.sh`. The
CLAUDE.md text supplied to this review already reads "CI step 5/11" for the determinism lint,
confirming that specific inaccuracy (flagged at v486) is fixed project-wide, not just in the
refactor commit's diff.

**[Strength]** `PROGRESS.md` is both within its line budget (243/250, `check-progress-dashboard.sh`)
and honest about review cadence for the first time in two reviews: "Last engineering review: v486 …
Next engineering review: Due now at v500" — the self-reported overdue state that caused most of
v486's process score drop is explicitly tracked and current, not silently stale.

**[Strength]** Every one of the 14 slices in this window (v487-v500) has a spec, a plan, and an
as-built, with lifecycle rows in `docs/progress/slice-lifecycle.md` (confirmed v494, v499, v500
rows read "Complete (combined CI gate)" with spec/plan/as-built links). No slice-number collisions
in this window (each of v487-v500 maps to exactly one spec file). This directly resolves v486's
worst SDD-process finding (3 collisions, 4 orphaned slices, 10/26 missing plans) for the new batch,
though it does not retroactively fix the v461/v464/v458/v468 orphans v486 already named — those are
unchanged and still open.

**[Not addressed]** `envelope.v8` is still cited in `docs/CODEMAP.md:24,27` as a current protocol
contract file alongside `messages.v8`, despite being stale (§1). CODEMAP was otherwise kept current
for this window's new files (per its own validator, `validate_codemap.py`, which still passes).

**[Low, unchanged]** `docs/progress/scenario-catalog.md` is still 171 lines against 269+ scenario
ids from the v486 count; none of the new v487-v500 slices appear to have added scenario-catalog rows
(file line count unchanged at 171 since v486's own count of the gap).

## Prior review follow-up (v486 → v500)

| # | v486 recommendation | Status | Evidence |
|---|---|---|---|
| 1 | Bound the client-chosen input tick | **Resolved** | `80465ab6` |
| 2 | Remove credential from `origin`, rotate token | **Still open (owner action)** | `git remote -v` still shows a plaintext token |
| 3 | Get `make ci-full` green / root-cause 9+2 failures | **Not verified this window** | No `make ci-full` run found in this window's commit history; PROGRESS only records `make ci` (fast pack) results |
| 4 | Harden client gate (fail on runtime errors, register dead tests) | **Resolved** | `1958456b`, `86d3b834` |
| 5 | Widen determinism lint to all of `game/` | **Resolved** | `5aeb7f01`, `18bb0697` |
| 6 | Record gameplay-debug mode on session; surface dropped connect/disconnect errors | **Resolved** | `eb0f9928` (log/meter dropped writes); session shows additional related fixes in the refactor run (not independently re-verified against `sim.go:520` in this pass) |
| 7 | Add Go consumers for the 3 server-owned goldens | **Resolved (structurally)**, leveled-potion case still missing | `6c25333c`; `use_consumable.json` is still level-1 only |
| 8 | ADR for replay contract; rewrite versioning invariant; fix/retire `envelope.v8` | **Still open** | No new ADR found; CLAUDE.md invariant unchanged; `envelope.v8` still 13 intents short |
| 9 | Fix inverted wall-occlusion throttle; perf fixes | **Resolved (throttle)** | `fb92ac47` ("make the wall-occlusion rebuild throttle actually throttle") |
| 10 | One real extraction per hotspot (`main.gd`, `run.py`, `validate_shared.py`, `sim.go`) + gofmt gate + race detector | **Partially resolved** | `86188689` (gofmt gate), `3a542f3d` (race detector on `internal/realtime`); `run.py`/`validate_shared.py` monoliths untouched |
| — | Slice-number collisions / missing plans / missing as-builts | **Resolved for this window** | 14/14 slices in v487-v500 have spec+plan+as-built, no collisions; v458/v461/v464/v468 orphans from before v486 remain unfixed |
| — | `.godot-version` enforcement | **Still open** | Installed Godot is 4.7.2; pin is 4.6.3; still only echoed, not enforced |
| — | Remote/CI on main | **Still open** | No `.github/`; CI remains self-reported |

## Top 5 refactors

1. **[High · Contract]** Finish closing the protocol-contract loop that v486 already called out and
   this window left half-done: delete `envelope.v8.schema.json` or regenerate its enum from
   `messages.v8` (13 of 43 real intents are still missing — `shared/protocol/envelope.v8.schema.json`
   `properties/type/enum`, cross-checked against `server/internal/inputdecode`), fix
   `docs/CODEMAP.md:24,27` to stop citing it as current, and rewrite the `CLAUDE.md` Key Invariants
   line ("Changes to `shared/protocol/` require a schema version bump") to the additive-in-place
   policy actually practiced (most recently exercised again in `b4a5493b`'s in-place
   `state_delta.v8` edit).

2. **[Med · Test]** Add the leveled-potion golden cases that two reviews in a row have now flagged as
   open (v460 #2, v486 §2, this review's §2): extend `shared/golden/use_consumable.json` with
   `item_level` cases (5, 10, capped heal, rejuvenation) and update the now-existing Go consumer
   (`server/internal/game/golden_use_consumable_test.go`) plus the GDScript consumer
   (`client/tests/test_golden.gd:150`) together. The Go-consumer gap this blocked is now closed, so
   this is a pure data/test-case addition with no structural work needed first.

3. **[Med · Maint]** Do the `run.py`/`validate_shared.py` extraction the last two reviews both
   recommended and neither has started: a step-registry for `execute_step` (`tools/bot/run.py`,
   currently 4,540 lines, still one `if action == "…"` chain) keyed the way `handlers.go` keys
   `inputHandlers`, and named per-domain functions pulled out of `cross_checks()`
   (`tools/validate_shared.py:203`, 3,188 lines). Both files are still only incrementally growing
   (+14, +6 over baseline this window) rather than shrinking, which is the exact failure mode rule 4
   of the Maintainability Ratchet exists to catch.

4. **[Med · Process]** Pay down ratchet slack broadly rather than letting it accumulate: 21 of 35
   grandfathered files are now over their own baseline (up from 20/35 at v486), for ~259 unclaimed
   lines. Pick the worst three outside the two Python monoliths already covered above —
   `client/scripts/blacksmith_panel.gd` (+22), `client/scripts/market_panel.gd` (+22), and
   `tools/bot/test_protocol.py` (+21) — and do a real extraction or lower the baseline with a
   documented reason, per rule 6, the next time any of them is touched.

5. **[Low · Process]** Give the `$autoloop` batch workflow an explicit shared-JSON reconciliation
   step. This window proved the risk is real: `shared/assets/town_presentation.v0.json` was edited by
   four different slice sessions (v491, v493, v494, v500) in the same batch and needed one
   post-integration fix to go green. Add a short checklist to
   `.claude/skills/autoloop/references/batch-workflow.md`'s "Coordinator loop" step 2 — e.g. "list
   every shared `shared/**/*.v0.json` file touched by more than one slice in this batch and diff each
   one explicitly before the final `make ci`" — rather than relying on the generic "resolve
   overlapping files" instruction that is already there.

*Evidence: read-only inspection on `main` @ `33c4dc35` (clean worktree). Commands used: `git log`/
`show`/`diff --stat` (`2cf0dcd1..HEAD`), `git remote -v`, `wc -l`, `rg`/`grep`, `make maintainability`,
`godot --version`, and small scratchpad Python scripts for envelope/messages intent-enum diffs and
ratchet-baseline-vs-actual comparison. `make ci`/`make ci-full` were not run by this review; CI
results cited above (11m13s combined pass, "after an integration fix") are taken from PROGRESS.md and
the as-built notes, not independently re-run.*
