# Godot plugins, demos, and asset shortcuts

**Purpose:** Before building client-side Godot features from scratch, check whether an existing
plugin, demo, or asset pack can accelerate the slice — especially for **presentation-layer**
work (UI, camera, animation polish, placeholder art). This doc is a living catalog and an
adoption checklist, not an implementation plan.

**When to read:** At the start of any slice that touches `client/`, new visuals, inventory UI,
isometric movement feel, monster/client AI presentation, **Steam integration**, or multiplayer
lobby UX. Agents must consult this file **before** proposing or writing new client systems.

Last updated: 2026-06-05

---

## Why shortcuts matter here

This project is AI-agent-built (ADR-0001). Reinventing solved problems burns slice budget and
creates maintenance surface agents must forever understand. Community plugins and CC0 asset packs
are force multipliers **when they respect our architecture**:

| Layer | Who owns it | Shortcut viable? |
|-------|-------------|------------------|
| Combat, loot rolls, inventory state, persistence | Go server (authoritative) | **No** — do not adopt client inventory *logic* plugins as source of truth |
| WebSocket protocol + reconciliation | Custom `net_client.gd` | **No** — no drop-in multiplayer framework |
| Shared rules / golden fixtures | `shared/rules/` + Go/GDScript evaluators | **No** — data stays in repo contracts |
| Inventory **UI**, tooltips, drag-drop UX | Godot client (display only) | **Yes** — mirror server state, send intents |
| Animation / locomotion **presentation** | Godot client (ADR-0007) | **Partial** — we already have `animation_controller.gd`; evaluate before replacing |
| Camera, isometric feel, collision **prototypes** | Godot client | **Yes** — borrow patterns; our stack is **3D + orthographic**, not 2D tilemaps |
| Placeholder `.glb` / textures / audio for POC | `client/assets/` + manifest | **Yes** — CC0 packs; still wire through ADR-0006 manifest |
| Steam lobby UI, invites, GodotSteam wiring | Godot client (platform layer) | **Partial** — borrow lobby/matchmaking UX; game traffic stays on Go WebSocket |
| Headless smoke / unit tests | `client/tests/` | **Yes** — e.g. GUT |

**Golden rule:** Plugins are **shortcuts for client presentation and dev velocity**, never for
authoritative gameplay. If a plugin wants to own item stats, loot, or equip validation, use only its
UI/nodes and feed them from `state_delta` / `session_snapshot`.

---

## Adoption checklist (run before integrating anything)

Use this for every candidate plugin, demo, or asset pack:

1. **License** — MIT / Apache-2.0 / CC0 preferred. Record license in PR/spec. Avoid NC or “Asset Store only” for core dependencies.
2. **Godot version** — Client pins **Godot 4.6.x** (`client/project.godot`, `.godot-version`). Reject or fork if plugin targets ≤4.2 without 4.6 CI proof.
3. **Authoritative boundary** — Can we use it **display-only** without duplicating sim logic? If it hard-codes item power or local-only save, it’s a poor fit.
4. **Agent ergonomics** — Text-friendly (`.gd`, `.tscn`), documented API, minimal “click-only in editor” setup. C++/GDExtension is OK if prebuilt binaries exist for macOS/Linux/Windows.
5. **Maintenance** — Recent commits/releases, active issues, Asset Library presence.
6. **Integration cost** — Estimate: vendoring under `client/addons/`, adapter layer, headless smoke still green (`make client-smoke`).
7. **Slice scope** — Prefer “reference / copy pattern” over “full plugin dependency” for one-off needs.

Document the outcome in the slice **plan** (`docs/plans/…`): *adopt*, *borrow pattern*, or *reject (reason)*.

---

## Catalog by need

### Inventory & equipment UI

Server already emits `inventory_add`, `inventory_update`, equip state in snapshots. Client today
stores inventory in `main.gd` and resolves visuals via `equipment_visuals.gd` — **no grid UI yet**.

