# Spec: `data-driven-skill-catalog`

Status: Complete
Date: 2026-06-10
Branch: `main`
Slice: v59 - data-driven skill catalog
Baseline: v58 `boss-pattern-variety`
Related:

- [`../../PROGRESS.md`](../../PROGRESS.md)
- [`../adr/0001-technology-stack.md`](../adr/0001-technology-stack.md) - authoritative server, shared rules as data, bounded formula catalog
- [`../researchs/godot-plugins-and-shortcuts.md`](../researchs/godot-plugins-and-shortcuts.md) - client presentation shortcut checklist
- [`v44_spec-skill-points-and-magic-bolt.md`](v44_spec-skill-points-and-magic-bolt.md) - current skill point and Magic Bolt loop

## 1. Purpose

Magic Bolt exists as the first server-authoritative active skill, but the surrounding system still
treats it as a one-off: server validation has Magic Bolt-specific assumptions and the Godot skill
panel/skill bar hardcode one skill's identity, text, and placement. This slice turns Magic Bolt into
the first entry in a schema-backed skill catalog that is ready for future content additions.

The goal is not to add many new skills. The goal is to make the existing Magic Bolt loop prove the
right content workflow:

- Skill mechanics live in shared rules data.
- Skill tree placement and prerequisites are declarative.
- Presentation metadata lives separately from authoritative mechanics.
- Go and Godot use bounded helper/evaluator types, not free-form formulas.
- New skills that use existing capabilities can be added by data; new capabilities still require
  code, schema, validation, and tests.

Magic Bolt now requires **15 magic** before it can be learned and used. Existing local character data
may be discarded during implementation/testing; preserving stale characters that learned Magic Bolt
without the new requirement is explicitly not required.

## 2. Non-Goals

- No new active skills beyond Magic Bolt.
- No class selection, class starting stats, or class-locked character creation. `class: "mage"` is
  catalog metadata for now.
- No respec, skill-point refund, passive skills, or alternate branches.
- No free-form expression strings such as `"basic_attack * 2"`. Formulas must be selected from a
  bounded schema-backed evaluator catalog.
- No new damage capability such as AoE, piercing, homing, DOT, buffs/debuffs, summons, traps, auras,
  chained projectiles, or status effects.
- No final combat balance pass. Damage/cooldown numbers may change only as needed to keep the
  existing proof deterministic.
- No production skill VFX/audio.
- No external gameplay or skill-tree plugin adoption. The plan should record **reject** for plugins
  that want to own skill authority; this slice should extend the in-repo display-only client UI.
- No production migration for stale local character data; a local `make db-reset`/character wipe is
  acceptable before bot proof.

## 3. Acceptance Criteria

1. `shared/rules/skills.v0.json` uses a catalog shape for Magic Bolt that includes identity,
   class metadata, tree placement, kind, targeting, max rank, requirements, cost, damage, projectile,
   and cooldown.
2. `shared/rules/skills.v0.schema.json` rejects unknown authoritative gameplay fields and validates
   every supported helper/evaluator type.
3. Magic Bolt requirements declare `magic >= 15`. The server rejects learning and casting Magic Bolt
   while the actor does not satisfy the requirement, without mutating skill points, mana, cooldowns,
   or combat state.
4. The protocol bot proof levels a clean character far enough to allocate magic to 15, learns Magic
   Bolt, casts it, rejects cooldown recast, recovers, and proves reconnect/replay/fresh-session
   persistence.
5. Server skill validation is generic over all catalog skills that use supported capabilities; it no
   longer validates only `magic_bolt` with a special path.
6. Skill progression views remain deterministic and sorted by skill id.
7. The server uses bounded evaluators/helpers for Magic Bolt's requirements, mana cost, damage,
   projectile settings, and cooldown.
8. Adding an unsupported skill `kind`, requirement type, damage formula type, cooldown type, or
   projectile shape fails validation with a clear error.
9. A new `shared/assets/skill_presentations.v0.json` or equivalent presentation catalog supplies
   Magic Bolt's client-facing label/icon/color/tooltip fragments without becoming gameplay
   authority.
10. Godot loads shared skill metadata through a `class_name ... extends RefCounted` static loader
    pattern, matching `ItemRulesLoader`.
11. The Godot skills panel renders Magic Bolt name, tree position, rank, requirement status, spend
    affordance, and tooltip content from catalog/server state rather than hardcoded `MAGIC_BOLT_ID`
    branches.
