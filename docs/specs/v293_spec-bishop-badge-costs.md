# v293 Spec: Bishop Badge Costs

Status: Complete
Date: 2026-06-19
Codename: `bishop-badge-costs`
Baseline: v292 `badge-reward-foundation`

## Purpose

Make bishop respec and revive-all services consume the new account-wallet badges instead of being
free normal-play services. Respec consumes `respec_badge`; revive-all consumes
`resurrection_badge`. Both costs are data-driven in shared main config and surfaced in the existing
bishop panel.

## Non-goals

- Do not implement stat-badge or skill-badge spending.
- Do not implement character selection, durable corpse resurrection, death recovery UX, or revive
  results beyond the existing `bishop_revive_all` no-op service event.
- Do not remove debug-only bishop level/stat/skill-point actions; they remain gated by gameplay
  debug.
- Do not add a protocol schema version. Use existing state-delta `resource_wallet_update` and event
  JSON fields for resource cost metadata.
- Do not add production badge art or a new town service panel.

## Acceptance Criteria

- Shared rules define:
  - bishop respec resource ID/count: `respec_badge` x1,
  - bishop revive-all resource ID/count: `resurrection_badge` x1.
- Schema, Python validation, and Go load-time validation reject negative counts, missing resource IDs
  when count is positive, and non-currency/equippable resource IDs.
- Opening bishop service reports respec affordability from both existing gold cost and respec badge
  balance.
- Respec rejects with `missing_resource` when the wallet lacks the configured respec badge.
- Successful respec consumes one `respec_badge`, emits `resource_wallet_update`, resets stats/skills
  exactly as before, and records resource cost metadata on `bishop_respec`.
- Revive-all rejects with `missing_resource` when the wallet lacks the configured resurrection badge.
- Successful revive-all consumes one `resurrection_badge`, emits `resource_wallet_update`, preserves
  the existing `bishop_revive_all` event shape for revived amount, and records resource cost metadata.
- Bishop panel shows badge requirements, disables actions when the relevant badge is missing, and
  updates after wallet changes.
- Existing bishop bot/client scenarios are updated so they acquire badges through the v292 quest
  reward path before using the bishop.

## Scope And Likely Files

- Shared: `shared/rules/main_config.v0.json`, `shared/rules/main_config.v0.schema.json`.
- Validation: `tools/validate_main_config.py`, `server/internal/game/main_config_validation.go`,
  `server/internal/game/rules.go`.
- Server: extract `handleBishopRespec` from `handlers.go` into a focused bishop file, update
  `bishop_revive.go`, add wallet-cost helpers/tests.
- Client: `client/scripts/bishop_panel.gd`, `client/scripts/main.gd`, client bot bishop assertions
  and scenario JSON.
- Bot: update protocol `45_town_bishop_respec` to acquire a respec badge before respec; avoid
  growing `tools/bot/run.py`.
- Docs: v293 plan/as-built/lifecycle updates.

## Test And Bot Proof

Focused checks:

```bash
make validate-shared
(cd server && go test ./internal/game -run 'Bishop|MainConfig|BadgeReward|QuestTurnIn' -count=1)
godot --headless --path client --script res://tests/test_bishop_panel.gd
make bot scenario=45_town_bishop_respec
make bot-client scenario=32_town_bishop_respec_panel HEADLESS=1
make maintainability
```

Manual visual verification command:

```bash
make bot-visual scenario=32_town_bishop_respec_panel
```

## Asset And Plugin Decision

- Adopt: v292 account `resource_wallet`, badge item definitions, bishop town service, and existing
  bishop panel.
- Borrow: v292 deterministic depth-125 quest reward proof to acquire badges in bot scenarios.
- Reject: external assets/plugins, production badge icons, a new resource UI, and protocol schema
  versioning for this small service-cost slice.

## ADR Alignment

- ADR-0001: service costs stay server-authoritative and all outcomes are produced by the sim.
- ADR-0014: respec and resurrection become earned resources instead of free normal-play power resets.

## Open Questions And Risks

- No blocking questions. The default cost is one matching badge per bishop service use.
- The existing revive-all service still has no durable resurrection implementation; this slice only
  adds the resource gate and consumption foundation requested for that service.
