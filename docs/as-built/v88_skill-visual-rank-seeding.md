# v88 As-built - Skill Visual Rank Seeding

## What shipped

`make skill-visual` accepts a `rank` parameter, defaulting to rank 1, and an optional `level`
parameter for visual replays.

The reusable `skill_visual` runtime now seeds the selected character before session creation with:

- the skill owner's class;
- the requested skill rank, bounded by `max_rank`;
- the minimum required level for that rank unless `level` is provided;
- base stats raised only as needed to satisfy rank requirements.

The replay no longer kills an XP setup dummy or spends a skill point before casting.

## What it proves

- Skill visual replays can inspect rank-scaled effects directly.
- Visual setup does not add extra combat or level-up noise before the skill cast.
- The seeding path stays debug-gated through the server's existing auth plus `X-Debug-Token`
  boundary.
- The default protocol bot suite skips the parameterized `skill_visual` scenario unless selected
  explicitly, so `make bot` and CI do not require skill-specific environment variables.

## Verification

```bash
.venv/bin/pytest tools/bot/test_skill_visual.py -q
cd server && go test ./internal/http
make skill-visual skill=holy_shield rank=5 DRY_RUN=1
.venv/bin/pytest tools/bot/test_protocol.py tools/bot/test_skill_visual.py -q
make ci
```

## Deferred

- Generated rank-5 visual proof rows still report as missing in `make skill-visual-list`.
