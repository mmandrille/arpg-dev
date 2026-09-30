# arpg-dev — Shared contracts, Python tooling & SDD process review at slice **v486**

**Date:** 2026-09-29
**Scope:** `shared/` (protocol, rules, presentation assets, goldens), `tools/` (bot, validators, asset
tooling, showme), `Makefile` / `make/*.mk` / `scripts/*.sh` orchestration, maintainability ratchet, and
the SDD trail (specs/plans/as-built/PROGRESS/ADRs/CODEMAP) for v461–v486 against the v460 baseline (`e75e64d0`).
**Baseline:** `main` @ `2cf0dcd1` (*test: v484: remote player class test uses the data-driven class
fallback*), review branch `claude/review-v486`. **The worktree was not clean:** at session start, 18
`client/assets/**/*.glb.import` files had uncommitted edits. They are Godot 4.7 import keys
(`mesh_library/use_node_names_as_mesh_names`, `gltf/texture_map_mode`), which is the environment-drift
finding in §3.
**Stats:** Python **23,801** lines (`git ls-files '*.py'`); scenarios **269** (150 protocol + 119 client,
of which 3 are `benchmark` tier); CI pack **35** (22 + 13, unchanged since v460); protocol **v8** (36
versioned schema files v0–v8 + 37 examples); rules **22** instances + 22 schemas; presentation catalogs
**25** + 25 schemas; goldens **36** + 36 schemas; specs 459 / plans 432 / as-built 480; `run.py`
**4,540** lines; `validate_shared.py` **3,188** lines; grandfathered files **35 / 66,266 lines**
(my recount at `2cf0dcd1`; the lead reported 66,387 from `make maintainability` at v484; v460 had 36 / 67,491).
**Overview:** [`../20260929_v486-overview.md`](../20260929_v486-overview.md)

---

## Summary

The biggest contract win since v460 is v486. For the first time, live server→client payloads are
validated against the v8 schemas (`tools/bot/payload_schema.py`, strict in `scripts/ci.sh:390`). The
same slice also shows how much the "schema is the contract" claim had been untrue: the first surveys
found **44 drifts**. Nearly all were fixed by bending the schema to the server (as-built
`v486_live-payload-schema-gate.md` §Drift fixed); only two server-side fixes landed. Several contract
surfaces are still stale or uncovered:
- `envelope.v8` is out of date.
- Client→server intents are never checked on live traffic.
- Three goldens for server-authoritative formulas have no Go consumer.

Python tooling structure improved at the edges: validators are extracted with injected callbacks, the
GLB tooling is tested, and the `helpers=globals()` debt is at zero. The two monoliths are untouched:
`execute_step` is 1,554 lines and `cross_checks` is 2,971 lines.

The SDD process is the weakest area:
- The periodic review is 16 slices overdue.
- 4 shipped slices have no as-built or lifecycle row.
- 10 of the 26 slices had no plan.
- 3 slice numbers were reused.
- There is no remote CI.
- Godot on the host (4.7.2) drifts from the pin.
- The `origin` URL embeds a credential.

---

## 1. Architecture

**[Strength]** Contract layering is sound. Every rules, presentation and golden instance has a sibling
schema: `schema_for()` maps `*.v0.json` to `*.v0.schema.json` (`tools/validate_shared.py:95-118`), and
the counts are 22/22, 25/25 and 36/36. Presentation data stays out of `shared/rules` (for example
`shared/assets/armor_look.v0.json` and `kit_hero_presentation.v0.json`), so ADR-0001 D2 authority is
not diluted.

**[Strength]** The v486 gate is well built:
- It is a leaf module that validators are compiled into once per message type
  (`payload_schema.py:1-22`).
- It dispatches on the `op`/`event_type` discriminator. A test proves the result equals the plain
  validator, and the as-built measures +2.4% CI time.
- A negative control proved it fails CI.
- It is hooked at the single ingest path (`state_ingest.py`), not in `run.py`.

