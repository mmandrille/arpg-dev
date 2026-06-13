# v138 Spec: CODEMAP & Reduction Ratchet

Status: Complete
Date: 2026-06-13
Codename: `codemap-and-reduction-ratchet`

## Purpose

Cut the per-feature context an AI agent must load, and convert the maintainability ratchet from a
*containment* device (freeze the monoliths) into a *reduction* device (shrink them as a side-effect
of normal work). This is the foundation slice of a 5-proposal maintainability/cohesion program
sourced from the v130 engineering review (see **Downstream Roadmap**).

Two deliverables in one cohesive slice — both are docs/tooling, neither changes game behavior,
protocol, persistence, or replay:

1. **`docs/CODEMAP.md` — a domain → files index.** Today an agent has no map of which files a
   feature spans, so it greps broadly and reads 7–9k-line coordinators (`sim.go` 7801, `main.gd`
   6745, `game_test.go` 9116, `run.py` 5189, `repos.go` 3052) to find ~200 relevant lines. A
   committed, validated index that answers "to work on domain X, load these files" cuts that cost in
   one read and makes the splits in the downstream roadmap discoverable.

2. **Reduction ratchet.** Three concrete upgrades to `scripts/check-file-size-ratchet.sh` +
   policy + CI wiring:
   - **Lower-bound ratchet (the reduction engine):** the script already fails when a grandfathered
     file exceeds `baseline + 25`. Add the symmetric rule — fail when a grandfathered file has shrunk
     so its committed baseline now sits more than the growth allowance *above* the actual line count,
     instructing the author to lower (or drop, if ≤600) the baseline. Once a file is cut, its baseline
     locks at the new low-water mark and can never grow back. Containment → reduction.
   - **Trend metric:** `make maintainability` prints a one-line summary — number of grandfathered
     files still over target and total grandfathered line count — so each milestone review can show
     the curve bending down rather than only "still grandfathered."
   - **CI enforcement:** `make ci` must actually run the ratchet. It does **not** today — `ci:`
     (`make/ci.mk:7`) only invokes `scripts/ci.sh`, whose 9 steps never call the ratchet, so
     `make maintainability` is advisory and relies on agent discipline. This is the single highest-value
     change and the root cause of the "file-size drift already present at the v130 boundary" the review
     noted.
   - **Documented policy:** a *touch-to-shrink* rule (a slice that edits a grandfathered file must
     leave it ≤ its baseline — growth allowance applies only to untouched incidental drift) and a
     *new-domain* rule (a genuinely new gameplay domain starts in its own file, never inside a
     coordinator) added to `CLAUDE.md` and the plan checklist.

## Non-goals

