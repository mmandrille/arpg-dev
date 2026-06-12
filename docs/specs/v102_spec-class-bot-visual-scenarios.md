# v102 Spec: Class Bot-Visual Scenarios

Status: Draft
Date: 2026-06-12
Codename: `class-bot-visual-scenarios`

## 1. Purpose

Add class-foundation bot-visual scenarios for Paladin, Barbarian, and Sorcerer so every playable
class has an approved scenario like `rogue_class_foundation`.

Each class-foundation scenario proves the class identity in one readable script: starter gear is
equipped, the hero moves/repositions, at least three basic attacks land, and every current class
skill is cast. Rogue remains the existing fourth class scenario and must conform to the same
coverage rule.

This slice also adds a guardrail for future content: every new class must ship with a
`*_class_foundation` scenario, and every new class skill must appear in that class scenario unless
the spec explicitly documents a temporary deferral.

## 2. Non-Goals

- Do not add new classes or skills.
- Do not rebalance skill damage, mana costs, cooldowns, weapon speed, or monster stats.
- Do not rename existing skill ids, including the current `ligthing` id.
- Do not build a cinematic/choreography framework beyond existing declarative bot scenario steps.
- Do not replace existing skill-visual scenarios; class-foundation scenarios complement them.
- Do not adopt external Godot plugins or asset packs. The plan should record this as a reject:
  existing protocol bot and Godot replay tooling are sufficient for this slice.

## 3. Acceptance Criteria

1. `paladin_class_foundation` exists as a protocol and bot-visual runnable scenario.
2. The Paladin scenario creates a Paladin, asserts starter sword + shield, moves/repositions, lands
   at least three basic attacks, casts `holy_shield`, and casts `heal`.
3. `barbarian_class_foundation` exists as a protocol and bot-visual runnable scenario.
4. The Barbarian scenario creates a Barbarian, asserts starter 2h axe, moves/repositions, lands at
   least three basic attacks, casts `rage`, and casts `cleave`.
5. `sorcerer_class_foundation` exists as a protocol and bot-visual runnable scenario.
6. The Sorcerer scenario creates a Sorcerer, asserts starter 2h staff, moves/repositions, lands at
   least three basic attacks, casts `magic_bolt`, casts `ice_shard`, and casts `ligthing`.
7. `rogue_class_foundation` still passes and is covered by the same class/skill validation rule.
8. Scenario validation fails if any class in `shared/rules/character_progression.v0.json` lacks a
   matching `*_class_foundation` scenario.
9. Scenario validation fails if any skill in `shared/rules/skills.v0.json` is missing from its
   class-foundation scenario, unless the scenario or validator has an explicit documented deferral.
10. The scenarios use debug progression or pre-purchased skill ranks where useful to keep the proof
    short and focused. Skill allocation is not required unless it makes the scenario more readable.
11. Timed effect endings are only waited for when the effect duration is under 10 seconds and the
    end state is part of the visual approval.
12. Every class-foundation scenario has visual metadata with a focused camera and post-complete hold.

## 4. Scenario Scripts

### Paladin

- Start as `paladin`.
- Assert `starter_paladin_sword` in `main_hand` and `starter_paladin_shield` in `off_hand`.
- Seed enough progression to cast `heal` and `holy_shield`.
- Move from spawn toward a compact target group.
- Cast `holy_shield` before or during engagement.
- Land three basic attacks with the main-hand sword.
- Cast `heal` after damage is available, or against a setup that makes the heal event visible.
- Do not wait for `holy_shield` to expire; its current duration is longer than 10 seconds.

### Barbarian

- Start as `barbarian`.
- Assert `starter_barbarian_axe` in `main_hand` and no incompatible offhand.
- Seed enough progression to cast `rage` and `cleave`.
- Move/reposition around a small target group so the axe user visibly closes distance.
- Cast `rage`.
- Land three basic attacks with the axe.
- Cast `cleave` against at least one target, preferably multiple targets if the lab layout can make
  that deterministic without slowing the run.
- Do not wait for `rage` to expire; its current duration is longer than 10 seconds.

### Sorcerer

- Start as `sorcerer`.
- Assert `starter_sorcerer_staff` in `main_hand` and no occupied offhand item.
- Seed enough progression to cast `magic_bolt`, `ice_shard`, and `ligthing`.
- Move/reposition or kite to show the ranged class moving before spell use.
- Land three staff basic attacks.
- Cast `magic_bolt`.
- Cast `ice_shard`.
- Cast `ligthing`.
- Wait for the cold slow end only if implementation can prove the duration is under 10 seconds and
  the end state is useful for visual approval.

## 5. Scope And Likely Files

- Add scenario JSON:
  - `tools/bot/scenarios/50_paladin_class_foundation.json`
  - `tools/bot/scenarios/51_barbarian_class_foundation.json`
  - `tools/bot/scenarios/52_sorcerer_class_foundation.json`
- Update existing scenario if needed:
  - `tools/bot/scenarios/47_rogue_class_foundation.json`
- Add or update bot scenario validation/unit coverage:
  - `tools/bot/test_protocol.py`
  - `tools/bot/run.py` only if a small reusable helper is needed for target selection, movement, or
    scenario discovery.
- Add or update shared world data only if existing labs are awkward:
  - `shared/rules/worlds.v0.json`
- Docs:
  - this spec
  - v102 plan
  - v102 as-built and `PROGRESS.md` during finish

No protocol schema bump is expected. No server gameplay code is expected unless an existing bot
step cannot express an already-supported player action.

## 6. Test And Bot Proof

Required targeted proof:

```bash
make bot scenario=paladin_class_foundation
make bot scenario=barbarian_class_foundation
make bot scenario=sorcerer_class_foundation
make bot scenario=rogue_class_foundation
make bot-visual scenario=paladin_class_foundation
make bot-visual scenario=barbarian_class_foundation
make bot-visual scenario=sorcerer_class_foundation
```

Required automated coverage:

- Python validation/unit coverage that every class has one class-foundation scenario.
- Python validation/unit coverage that every class skill is referenced by its class-foundation
  scenario.
- `make validate-shared` if any shared world data changes.
- `make test-py` for scenario validation changes.
- Final `/finish` should run the normal project close-out gate.

## 7. Open Questions And Risks

| ID | Question / Risk | Resolution |
|----|------------------|------------|
| Q-1 | Should class skills be pre-purchased or allocated in-scene? | Default: pre-purchase/debug-seed for speed unless allocation improves readability. |
| Q-2 | Should timed status effects wait until expiration? | Default: only wait when the effect duration is under 10 seconds and the ending matters visually. |
| Q-3 | Should scenarios share a new deterministic class-demo lab? | Default: yes if existing worlds make target layout or camera approval awkward. |
| Q-4 | Should Rogue be updated by this slice? | Default: yes, keep Rogue under the same validator. |
| R-1 | Existing `tools/bot/run.py` is large. | Keep changes small; extract focused helpers if scenario validation or target selection grows. |
| R-2 | Scenario timing can become flaky when many casts/attacks are chained. | Prefer deterministic debug progression, soft lab targets, focused target selection, and short scripts. |