| Resource | Type | License | Notes | Fit for us |
|----------|------|---------|-------|------------|
| [Oen44/Godot-Inventory](https://github.com/Oen44/Godot-Inventory) | Project + autoload scenes | MIT | Universal inventory, equipment, affixes, tooltips, vendor UI. [Asset Store mirror](https://store.godotengine.org/asset/includespark/universal-inventory-system/). | **Borrow UI** — wire slots to server snapshot; **do not** use its itemization/affix logic as authority. Good for Diablo-like panels. |
| [peter-kish/gloot](https://github.com/peter-kish/gloot) | Addon (AssetLib) | MIT | ~925★; grid constraints, item stacks, UI controls; v3 for Godot 4.4+. | **Strong UI candidate** for grid/stash later; keep stack contents synced from server events only. |
| [expressobits/inventory-system](https://github.com/expressobits/inventory-system) | GDExtension addon | MIT | ~719★; separate UI/logic, multiplayer examples, craft/hotbar. | **Evaluate carefully** — “multiplayer” assumes client authority; strip to UI + intent sender. |
| [Wyvernbox](https://godotengine.org/asset-library/asset/1919) | Addon | MIT | ARPG-focused: affixes, merchants, crafting UI, ground items. | **UI/UX reference** for ARPG gestures (shift-click transfer, tooltips). Affixes must stay server-side. |
| [TheVulcoreTeam/RPGNodes](https://github.com/TheVulcoreTeam/RPGNodes) | Addon nodes | Check repo | Weight + Diablo slot inventories, stats nodes, GUT tests. | **Pattern reference** for slot grid; RPG stat nodes conflict with shared rules — avoid for stats. |

**Recommendation:** For first real inventory UI slice, spike **GLoot** or **Godot-Inventory** in a
branch: render server inventory, send `pickup_intent` / `equip_intent` on user action, assert no
local stat mutation. Reject plugins that cannot decouple UI from logic in \<1 day spike.

---

### Isometric camera, movement, and collision

ADR-0001 chose **3D low-poly + orthographic isometric camera**, not 2D tilemap RO-style sprites.
These demos are still valuable as **math/reference**, not drop-in movement code.

| Resource | Type | License | Notes | Fit for us |
|----------|------|---------|-------|------------|
| [Isometric Basics (Asset 2476)](https://godotengine.org/asset-library/asset/2476) | Demo | MIT | Grid setup; [companion video](https://youtu.be/dclc8w6JW7Y). | **Reference** for isometric projection math. |
| [Isometric Collision (Asset 2485)](https://godotengine.org/asset-library/asset/2485) | Demo | MIT | Walls + stay-on-grid; [video](https://youtu.be/8HvcHtauKoc). [GitHub mirror](https://github.com/Goldenlion5648/GodotIsometricCollisionExample). | **Reference** for 2D grid collision; adapt ideas to 3D `CharacterBody3D` + server-validated positions. |
| [2D Isometric Demo (Asset 112)](https://godotengine.org/asset-library/asset/112) | Official demo | MIT | Featured; depth sorting, occlusion, slide-around objects. Godot 3.5/GLES2 era. | **Art/layout inspiration**; port concepts to 3D Y-sort / camera, not the project itself. |
| [TileMapDual](https://github.com/pablogila/TileMapDual) | Addon | MIT | Dual-grid tilemaps (square, **isometric**, hex); separate display vs collision layers. | **Future environment slice** if we add 2D isometric tiles; not needed for current 3D `.glb` arena. |

**Recommendation:** Do **not** replace current 3D movement with 2D demo controllers. When tuning
feel, read 2476/2485 for grid↔screen mapping and collision UX, then implement against server
snapshots in `main.gd`.

---

### Animation & client-side state machines

ADR-0007 defines a **custom priority stack** in `animation_controller.gd` (terminal > one-shot >
locomotion), driven by protocol events — not by plugin state machines on the wire.

| Resource | Type | License | Notes | Fit for us |
|----------|------|---------|-------|------------|
| [LimboAI](https://github.com/limbonaut/limboai) | GDExtension | MIT | Behavior trees + HSM; 4.6 build on [AssetLib 4852](https://godotengine.org/asset-library/asset/4852). | **Optional** for rich **client-only** monster presentation or NPC ambient behavior. Server sim stays authoritative. |
| Built-in `AnimationTree` / `AnimationNodeStateMachine` | Engine | — | Native Godot. | **Prefer** for extending player/monster clips before adding LimboAI weight. |

**Recommendation:** Extend existing `animation_controller.gd` first. Consider LimboAI only when slice
needs client-side BT (e.g. critter wander) that must **not** affect combat outcomes.

---

### Steam, lobbies & future multiplayer (client platform layer)

ADR-0001 targets **Steam-publishable** builds via **GodotSteam / Steamworks GDExtension** (D4) and
defers full multiplayer orchestration, but the solo stack already runs through the production-shaped
**Go WebSocket session** (D2). A future multiplayer slice needs **lobby/matchmaking UX**, not a
second authoritative sim inside Godot.

| Resource | Type | License | Notes | Fit for us |
|----------|------|---------|-------|------------|
| [ViMayer/Godot-Steam-Local-Multiplayer-Lobby-Template](https://github.com/ViMayer/Godot-Steam-Local-Multiplayer-Lobby-Template) | Project template | MIT | GodotSteam addon; **Steam lobby** host/join (invite, lobby ID) and **local network** host/join (default `127.0.0.1:8080`, configurable in `Online.gd`). In-game lobby UI (Esc menu). Bundled `godotsteam` + Windows Steam API libs. | **Future — borrow lobby UI + GodotSteam bootstrap** only. Template ships a first-person demo that uses Godot’s built-in multiplayer sync — **do not** reuse that transport for combat/inventory. After lobby forms, clients should call our existing REST auth + `POST /v0/sessions` (or future party API) and connect `net_client.gd` to the **Go server** WebSocket. |
| [GodotSteam](https://github.com/GodotSteam/GodotSteam) (official) | GDExtension | MIT | Canonical Steamworks binding referenced in ADR-0001. | **Adopt when** Steam slice starts; prefer upstream docs/releases over vendoring from small templates alone. |
| Godot high-level **multiplayer** (`ENetMultiplayerPeer`, `@rpc`) | Engine | — | Peer sync, host migration patterns. | **Reject for game state** — conflicts with authoritative Go sim and replay (`make replay`). OK only if isolated to lobby signaling during spike (then delete). |

**How to use the ViMayer template without breaking architecture**

```text
Player opens game
      │
      ▼
Steam lobby OR local “play with friends” UI  ◄── borrow from template
      │
      ▼
Lobby agrees on session (host creates party / shares join code)
      │
      ▼
Go platform API: auth + session create/join  ◄── existing server (authoritative)
      │
      ▼
Each client: net_client.gd WebSocket + access_token  ◄── unchanged protocol
      │
      ▼
state_delta / intents — same as solo bot path
```

**Recommendation:** When a **Steam / multiplayer lobby** slice is specced (likely after ADR-0007
distribution notes and a dedicated netcode/multiplayer ADR), spike by importing **GodotSteam** plus
**lobby menu scenes** from the ViMayer template. Treat the template’s `Online.gd` local host/join as
a **reference for UI flow**, replacing its direct Godot multiplayer connection with our session API.
Verify headless CI still passes (Steam calls must be stubbed or gated when `OS.has_feature("steam")`
is false). Small repo (early-stage); cross-check patterns against official GodotSteam examples.

**Not in scope for this template:** server-side matchmaking, dedicated servers, anti-cheat, or
splitting realtime vs platform deployables (ADR-0001 D2, deferred ADR-0009).

---

### Placeholder art, audio, and POC assets

Production path is ADR-0006: manifest-driven `.glb` under `client/assets/`. Shortcuts must still
register in `assets/manifests/assets.v0.json` and `shared/assets/item_visuals.v0.json`.

| Resource | Type | License | Notes | Fit for us |
|----------|------|---------|-------|------------|
| [Kenney.nl assets](https://kenney.nl/assets) | 2D/3D/audio packs | CC0 | Modular Dungeon Kit, Platformer Kit, RPG UI, etc. Many packs appear on [Godot AssetLib (Kenney filter)](https://godotengine.org/asset-library/asset?filter=kenney). | **Excellent POC filler** — export to `.glb`, run `make gen-assets` / validate pipeline. |
| [Kenney Isometric Dungeon Tiles](https://kenney-assets.itch.io/isometric-dungeon-tiles) | 2D isometric | CC0 | 70+ tiles + sample character (8-dir). | **Reference art direction**; 2D sprites conflict with 3D pipeline unless explicit art pivot. |
| [Kenney Game Assets All-in-1](https://kenney.itch.io/kenney-game-assets) | Bundle | Paid / CC0 content | 60k+ assets; used by Expresso demos. | **Bulk POC** if team owns bundle; not required. |
| Expresso / GLoot / Wyvernbox **demo scenes** | Bundled Kenney GLBs | CC0 (demo assets) | Ready-made chests, items, UI sprites. | **Scavenge demos** for temporary UI icons and props; replace before Steam ship. |

**Recommendation:** For next visual slice needing “more game-like” arena, pull **Kenney Modular
Dungeon / Mini Dungeon** 3D kits into the manifest rather than procedurally generating placeholders.

---

### Testing & tooling

| Resource | Type | License | Notes | Fit for us |
|----------|------|---------|-------|------------|
| [GUT (Godot Unit Test)](https://github.com/bitwes/Gut) | Addon | MIT | Used by RPGNodes; in-editor + CLI tests. | **Adopt when** `client/tests/` outgrows hand-rolled smoke scripts; keep `make client-smoke` entrypoint. |
| Godot **headless** + `ARPG_AUTOPLAY` | Built-in | — | Already used by bot-visual / smoke. | **Keep** as CI gate over any new UI plugin. |

---

## What we should NOT shortcut

These are intentionally custom and tied to agent-playability / determinism:

- **Go authoritative sim** (`server/internal/game/`) — no Godot gameplay framework.
- **Protocol schemas** (`shared/protocol/`) — no generic RPC/codegen plugin without spec ADR.
- **Godot `@rpc` / ENet game sync** — no peer-authoritative combat or inventory; Go server owns outcomes.
- **Replay / bot scenarios** (`tools/bot/`, `make replay`) — must remain deterministic.
- **Loot formulas & item defs** (`shared/rules/`) — plugins’ affix/craft tables are not canonical.
- **Equipment visual resolver** (`equipment_visuals.gd`) — thin adapter over manifest; don’t replace with plugin equip logic.

---

## Suggested workflow for agents

```text
New client feature requested
        │
        ▼
Read this doc + ADR-0001 (D2 authority) + ADR-0007 (animation)
        │
        ▼
Search Godot Asset Library + GitHub (terms: inventory, isometric, ARPG, kenney, steam, lobby)
        │
        ▼
Run Adoption checklist ──reject──► Document reason in plan; build minimal in-repo
        │
     adopt / borrow
        │
        ▼
Spike in branch: addon under client/addons/ OR copy-only pattern
        │
        ▼
Prove: make client-smoke + make ci green; server authority unchanged
        │
        ▼
Record decision in slice plan + PROGRESS.md gaps if deferred
```

---

## Priority backlog (not approved — candidates only)

Ordered by likely impact on “feels like a game” vs integration risk:

| Priority | Candidate | Intended use | Blocker / note |
|----------|-----------|--------------|----------------|
| P1 | Kenney 3D dungeon/prop packs | Richer POC arena & props | Manifest + validate-assets wiring |
| P1 | GLoot or Godot-Inventory | First inventory/equip **UI** | Adapter to protocol intents |
| P2 | Wyvernbox gestures/tooltips | ARPG UX polish | Strip crafting/affix authority |
| P2 | GUT | Structured client unit tests | CI wiring |
| P3 | Isometric demos 2476/2485 | Movement/collision tuning notes | 3D adaptation required |
| P3 | LimboAI | Client ambient monster BT | Only if server AI slice deferred |
| P3 | ViMayer Steam lobby template + GodotSteam | Steam invites, lobby UI, local “play with friends” shell | Requires multiplayer ADR; strip Godot peer sync; stub Steam in CI |
| P4 | TileMapDual | 2D isometric world building | Conflicts with current 3D-only D1 unless ADR changes |

---

## Links index (quick reference)

**User-requested starting points**

- [Oen44/Godot-Inventory (GitHub)](https://github.com/Oen44/Godot-Inventory)
- [Isometric Basics — Asset 2476](https://godotengine.org/asset-library/asset/2476)
- [Isometric Collision — Asset 2485](https://godotengine.org/asset-library/asset/2485)
- [2D Isometric Demo — Asset 112](https://godotengine.org/asset-library/asset/112)
- [ViMayer — Steam & Local Multiplayer Lobby Template](https://github.com/ViMayer/Godot-Steam-Local-Multiplayer-Lobby-Template)

**High-signal additions**

- [GLoot](https://github.com/peter-kish/gloot)
- [Expresso Inventory System](https://github.com/expressobits/inventory-system)
- [Wyvernbox — Asset 1919](https://godotengine.org/asset-library/asset/1919)
- [RPGNodes](https://github.com/TheVulcoreTeam/RPGNodes)
- [TileMapDual](https://github.com/pablogila/TileMapDual)
- [LimboAI — Asset 4852](https://godotengine.org/asset-library/asset/4852)
- [Kenney assets](https://kenney.nl/assets)

**Steam / multiplayer (future)**

- [ViMayer — Godot Steam & Local Multiplayer Lobby Template](https://github.com/ViMayer/Godot-Steam-Local-Multiplayer-Lobby-Template)
- [GodotSteam](https://github.com/GodotSteam/GodotSteam)

**Internal**

- [ADR-0001 — Technology stack](../adr/0001-technology-stack.md) (authority, 3D isometric, agent-built)
- [ADR-0006 — Asset pipeline](../adr/0006-asset-pipeline.md) (manifest, `.glb`)
- [ADR-0007 — Animation state model](../adr/0007-animation-state-model.md)
- [PROGRESS.md](./PROGRESS.md) (current slice baseline)

---

## Maintaining this doc

When a slice adopts or rejects a plugin:

1. Move the row from **Priority backlog** to a new **Decisions log** subsection (add with date + outcome), or mark **Adopted** / **Rejected** inline.
2. If adopted, note vendored path (e.g. `client/addons/gloot/`) and adapter scripts.
3. Do not mark anything adopted here until `make ci` is green on the integrating branch.
