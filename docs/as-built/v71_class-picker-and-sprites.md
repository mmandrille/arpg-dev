# v71 As-Built — Class Picker and Sprites

Date: 2026-06-11

## What Shipped

- Character creation now exposes three selectable class blocks under the name field:
  Barbarian, Sorcerer, and Paladin.
- The selected class defaults to `barbarian`, can be switched before submit, and is sent as
  `character_class` to `POST /v0/characters`.
- Each class block has a code-native sprite icon and tooltip with class name, class skill, and
  starting stats.
- Character picker rows now render a class icon before the row text and include class data in the
  debug state used by tests and bots.
- Client bot support can select a class during create flow; the create/join menu scenario proves a
  Sorcerer create path and later row class presentation.

## Proof

- `make client-unit`
- `make ci`

`make client-smoke` was also attempted standalone, but the smoke login failed because no local
server was running. The same smoke path is covered by `make ci`, which starts Postgres and the
server before the client phase.

## Deferred

- Production bitmap portraits or model swaps remain deferred.
- Shared class presentation metadata remains a future cleanup if class visuals expand beyond this
  focused picker UI.