12. The Godot skill bar resolves its slot label/icon/tooltip from skill presentation data and still
    disables/re-enables from server-owned cooldowns.
13. Existing skill protocol schemas remain valid unless planning finds a required additive schema
    bump; any bump must update examples and tests in the same slice.
14. `tools/validate_shared.py` cross-checks skill rules, skill presentations, golden fixtures, and
    known evaluator types.
15. Go tests, GDScript unit/golden tests, protocol bot, client bot, and `make ci` pass.

## 4. Data Shape Draft

The exact field names can be adjusted in the plan, but the intended shape is:

```json
{
  "version": 0,
  "skills": {
    "magic_bolt": {
      "name": "Magic Bolt",
      "class": "mage",
      "tree": {
        "tier": 1,
        "column": 1
      },
      "kind": "projectile_attack",
      "targeting": "direction_or_target",
      "max_rank": 5,
      "requirements": {
        "level": 1,
        "stats": {
          "magic": 15
        },
        "skills": []
      },
      "cost": {
        "mana": {
          "base": 3,
          "per_rank": 0
        }
      },
      "damage": {
        "type": "rank_linear_range",
        "min_base": 4,
        "max_base": 6,
        "min_per_rank": 1,
        "max_per_rank": 1
      },
      "projectile": {
        "range": 9.0,
        "speed": 10.0,
        "visual": "magic_bolt_projectile"
      },
      "cooldown": {
        "type": "attack_interval_multiplier",
        "multiplier": 2.0
      }
    }
  }
}
```

Presentation data should be separate, for example:

```json
{
  "version": 0,
  "skills": {
    "magic_bolt": {
      "icon": {
        "label": "M",
        "shape": "bolt",
        "color": "#62b7ff",
        "accent": "#e8f7ff"
      },
      "summary": "Projectile spell",
      "projectile_visual": "magic_bolt_projectile"
    }
  }
}
```

`row` from the initial idea is represented as `tree.tier`; `tree.column` keeps later branching
layout explicit. `class` is metadata in v59, not a character creation rule.

## 5. Server Behavior

### 5.1 Catalog Loading And Validation

The Go rules loader reads the expanded skill catalog into typed structs. Validation must be
catalog-wide:

- every skill id is non-empty and unique by JSON object key
- every skill has a supported `kind`
- every `tree.tier` and `tree.column` is positive
- every requirement key is supported
- every stat requirement uses an existing base stat and a positive threshold
- every skill prerequisite references an existing skill id and legal rank
- every cost, damage, projectile, and cooldown helper uses a supported `type`
- every presentation key references an existing skill id

### 5.2 Requirements

Magic Bolt requires `magic >= 15`.

The server must use one requirement helper for both progression spend and skill cast. This keeps
future respec/stat-loss cases safe even though v59 does not implement respec.

Learning rejection should use a stable reason such as `skill_requirements_not_met`. Casting rejection
should use the same or a closely related reason and must not spend mana or alter cooldowns.

### 5.3 Skill Evaluation

Magic Bolt continues to be a server-owned projectile skill. The implementation may preserve the
current rank-linear damage values with the new `rank_linear_range` evaluator. A later slice may add
`base_attack_multiplier` or another formula type, but v59 should not introduce an expression parser.

The existing `skill_cast`, `skill_cooldown_started`, `skill_cooldown_rejected`,
`monster_damaged`, and progression events remain the proof surface.

## 6. Client Behavior

The client remains a renderer/input layer. It may use shared rules and presentation metadata to show
requirements, labels, and cooldown text, but the server owns learning, mana spend, cooldowns, and
damage.

Expected UI behavior:

- Before `magic >= 15`, the skills panel shows Magic Bolt as not learnable and exposes the missing
  requirement in debug state or tooltip text for bot proof.
- After the player reaches `magic >= 15` and has a skill point, the spend button becomes enabled.
- After spending, the skill bar shows Magic Bolt using presentation metadata.
- Skill cooldown display continues to come from `skill_cooldowns`, with local interpolation only as
  presentation.

Because this slice changes client skill UI but does not add new art/plugin dependencies, the plan
should record the Godot shortcut decision as **reject external plugin; extend in-repo display-only
UI**.

## 7. Scope And Likely Files