- **No source-coordinator splitting in this slice.** `sim.go`, `repos.go`, `main.gd`, `run.py`,
  `game_test.go`, and `validate_shared.py` are not carved here — that is the downstream roadmap
  (#2/#4). This slice builds the index and the enforcement engine that make those splits cheap,
  discoverable, and self-sustaining.
- No market expiration behavior change (#5 is downstream).
- No gameplay, protocol, persistence, replay, golden, or client UI change.
- No git-history-based "which files changed this slice" enforcement if it proves brittle on the
  work-on-`main` flow — touch-to-shrink lands as documented policy + plan checklist, with the
  lower-bound ratchet providing the mechanical guarantee (see Open Questions).
- CODEMAP is an index, not a tutorial: it lists files per domain, it does not re-document how each
  feature works (that stays in specs/as-builts/ADRs).

## Acceptance Criteria

### CODEMAP (#1)

- `docs/CODEMAP.md` exists with a short "how to use" header ("to work on domain X, load these
  files; consult `PROGRESS.md` for status and `docs/specs|plans|as-built/` for history") and a
  domain → files table.
- The table covers the major domains and, for each, lists concrete existing paths across the
  relevant layers (server / client / shared rules / bot / tests / migrations). At minimum:
  market, unique-items/effects, classes (paladin/rogue/ranger), skills, stash, shop/vendor,
  dungeon-gen, combat & damage-types, elite-aura, corpse/permadeath, town-services, session &
  realtime, replay, persistence/store, protocol, bot/scenarios, assets, i18n/text.
- Every file path referenced in `docs/CODEMAP.md` exists on disk, enforced by an automated check
  (`tools/validate_codemap.py`) wired into `make validate-shared` (or `make maintainability`) so the
  map cannot rot silently. A stale map is worse than none.
- `AGENTS.md` and `CLAUDE.md` point agents at `docs/CODEMAP.md` as the canonical "what files do I
  load for domain X" index, alongside the existing `PROGRESS.md`-first rule.

### Reduction Ratchet (#3)

- `scripts/check-file-size-ratchet.sh` fails with a clear, actionable message when a grandfathered
  baseline entry sits more than the growth allowance above the file's actual line count, telling the
  author to lower the baseline (or drop the entry when the file is ≤ the 600 target).
- `.maintainability/file-size-baseline.tsv` is refreshed to current exact line counts as part of
  this slice so the lower-bound check passes on landing.
- `make maintainability` prints a trend summary line (grandfathered file count + total grandfathered
  lines).
- `make ci` runs the ratchet and fails CI when it fails (via `ci:` depending on `maintainability`,
  or an explicit ratchet step in `scripts/ci.sh`).
- `CLAUDE.md` "Maintainability Ratchet" section is rewritten for reduction: lower-bound ratchet,
  touch-to-shrink rule, new-domain rule, and the trend expectation. The plan checklist item that
  already asks about over-limit files is extended to ask "did any touched grandfathered file stay at
  or below its baseline?"
- All existing tracked files pass under the new rules; `make maintainability` and `make ci` are
  green.

## Likely Files

- `docs/CODEMAP.md` (new)
- `tools/validate_codemap.py` (new) + `tools/test_validate_codemap.py` (new)
- `scripts/check-file-size-ratchet.sh`
- `make/ci.mk` and/or `scripts/ci.sh`
- `make/shared.mk` or `make/tools.mk` (wire `validate_codemap` into `validate-shared`)
- `.maintainability/file-size-baseline.tsv` (refresh to current counts)
- `CLAUDE.md`, `AGENTS.md`
- `PROGRESS.md`
- `docs/as-built/v138_codemap-and-reduction-ratchet.md`

## Test And Bot Proof

- `.venv/bin/python -m pytest tools -q` (includes new `test_validate_codemap.py`: asserts every
  CODEMAP path exists, and asserts the validator fails on a fabricated missing path).
- `make validate-shared` (CODEMAP path validation runs and passes).
- `make maintainability` (lower-bound ratchet + trend line; passes on refreshed baseline).
- A focused proof that the ratchet *fails* correctly: run the script against a temp baseline whose
  entry is far above a temp file's count and assert non-zero exit (pytest harness in `tools/`, or a
  documented manual check in the plan if a temp-fixture harness is disproportionate).
- `make ci` (now exercises the ratchet end-to-end).

No new bot scenario is required: there is no runtime gameplay, protocol, or client behavior change.
Existing CI bot/replay coverage remains the regression proof that nothing moved.

## Downstream Roadmap (documented here, executed in later slices)

This spec is the foundation of a program; the remaining proposals are separate specs (suggested
v139+) so each stays a small SDD slice. They are sequenced so #1/#3 make the rest cheap and
self-sustaining.

- **#2 — Server coordinator splits (package-internal, zero behavior/replay risk).**
  - `repos.go` → `market_repo.go`: move the contiguous market block `ListActiveMarketListings` …
    `scanMarketOfferItemRows` (`server/internal/store/repos.go:1709`–`2590`, ~900 lines) beside the
    existing `market_purchase.go`. Aligns the store layer with the already domain-split HTTP layer.
  - `sim.go` seams: extract `sim_load.go` (persistence-bridge `Load*`/`*ForPlayer` block,
    `sim.go:772`–`1078`) and `sim_players.go` (player lifecycle/spawn, `sim.go:1086`–`1364`),
    shrinking the file an agent reads to understand the deterministic tick loop.
  - Mirror test splits: start new market/loot/combat proofs in focused `*_test.go` files instead of
    growing `game_test.go` (9116) and `store_test.go`.
- **#4 — Client `main.gd` decomposition.** Continue the v127 `TownServiceBridge` pattern: peel
  ground-loot presentation (rarity tinting lives at `main.gd:5255`), world-entity lifecycle,
  combat-cue presentation, and panel-lifecycle management into focused
  `class_name … extends RefCounted` helpers, each proven by a headless test like
  `client/tests/test_town_service_bridge.gd`.
- **#5 — Market expiration freshness contract (the v130 correctness item).** Today expiry
  side-effects (offer refunds, item-to-stash restoration, `listing_expired` audit row) run only
  inside `ExpireMarketListings` (`repos.go:2049`), which is invoked only from
  `ListActiveMarketListings` (`repos.go:1710`); `GetMarketSummary` (`repos.go:2140`) never calls it
  and offer/detail queries merely filter with `expires_at > now()`. A listing that expires while the
  board is never opened leaves the bidder's escrow unrefunded and the seller's item unrestored
  indefinitely. Funnel every market read entrypoint through a single freshness call and/or run the
  sweep on a deterministic periodic server tick; document the contract ("expiry side-effects are
  guaranteed to have run before any market read returns"). Lands naturally inside the `market_repo.go`
  boundary from #2 — which is why #2a is sequenced first.

## Open Questions And Risks

- **Touch-to-shrink enforcement mechanism.** Airtight "which grandfathered files did this slice
  edit?" needs a stable base ref, which is fuzzy on the work-directly-on-`main` flow. Recommendation:
  ship touch-to-shrink as documented policy + plan-checklist gate, and rely on the lower-bound
  ratchet for mechanical enforcement (you literally cannot leave a file above its low-water mark +
  allowance). `/plan` decides whether to also attempt a best-effort `git diff <base>` check or keep
  it policy-only.
- **CODEMAP domain granularity.** The minimum domain list above is a floor; `/plan` finalizes the
  exact rows and whether to split "combat" vs "damage-types" etc. Risk: too fine = maintenance tax;
  too coarse = still loads too much. Aim for domains an agent would name when asked "what am I working
  on?"
- **Roadmap numbering.** This spec assumes #2/#4/#5 become specs v139+. Confirm whether to pre-create
  those spec stubs now or let `/next` schedule them after v138 ships.
- **Lower-bound ratchet false positives.** Refreshing the baseline to exact current counts on landing
  avoids day-one failures; thereafter normal feature deletions that drop a file >25 lines below
  baseline will (intentionally) require committing the lower baseline. This is the desired friction,
  not a bug — but document it so agents expect it.
