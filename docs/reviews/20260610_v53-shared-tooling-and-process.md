# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v53**

**Date:** 2026-06-10
**Scope:** `shared/` (protocol, rules-as-data, golden), `tools/` (Python: bot, validator, assets, replay, play), `docs/` (PROGRESS, specs, plans, ADRs), root build (`Makefile`, `make/*.mk`, `scripts/`)
**Baseline:** `main` @ `be32b20`
**Stats:** `shared/` 20,472 LOC · 36 protocol schema files (v0–v8 × 4 families) · 35 golden fixtures · 53 specs + 53 plans · 9 ADRs · `tools/` 9,935 LOC · `pytest tools/` 59 passing
**Overview:** [`20260610_v53-overview.md`](20260610_v53-overview.md)

---

## Summary

This layer holds the project's two best assets — **rules-as-data with no logic leakage** and **cross-language golden fixtures** — and its most underrated risk: the two Python tools (`run.py` 4,824, `validate_shared.py` 2,779) are god-files, the protocol keeps all 9 versions alive with no deprecation path, and the documentation an agent reads first (`CLAUDE.md`) is 43 slices stale. The SDD process itself (53/53 spec↔plan↔slice) is exemplary and is the reason this project reached v53 coherently.

**Sub-scores:** Contracts-architecture 7.5 · Technical 6.5 · Maintainability 6.5 · Documentation 6.5 · SDD process 9.5.

---

## 1. Architecture (contracts + process)

- **Protocol version accumulation is unbounded. [High]** All 9 versions (v0–v8) of all 4 families (36 files) are kept in `shared/protocol/`. No `deprecated` marker, no negotiation layer, no pruning policy. `validate_shared.py:174–220` asserts v4–v8 are *present* (tested as alive); v0–v3 receive no validation (pure accumulation). At ~v8 after 53 slices, this trends to v15+ before 1.0 with no removal criteria.
- **Inter-version delta is trivial — confirming heavy duplication. [Med]** `envelope.v7→v8` = 2 lines (`$id`/title); `messages.v7→v8` = 1 substantive line; `session_snapshot.v7→v8` = 2 lines; `state_delta.v7→v8` = 25 lines (one `oneOf` branch). New versions copy ~everything verbatim. A `$ref` base + overlay strategy would cut ~70% of the schema surface.
- **Rules-as-data is clean. [Strength]** `character_progression.v0.json:41–52` uses pure parameterized formula descriptors (`{"type":"linear","base":...,"per_str":...}`). No scripts, no conditionals, no computed fields. ADR-0001/D6 honored.
- **Golden coverage lags delivery. [Med]** v50–v53 (stash, mystery seller, ranged AI, boss bar) ship **no** golden fixtures. The mystery seller adds a new `mystery_shop_offer` wire shape (`state_delta.v8`) with no golden. Coverage trails slices by ~4.
- **Build orchestration is coherent. [Strength]** `Makefile` → 9 `make/*.mk`; `scripts/ci.sh` runs 8 labeled phases under `set -euo pipefail` with pid-based server lifecycle and a quiet wrapper.

## 2. Technical

- **`validate_shared.py` is a god-script. [High]** 2,779 lines, ~11 top-level functions, a ~2,600-line `cross_checks()` fusing three concerns (schema meta-validation, data consistency, **formula simulation** — e.g. a `ShopRNG` class at `:1835`). Should split into `validate_protocol.py` / `validate_rules.py` / `validate_golden.py`.
- **`bot/run.py` is a 4,824-line monolith. [High]** 156 functions across HTTP steps (`:205`), WS slice (`:337`), assertions (`:2963`), main (`:4650`). The 4-section comment structure helps, but nothing is importable in parts — any reuse pulls the whole driver.
- **Examples validate only against v8. [Med]** `validate_shared.py:73–79` routes every `protocol/examples/*` to the v8 schema. An example authored for v6/v7 passes silently as long as it's v8-valid; no example-to-version provenance.
- **`health_regen` naming split. [High]** `health_regen_per_second` (in `character_progression.v0.json:51`, `validate_shared.py:389`, `character_progression` golden) vs `health_regen_per_10_seconds` (in `item_templates.v0.json`, `shops.v0.json`, `validate_shared.py:696,1986`). Two names/units for one stat across the contract boundary, **not** cross-checked. Verified present in 6+ files.
- **`Any` overuse. [Low]** ~151 `: Any` hits in `run.py`; wire dicts justify most, internal state could be tighter.
- **`pytest tools/` clean.** 59 tests, ~0.17s.