```text
docs/specs/v59_spec-data-driven-skill-catalog.md - this spec
docs/plans/v59_2026-06-10-data-driven-skill-catalog.md - implementation plan
PROGRESS.md - lifecycle update at finish
docs/as-built/v59_data-driven-skill-catalog.md - as-built summary at finish

shared/rules/skills.v0.json - expanded Magic Bolt catalog shape and requirements
shared/rules/skills.v0.schema.json - catalog schema and evaluator validation
shared/assets/skill_presentations.v0.json - skill presentation metadata
shared/assets/skill_presentations.v0.schema.json - presentation schema
shared/golden/skill_points_and_magic_bolt.json - updated requirement/stat path and expected values
shared/golden/skill_points_and_magic_bolt.v0.schema.json - fixture schema updates if needed
shared/protocol/examples/session_snapshot.json - update if skill metadata examples change
shared/protocol/examples/state_delta.json - update if rejection/event examples change

server/internal/game/rules.go - skill structs, catalog validation, requirement/evaluator helpers
server/internal/game/handlers.go - generic skill requirement enforcement in spend/cast handlers
server/internal/game/sim.go - skill view/cooldown/damage helpers as needed
server/internal/game/game_test.go - requirement, validation, spend/cast, cooldown, golden tests
server/internal/replay/replay_test.go - replay proof if existing skill fixture changes

client/scripts/skill_rules_loader.gd - shared skill/presentation static loader
client/scripts/skills_panel.gd - data-driven skill row, requirement status, tooltip
client/scripts/skill_bar.gd - data-driven slot label/tooltip/icon state
client/scripts/main.gd - wire loader metadata into skill UI if needed
client/scripts/bot_scenario_runner.gd - assertions for skill requirement/presentation state
client/tests/test_skills_panel.gd - requirement/rendering state tests
client/tests/test_skill_bar.gd - presentation lookup/cooldown state tests
client/tests/test_golden.gd - updated skill golden parity if needed

tools/validate_shared.py - cross-check skills, presentations, evaluator types, golden drift
tools/bot/scenarios/32_skill_points_and_magic_bolt.json - protocol proof update
tools/bot/scenarios/client/19_skill_points_and_magic_bolt.json - client proof update
```

## 8. Test And Bot Proof

Required verification:

```bash
make validate-shared
cd server && go test ./internal/game/... -run Skill
make client-unit
make bot
SCENARIO=19_skill_points_and_magic_bolt.json ./scripts/bot_client_local.sh
make ci
```

The plan may run a local database reset before end-to-end proof because this slice intentionally
changes character requirements and the user approved deleting local characters for a clean baseline.
Use the existing project command, expected to be `make db-reset`, rather than hand-editing database
tables.

Protocol bot proof must include:

- clean character starts with `magic: 5`
- Magic Bolt cannot be learned before `magic >= 15`
- the character levels enough to allocate ten stat points into magic
- Magic Bolt can be learned once requirements and skill point are available
- Magic Bolt can be cast and starts the expected cooldown
- cooldown recast rejection does not spend mana or reset cooldown
- reconnect, replay, and fresh-session checks retain rank and magic stat

Client bot proof must include:

- skills panel shows Magic Bolt from catalog metadata
- spend button is disabled while magic requirement is unmet
- after magic allocation, spend button enables
- skill bar renders Magic Bolt from presentation metadata after learning
- cooldown disabled/recovery state still follows server `skill_cooldowns`

## 9. Open Questions And Risks

| # | Question / Risk | Default / Mitigation |
|---|-----------------|----------------------|
| R-1 | Requiring `magic >= 15` extends the v44 bot path because level 3 does not grant enough stat points. | Update protocol/client scenarios to level farther before learning Magic Bolt. Prefer semantic assertions over exact monster counts where possible. |
| R-2 | Existing local characters may have Magic Bolt without the new requirement. | No backward-compatibility work; local DB reset is allowed for this active-development slice. |
| R-3 | Data-driven skills can drift into an unvalidated mini scripting language. | Keep formulas as bounded evaluator types and reject unknown authoritative fields. |
| R-4 | Client UI may still hardcode one-skill assumptions. | It is acceptable to display one skill in v59, but the code should iterate catalog/progression rows rather than branch on `magic_bolt`. |
| R-5 | Protocol schema bump may not be needed. | Avoid a bump unless payload shape changes; if a bump is needed, update all examples/tests together. |

## 10. Planning Notes

- This slice should be implemented shared -> server -> client -> bot -> docs.
- The next slice after v59 should likely be the v60 engineering review unless project direction
  changes before then.
- If planning finds the Magic Bolt requirement path too large for one client bot scenario, keep the
  protocol proof full and make the client proof focused on disabled/enabled UI state with a test
  setup helper.