**[High]** In practice the protocol is versioned only by filename. `state_delta.v8.schema.json` was
created on 2026-06-10 and has since been edited in **56 commits**. v485 added an entity prop with an
explicit non-goal of "A protocol version bump" (`docs/specs/v485_spec-remote-player-class.md:45`).
There is also no runtime version negotiation: a grep for `protocol_version` in `server/internal`,
`client/scripts` and `tools/bot` finds nothing. That contradicts the CLAUDE.md invariant "Changes to
`shared/protocol/` require a schema version bump" (`CLAUDE.md` Key Invariants). Either the policy or the
practice is wrong. The honest policy is "additive-optional in place; bump only on breaking change",
written down in one place.

**[Med]** The server, not the schema, is the de facto contract. In v486, 42 of 44 drifts were resolved
by schema catch-up, including dropping bounds (`party.maxItems 2`, the `hex_seed` pattern) and turning
`rolled_stats` into an open map. That is the right call for a gate's first day, but it means the schema
has been rubber-stamping server output. The v486 follow-up (a static Go `json:` tag ↔ schema
cross-check, PROGRESS.md:136) is the structural fix, and it is still open.

**[Med]** `envelope.v8.schema.json` is dead, and it is wrong:
- Its `type` enum lists 24 intents. The server decodes 45 distinct `*_intent` strings. 21 are missing,
  including `swap_weapon_set_intent`, `channel_skill_intent`, `companion_command_intent`, the
  `resource_bag_*` intents, `shop_reroll_intent` and `unique_chest_take_item_intent`.
- `messages.v8` carries its own, current copy of the enum; only debug intents are missing there.
- Nothing validates against `envelope.v8`. `validate_shared.py:332-337` only checks that the file exists.
- `docs/CODEMAP.md:24,27` still points agents at it as the protocol contract.

**[Med]** Client→server intents are unguarded on live traffic. The gate covers only `session_snapshot`
and `state_delta` (`payload_schema.py:40-43`). The intent direction is proven only by the 37
hand-written examples, and `server/internal/inputdecode` is the real contract. The Godot client bot and
the `check_persistence` reconnect snapshots are also unvalidated (PROGRESS.md:139). Scenario coverage
bounds the gate as well: events that no protocol scenario produces are unchecked, and
`schema_validation_suspended()` turns the gate off in the two densest combat windows
(`combat_soak_runtime.py:135,241`).

**[Low]** 32 legacy protocol schemas (v0–v7 × 4) are kept alive only by existence assertions
(`validate_shared.py:338-362`, "protocol vN schema set is present"). No code consumes them. Archive
them or document why they are retained.

## 2. Technical

**[High]** Three server-authoritative goldens have no Go consumer. The Go tests load 30 goldens through
`loadGolden` (`server/internal/game/game_test.go:124`), plus `upgrade_success_chance.json` and
`dungeon_obstacles.json` through direct paths. None of them reads:
- `use_consumable.json`, which is consumed only by `client/tests/test_golden.gd:150` and
  `validate_shared.py:2642`;
- `shop_appraisals.json`, which is consumed only by `validate_shared.py:2268`;
- `skill_points_and_magic_bolt.json`, which is consumed only by `test_golden_skill_progression.gd` and
  Python.

For potion heal, sell appraisals and skill-point gating, the authority never asserts the "cross-language
contract". The goldens only prove that the client and the validator agree with each other.

**[High]** The v460 #2 leveled-potion golden is still open. `use_consumable.json` is still "level-1
red_potion" only, with 2 cases and zero `item_level` fields. The v460 `3 × item_level` and rejuvenation
formulas have no golden in any language.

**[Strength]** The v460 #1 `status_effects` recommendation is closed.
`validate_shared.py:223-224` loads the catalog, `:3137` rejects unknown `*_status_id` refs, and
`750023a5` landed it.

**[Strength]** The concurrency fix is correct by construction:
- `scripts/server_helpers.sh` trusts only its own server's `server listening` line, with the pid
  matched and the bound port checked against `ADDR`/`BASE_URL`, before probing `/readyz` while that pid
  is alive.
