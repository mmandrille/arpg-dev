# v205 Spec: Boss Enrage Phase

Status: Complete
Date: 2026-06-15
Codename: boss-enrage-phase

## Purpose

Add a server-authored enrage transition for Cave Warden so boss fights change pace at low health. The first slice keeps the mechanic small and readable: when the boss reaches a configured health ratio, the server marks it enraged, emits an event, exposes the state in snapshots, and shortens future boss pattern cooldowns.

## Baseline

Builds on v204 and the existing boss floor system:

- Boss templates and pattern decks are data-driven in `shared/rules/boss_templates.v0.json` and `shared/rules/boss_patterns.v0.json`.
- The server already owns boss phase timing, damage, summons, and phase events.
- The client already renders boss phase/readability state through existing boss phase snapshots/events.

Asset/plugin decision: reject external assets/plugins for this slice. Enrage uses existing boss phase UI and optional debug/client color/status text; production boss VFX/audio remain future work.

## Non-goals

- No new boss template, random/weighted pattern selection, additional summon patterns, or new attack shapes.
- No direct damage multiplier increase in this first enrage slice; avoid surprise spike deaths and keep ADR-0014 fair-death rules intact.
- No production enrage art, audio, bespoke animation, or boss portrait work.
- No client-side combat authority; enrage state and timing remain server-owned.

## Acceptance criteria

- `shared/rules/boss_templates.v0.json` defines Cave Warden enrage data: threshold health ratio and cooldown multiplier.
- Shared schema/rule validation rejects invalid enrage settings such as threshold outside `(0,1]` or non-positive cooldown multipliers.
- When Cave Warden's HP falls to or below the threshold, the sim marks the boss enraged exactly once and emits a `boss_enraged` event with boss id/template id and threshold data.
- Boss entity snapshots/views expose an `enraged` boolean and enough threshold data for client/debug assertions.
- After enrage, newly completed boss patterns use the configured cooldown multiplier, with a floor of at least 1 tick when the base cooldown is positive.
- Existing telegraph-first damage behavior remains unchanged: active phases are still preceded by telegraphs, and enrage does not add instant damage.
- Bot proof damages Cave Warden below the threshold and observes `boss_enraged`; focused Go coverage verifies the enraged boss state.

## Scope and likely files

- Shared rules/schema: `shared/rules/boss_templates.v0.json`, `shared/rules/boss_templates.v0.schema.json`.
- Server sim/rules: `server/internal/game/rules.go`, `server/internal/game/boss_template_rules.go`, `server/internal/game/boss_patterns.go`, boss entity/view fields in `server/internal/game/sim.go`, `server/internal/game/companion_ai.go`, and `server/internal/game/types.go`.
- Protocol schemas: additive `enraged`/threshold fields in current state/snapshot schemas if entity/event views require schema coverage.
- Server tests: focused boss enrage validation/state/cooldown coverage near existing boss pattern tests.
- Bot scenario: update or add a protocol boss-floor scenario that observes enrage.
- Client: optional boss health bar/debug state only if existing state rendering needs an exposed field for client bot proof.
- Docs: v205 plan/as-built/lifecycle updates.

## Test and bot proof

- `make validate-shared`
- `cd server && go test ./internal/game -run 'TestBossEnrage|TestBossSummonedAdds'`
- `make bot scenario=boss_enrage_phase`
- `make bot scenario=boss_floor_gate`
- `make maintainability`
- Final `make ci`

## Open questions and risks

- No blocking questions.
- Risk: boss-floor protocol scenarios can become long if they wait for too many phases. Prefer a focused bot proof that uses existing damage/kill helpers to cross the threshold quickly.
- Risk: adding protocol fields requires schema examples/tests. Keep fields additive and limited to current v8 schemas unless existing validation requires older schema updates.
