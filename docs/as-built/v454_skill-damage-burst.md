# v454 As-Built — skill-damage-burst

Date: 2026-07-08

## Shipped

- `skill_damage_burst` event with `hits[]` on wire; Volley collapses per-hit `monster_damaged`.
- `resolution` field on skills (`instant_ray` for magic_bolt, `instant_aoe` for volley).
- Magic Bolt instant line resolve on cast tick; no projectile entity spawn.
- Client burst → per-hit combat presentation; bot ingest expands bursts for assertions.
- ADR-0016 combat processing budget (proposed).

## Verification

```bash
make validate-shared
cd server && go test ./internal/game/... -run 'RangerVolley|MagicBolt' -count=1
make bot scenario=ranger_volley_and_visual_showcase
make bot scenario=32_skill_points_and_magic_bolt
```