- `scripts/test_db.sh` derives the per-checkout DB from `basename + cksum(path)` and handles the
  create race.

Neither script has a unit test. There is no `tools/**/test_*` for them.

**[Low]** `test_db.sh prune` drops every `arpg_test_*` database with an empty comment
(`test_db.sh:61-66`). A database created before the comment convention, or one whose `COMMENT` failed,
is dropped even while a run is using it.

**[Low]** The asset validators fail open. `validate_assets.py:309,322` wrap the dungeon-kit and
kit-monster checks in `if kit_path.is_file()`, so a deleted catalog passes silently. The armor-look
region suffixes (`_Body`, `_Helmet`, …) are never checked against real GLB mesh names. `glb_reader` can
already list them (for example `Knight_Helmet`, `Barbarian_BearHat`).

**[Low]** GLB I/O is triplicated:
- There are two parsers (`glb_reader._parse_glb_bytes`, `glb_mesh_io.parse_glb`) and three writers
  (`gen_glb.py`, `glb_mesh_io.write_glb`, `gltf_to_glb.write_glb`).
- `bake_material_palette.py:32` imports the private `_pad` from `gltf_to_glb`.
- `glb_mesh_io` is openly a "moved verbatim" rescue from the deleted `rig_hero_glbs.py`.

**[Low]** "Pinned tooling" is not pinned. `pyproject.toml` has lower bounds only
(`jsonschema>=4.23.0` and others) and there is no lockfile, while `make tools` is described as
"install pinned tooling".

## 3. Maintainability

**[Strength]** The ratchet trend is down, slowly:
- 36 → 35 files and 67,491 → ~66.3k lines since v460.
- `sim.go`'s baseline dropped 6,782 → 6,621.
- The extraction-coupling baseline is **empty** (`.maintainability/extraction-coupling-baseline.tsv`:
  "No grandfathered `helpers=globals()` sites").
- `validate_shared.py` shrank in both v483 (−5 net) and v484 (−3 net), because new checks went to
  focused modules (`validate_armor_look.py`, 60 lines, pure, with injected `slot_matches`).

**[Med]** The ratchet slack is being consumed rather than touch-to-shrink being followed:
- 20 of 35 grandfathered files sit above baseline, by 263 lines in total.
- Six are at +20 to +24: `dungeon_gen.go` +24, `blacksmith_panel.gd` +22, `market_panel.gd` +22,
  `test_protocol.py` +21, `fog_of_war_overlay.gd` +20.
- On tools: `run.py` +14, `validate_shared.py` +6.

CLAUDE.md Maintainability Ratchet rule 4 says an edited grandfathered file should end at or below
baseline. That is not being done.