## 3. Maintainability

- **`PROGRESS.md` at 1,680 lines is near the sustainability edge. [Med]** Carries the 53-row slice table, a 50-entry "what each slice proved" log, the deferred-work registry, flow diagrams, and the agent checklist. Disciplined, but the "Open gaps" section sits behind ~1,400 lines of history with no archival/rotation. 53 more slices ≈ 3,000 lines.
- **SDD 1:1 discipline is real. [Strength]** 53 specs ↔ 53 plans, matching codenames (e.g. `v51_spec-mystery-seller-core.md` ↔ `v51_2026-06-09-mystery-seller-core.md`). This is the backbone of the whole effort.
- **Bot scenarios are sequential, not composable. [Med]** 38 protocol + 26 client scenarios as numbered JSON step scripts; no inheritance/fixture reuse, so late scenarios re-walk dungeon-entry setup from scratch.
- **Golden regeneration is fully manual. [Med]** No `make regen-golden`, no `--update` path in `validate_shared.py`. Hand-editing fixtures on a formula change is a correctness hazard for the project's primary cross-language guarantee.
- **Per-version cost compounds. [Med]** A new message field today = 4 new schema files + a ~50-line presence block in `validate_shared.py` + `schema_for()` update + examples. No base composition.

## 4. Documentation

- **`CLAUDE.md` is 43 slices stale. [High]** `CLAUDE.md:13` — *"Current snapshot (2026-06-05): slices through v10"* — actual is v53. The file instructs agents to read it first; the inline snapshot actively misleads (the `PROGRESS.md` pointer is correct).
- **ADR-0001 is excellent. [Strength]** ~400 lines, all 8 decisions with rationale, rejected alternatives, and honest consequences. A real design record.
- **ADR-0007 is minimal but accurate**, with an "As Built: v4" addendum — a healthy pattern.
- **ADR-0013 is orphaned. [Med]** Status "Proposed" with 10 open questions, but v51 already implemented the mystery seller on `main`. The actual decisions (bag-only, magic/rare floor, per-character stock) are not written back.
- **ADRs 0002–0005 missing. [Med]** ADR-0001 promised follow-up ADRs (wire protocol, shared-rules contract, determinism enforcement, netcode); none exist. Those decisions live only in specs/code.
- **No as-built architecture doc. [Med]** Only ADR-0001 (initial intent) + 53 slice specs. No current module map, no statement of the single live protocol version, no determinism-contract reference doc.
- **Spec quality is high** (purpose, non-goals, cross-refs consistently present).

---

## Top 5 improvements

1. **Fix `CLAUDE.md` staleness now** (`:13`): v10 → v53 or replace the inline snapshot with a pure pointer. *(trivial/high)*
2. **Unify `health_regen`** to one canonical name/unit across `character_progression`, `item_templates`, `shops`, golden, and `validate_shared.py`; add a cross-check. *(low/high)*
3. **Add a protocol deprecation policy** ("retired when no code references it and no live session uses it"), archive v0–v5 to `shared/protocol/archive/`, drop them from presence checks. Halves the active schema count. *(med/med)*
4. **Add `make regen-golden`** (drive the Go sim to emit canonical fixtures), backfill v50–v53. Closes the manual-edit hazard on the primary correctness guarantee. *(med/high)*
5. **Split `validate_shared.py`** into protocol/rules/golden modules (with `--write-golden` living in the golden module); likewise carve importable modules out of `run.py`. *(med/med)*

Plus: write `docs/ARCHITECTURE.md` (as-built), reconcile ADR-0013 to as-built, and add the missing foundational ADRs (or note in ADR-0001 that those decisions were enacted via specs).

---

*Evidence: reads of `validate_shared.py`, `bot/run.py`, `shared/protocol/*` (v7↔v8 diffs), `shared/rules/character_progression.v0.json`, `item_templates.v0.json`, `shops.v0.json`, `docs/adr/0001/0007/0013`, `docs/PROGRESS.md`, `CLAUDE.md`, `scripts/ci.sh`; verified `health_regen` split across 6+ files and `CLAUDE.md:13` v10 snapshot; `pytest tools/` 59 passing.*
