# Data-driven content libraries

Status: Draft design note
Date: 2026-06-10
Context: v59 `data-driven-skill-catalog`; candidate input for the v60 engineering review

## Purpose

This note records the intended direction for moving game content libraries toward a fuller
data-driven model. The goal is to make routine content additions possible by adding validated data
files, while keeping new game capabilities explicit in code, schema, tests, and golden fixtures.

## Core rule

File-path dictionaries should be **indexes/manifests**, not the gameplay model itself.

Gameplay identity remains stable IDs:

- `rusty_sword`
- `cave_blade`
- `magic_bolt`
- `dungeon_mob`
- `town_vendor`

File paths only organize and load those definitions. Runtime contracts, persisted rows, protocol
payloads, loot tables, shop offers, equipment visuals, skill ranks, and replay data should refer to
stable IDs, not file paths.

## Current state

Items are already mostly data-driven:

- `shared/rules/items.v0.json` contains fixed item definitions.
- `shared/rules/item_templates.v0.json` contains rolled equipment templates.
- `shared/assets/item_visuals.v0.json` keeps equipment visuals separate from gameplay.
- `shared/assets/item_presentations.v0.json` keeps item icon/ground presentation separate from
  gameplay.
- `client/scripts/item_rules_loader.gd` already follows the shared singleton loader pattern.

Skills are partially data-driven:

- `shared/rules/skills.v0.json` contains the Magic Bolt definition.
- The server spend/cast path is mostly generic around skill IDs.
- Before v59, server validation still special-cases `magic_bolt`, and the client skill panel/bar
  hardcode Magic Bolt identity, label, and placement.

## Target manifest shape

A future library index can describe where content definition files live while preserving stable IDs
inside those files:

```json
{
  "version": 0,
  "rules": {
    "items": {
      "fixed": [
        "items/fixed/currency.v0.json",
        "items/fixed/consumables.v0.json",
        "items/fixed/quest.v0.json"
      ],
      "equipment_templates": [
        "items/equipment/weapons.v0.json",
        "items/equipment/armor.v0.json",
        "items/equipment/jewelry.v0.json"
      ]
    },
    "skills": {
      "mage": [
        "skills/mage/projectiles.v0.json",
        "skills/mage/utility.v0.json"
      ],
      "warrior": [
        "skills/warrior/melee.v0.json"
      ]
    },
    "classes": [
      "classes/classes.v0.json"
    ]
  },
  "assets": {
    "items": {
      "visuals": [
        "items/visuals/equipment.v0.json"
      ],
      "presentations": [
        "items/presentations/icons.v0.json"
      ]
    },
    "skills": {
      "presentations": [
        "skills/presentations/icons.v0.json",
        "skills/presentations/projectiles.v0.json"
      ]
    }
  }
}
```

The manifest is a load-order and organization tool. The merged in-memory model should remain close
to today's `Rules.Items`, `Rules.ItemTemplates`, `Rules.Skills`, and presentation maps.

## Loader requirements

A future loader should:

- load manifests in deterministic declared order
- resolve paths relative to the manifest file
- reject duplicate IDs across all loaded content files
- reject unknown top-level groups unless the schema declares them
- validate every loaded file against its own schema before merging
- validate cross references after merging
- preserve stable sorted ordering where views are exposed through protocol or replay
- keep file paths out of runtime state, persistence, and protocol payloads

## Capability rule

Adding content is data-only when it uses an existing supported capability:

- new fixed item with known category
- new equipment template with known slot/stat/requirement types
- new skill with known `kind`, requirement, cost, damage, projectile, and cooldown helpers
- new item or skill presentation metadata

Adding behavior requires code and contracts:

- new item stat with gameplay effect
- new skill formula helper
- new skill kind such as aura, trap, summon, chained projectile, or DOT
- new targeting model
- new persistence or protocol surface
- new loot/shop/economy rule that changes authoritative outcomes

Behavior additions must include:

- schema change
- Go authoritative implementation
- GDScript prediction/display support if the client consumes it
- `tools/validate_shared.py` cross-checks
- golden fixture updates when formulas or deterministic outcomes change
- protocol/client bot proof when gameplay or protocol behavior changes

## Candidate rollout

1. v59: make Magic Bolt a catalog-driven skill entry and split skill presentation metadata.
2. v60 review: evaluate whether the repo is ready for a library manifest slice.
3. Follow-up slice: introduce a manifest for skills only, with no behavior change.
4. Follow-up slice: split item fixed definitions/templates by family behind a manifest, with no
   behavior change.
5. Follow-up slice: add class definitions and class-to-skill tree availability.

## Non-goals for the manifest itself

- No executable scripts in JSON.
- No free-form formula expression language.
- No persisted file-path references.
- No plugin-owned gameplay authority.
- No compatibility layer for stale local development data unless a production migration exists.