**[High]** Both Python monoliths are untouched (v460 #3 and #4 still open):
- `execute_step` in `tools/bot/run.py` is a single **1,554-line** function with **88**
  `if action == "…"` branches (`run.py:438` through `:1956`). This is the Python twin of the
  `applyInput` switch that the Go side already replaced with a registry. `quest_steward_pick` is still
  inline (`run.py:1594`).
- `cross_checks()` is still **2,971** lines (`validate_shared.py:203`).

Agent rule 6 ("run.py split freeze") is being read as "never touch". The unblocked path is a step
registry keyed by action, like `handlers.go`, with typed `BotContext` injection (`bot_context.py`
already exists).

**[Low]** The run.py helper bridges are narrow but untyped. `_wait_runtime_helpers`,
`_coop_runtime_helpers` and `_combat_soak_helpers` (`run.py:2335,3592,4082`) are `dict[str, Any]`
string-keyed service bags. They are legal under Extraction independence, but they are stringly typed
where `BotContext` shows the typed pattern.

**[Med]** Environment drift is producing churn:
- The host runs Godot 4.7.2 against the `4.6.3-stable` pin (`.godot-version`), and
  `client/project.godot:15` declares `"4.6"`.
- `.godot-version` is only echoed in error messages (`scripts/*.sh`); it is never enforced.
- Result: 18 dirty `.import` files in this fresh review worktree.
- A stale `.import` commit or a headless-import mismatch will eventually land on main.

## 4. Documentation

**[Strength]**
- `docs/CODEMAP.md` is current. It lists `payload_schema.py`, `validate_armor_look.py`, the kit
  loaders and `server_helpers`, and `validate_codemap.py` enforces the paths.
- ADR-0018's rollout table (`0018-…md:336-341`) tracks P0–P4b to real as-builts.
- Visual as-builts commit their captures (for example `docs/as-built/assets/v483/gear-*.png`; 49 doc
  PNGs are tracked).

**[Med]** PROGRESS.md is sliding back into a changelog. The CI gate caps it at 250 lines, and it has 238
(`scripts/check-progress-dashboard.sh:7`). The gate is being dodged with width instead:
- line 26 (Latest completed slice) is **1,462 characters** and lists 18 slices;
- line 147 is 1,515 characters;
- ten lines exceed 600 characters;
- the file is 29.9 KB.

Stale entries:
- The "Protocol schema drift (v485)" gap (PROGRESS.md:111) was closed by v486. `combat_stats` is now in
  `state_delta.v8`.
- The "Before P3: owner downloads … Adventurers 2.0" item (PROGRESS.md:162) is obsolete, since P3a–P3c
  shipped.
- ADR-0018's Status line still says an open item "carries into P3".
- The three pre-existing extended failures and the `TestDungeonTeleportersReplayGolden` flake are still
  listed from v448, with no resolution evidence 38 slices later.

**[Med]** CLAUDE.md is inaccurate in two places:
- It says `make lint-determinism` is "CI step 3/9" (`CLAUDE.md:259`; also `PROGRESS.md:207`). It is
  **step 5/11** (`scripts/ci.sh:311`).
- The protocol version-bump invariant contradicts practice (§1).

Step 11 is labelled "(optional)" (`ci.sh:477`) but fails CI like any other step.

**[Low]** Research docs are drifting:
- `docs/researchs/kaykit-asset-inventory.md:55-60` lists unprefixed parts (`Helmet`, `HelmetVisor`,
  `BearHat`, `Hat`). The GLBs actually contain `Knight_Helmet`, `Knight_HelmetVisor`,
  `Barbarian_BearHat` and `Mage_Hat`.
- `armor_look` works only because it suffix-matches `_Helmet` and the other headgear names.

**[Low]** The inventories are sparse:
- 171 of 269 scenario ids never appear in `docs/progress/scenario-catalog.md`; no scenario added since
  v460 does.
- `slice-codename-index.md` skips v458–v464 and v466–v468.
- Five as-builts still say `Commit: pending` (v481, v482, v483, v484, v486).

## Prior review follow-up (v460)

| # | v460 recommendation | Status | Evidence |
|---|---|---|---|
| 1 | Load `status_effects` and validate unique-effect `*_status_id` | **Resolved** | `validate_shared.py:223-224,3137`; `750023a5` |
| 2 | Leveled-potion golden cases | **Still open**, and worse than stated | `use_consumable.json` is level-1 only, and Go never consumes it (§2) |
| 3 | Decompose `cross_checks()` | **Still open** | 2,971-line body at `validate_shared.py:203`. New checks now go to modules (partial mitigation) |
| 4 | Extract `quest_steward_pick` | **Still open** | `run.py:1594`, inside the 1,554-line `execute_step` |
| 5 | Close v458; refresh the scenario catalog | **Still open, and regressed** | v458 spec says "Implemented, awaiting `/finish`", with no as-built or lifecycle row. v468, v461-surface-kit and v464-surface-material-kit are now in the same state. The catalog lacks 171 ids |
| — | Plans skipped (v454–v456) | **Changed, and worse** | 10 of 26 slices have no plan: v470, v473, v474, v475, v477, v482, v483, v484, v485, v486. v485 was a protocol change |
| — | `helpers=globals()` debt | **Resolved** | Coupling baseline is empty |

**SDD trail detail (v461–v486):**
- Slice numbers were reused three times: two v461 specs (`dungeon-surface-kit`,
  `entity-locomotion-polish`), two v464 specs, and v476 in commit `cd800675` (CC0 beasts, later
  documented as v482) next to `b164365c` (recorded load shed).
- Two of the orphans shipped real code with no as-built or lifecycle row:
  - `49dedb6a` (v461 dungeon surface kit) changed `shared/rules/dungeon_generation.v0.json`,
    `worlds.v0.json` and Go `dungeon_obstacle_variety.go`. It is **not** spec-gate exempt.
  - `eade2594` (v464) added `surface_material_presentation.v0.json` and scenario 102. Both specs still
    say `Status: Draft`.
- The review cadence broke. The last review was v460; the next was due at ~v470 and is 16 slices late
  (PROGRESS.md:28-29).
- There is no `.github/` in the repo, so "CI gate" means "the agent ran `make ci` locally".
  - v469 and v470 shipped with `make ci` "owed" because of the Xcode license (slice-lifecycle rows
    36-37).
  - v472 shipped with "make unavailable on host".
- The `origin` URL embeds a credential (`https://<REDACTED>@github.com/mmandrille/arpg-dev.git`). It
  sits in plain text in `.git/config`, which every agent and tool session can read.

## Top 5 shared/tooling/process refactors

1. **[High · Security/Process]** Remove the credential from `origin`:
   `git remote set-url origin https://github.com/mmandrille/arpg-dev.git`, then use `gh auth` or the
   osxkeychain helper. Rotate the token. Then add a minimal remote CI, even `make ci` minus Godot on a
   runner, so "green" is not self-reported.
2. **[High · Correctness]** Make the authority consume its goldens. Add Go tests for
   `use_consumable.json` (extended with leveled cases: levels 5 and 10, capped heal, rejuvenation),
   `shop_appraisals.json` and `skill_points_and_magic_bolt.json`. Then add a `validate_shared` rule:
   every golden that names a server-owned formula must have a `loadGolden` consumer.
3. **[High · Contract]** Close the protocol contract loop:
   (a) add the static Go `json:` tag ↔ v8 schema cross-check (the v486 follow-up);
   (b) delete `envelope.v8`, or regenerate its enum from `messages.v8`, and fix CODEMAP;
   (c) validate outbound intents in the bot against `messages.v8` in strict mode;
   (d) rewrite the CLAUDE.md versioning invariant to match the additive-in-place practice.
4. **[Med · Maint]** Replace `execute_step`'s 88-branch chain with an action → handler registry. Handlers
   take a typed `BotContext`, and each move proves import independence. Start with `quest_steward_pick`
   (v460 #4). In parallel, split `cross_checks()` into named per-domain functions.
5. **[Med · SDD]** Pay down the SDD trail and make it enforceable:
   - Write as-builts and lifecycle rows for v458, v461 dungeon-surface-kit, v464 surface-material-kit
     and v468, and renumber or annotate the duplicates.
   - Add a `check-progress-dashboard` rule that every `docs/specs/vN_*` has an as-built or an explicit
     `Status: Abandoned`, and cap PROGRESS line width.
   - Either reinstate plans, or amend CLAUDE.md with an explicit "small-slice: plan optional"
     criterion. The silent skip is the problem.
   - Enforce `.godot-version` in `scripts/godot_ci_flags.sh`, or bump the pin to 4.7.x.

*Evidence: read-only inspection at `2cf0dcd1`. Commands used: `git log`/`show`/`diff --stat`
(`e75e64d0..HEAD`), `wc -l`, `grep`/`rg`, and small scratchpad Python scripts for golden-consumer
mapping, intent-enum diffs, ratchet headroom, scenario-catalog coverage and GLB mesh names. The v486
figures (44 drifts, +2.4% CI, negative control) come from `docs/as-built/v486_live-payload-schema-gate.md`
and were not re-run: `make ci-full` was in flight, owned by the lead. That run is also the first strict
schema-gate pass over the full 150-scenario protocol matrix, so any new drift it reports confirms §1.*
