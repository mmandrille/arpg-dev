# ADR-0001: Technology Stack & Foundational Architecture

- **Status:** Accepted
- **Date:** 2026-06-04
- **Deciders:** Project owner (PM / tech lead)
- **Context tags:** action-RPG, looter, isometric, AI-built, multiplayer-ready

---

## Context

We are building an action RPG (Diablo-like, looting- and inventory-customization-centric)
with the following hard requirements:

1. **Genre/feel:** isometric, simple/readable graphics, inspired by Ragnarok Online (2005).
2. **Architecture:** client and server are two separate apps from the start (two folders in one
   repo). The first playable slice may be solo content, but it runs through the same remote-capable
   server boundary, auth/session flow, telemetry, and realtime protocol that multiplayer will use.
3. **Distribution:** the client must run on macOS, Windows, and Linux, and be Steam-publishable.
4. **Backend:** designed as a deployable web service from day one, while still runnable locally
   for development through the same interfaces.
5. **Development model:** the owner acts as PM / tech lead; **almost all code, assets, and
   content are produced by AI agents** (Cursor, Claude, Codex). This is the dominant constraint
   and biases every choice toward **text-based, code-first, large-training-corpus** technologies
   and against editor-bound or visual-scripting tools.
6. **Assets:** graphic assets are AI-generated; we must pick a format and pipeline that AI tools
   and agents can actually produce and process.
7. **Agent-playability & observability:** agents must be able to *play* the game during
   development to find and fix bugs, with deep, structured logging/debugging of game sessions
   and the backend so an agent can understand what happened when something breaks.

### Key tension resolved during design

The nostalgia target (RO-2005 = 2D sprites) conflicts with the production model
(AI-generated assets + AI-built pipeline). AI image generation is excellent at one-off images
but unreliable at the consistent 8-direction × multi-frame sprite sheets a 2D ARPG needs.
A **3D low-poly pipeline reproduces the RO look/angle** while being dramatically more
sustainable for AI generation (animate a rig once; render every angle) and more agent-friendly
(text-adjacent glTF assets, scriptable Blender). We therefore chose 3D over true 2D sprites.

---

## Decisions

### D1 — Rendering: 3D low-poly under an isometric/orthographic camera
Real 3D models, isometric camera to mimic the RO/Diablo angle. Each rig is animated once and
rendered from any direction. Art-direction constraints (poly budget, limited palette, flat-ish
shading) are codified to **force** the "simple, readable" look rather than relying on sprite
fidelity.

**Rejected:** true 2D sprites (AI multi-angle animation coherence is unsolved at quality);
hybrid 2D-sprites-on-3D-tiles (same 2D coherence problem + extra engine complexity).

### D2 — Thin client / authoritative server app, from day one
The client is a renderer + input layer. The **backend owns all authoritative persistent state**:
inventory, loot rolls, character progression, HP, persistence. Even in solo play, the client talks
to the server through the production-shaped protocol and session model. Local development may run
the server on localhost, but localhost is only a deployment target, not a different architecture.

The server includes the multiplayer-critical foundations immediately: authentication, account
identity, session creation/resume, authoritative game sessions, persistence, structured logs,
metrics, and health endpoints. Multiplayer later should add more player/session orchestration,
matchmaking, scaling, deployment, and abuse controls; it should **not** require replacing the
client/server authority boundary.

One backend process today serves **two logical roles** — a *realtime game server* (stateful
WebSocket session) and *platform services* (auth, accounts, persistence, loot ledger, metrics).
They are split into separate deployables only when multiplayer scale or operational isolation
demands it (YAGNI).

**Rejected:** client-authoritative with persistence-only backend (retrofitting authority later =
rebuilding the sim); local-only backend now with hosted architecture later (creates a migration
project before multiplayer can even begin).

### D3 — Real-time loop: client prediction + server confirmation
The client predicts movement locally for snappy feel and may play immediate attack/skill feedback
animations. The backend validates against shared rules and owns combat outcomes (hit/miss, damage,
death, loot, HP), then the client reconciles. This avoids jarring rollback of loot/combat facts
while keeping input feel responsive. Requires a persistent low-latency channel.

**Rejected:** full server-resolves-every-tick (too heavy / laggy for a solo-first project);
combat fully client-side (trivially cheatable in multiplayer).

### D4 — Client engine: Godot 4 (GDScript)
- Scene/resource files are text (`.tscn`/`.tres`) → agents can diff and edit them, not just
  poke a GUI.
- GDScript is Python-like → matches the owner's background and agent strengths.
- Open-source → no licensing rug-pull on a multi-year hobby project.
- Turnkey Windows/macOS/Linux export; Steam via **GodotSteam / Steamworks GDExtension**.
- **Headless mode** enables automated/agent-driven client runs (see D8).

**Rejected:** Unity (editor-centric workflow + merge-hostile scenes fight pure-agent dev;
licensing trust issues); Unreal (Blueprints are visual = agent-hostile; C++ overkill for
low-poly); Bevy/Rust (small corpus + churning API = poor bet for AI-led dev today).

### D5 — Backend: Go
- Single static binary → simple local/dev execution and simple service deployment.
- Real concurrency for when sessions multiply.
- Strong agent support.
- Mature HTTP/WebSocket, auth, metrics, profiling, and deployment ecosystem.

One process now (both roles from D2). Go and GDScript cannot share code, which makes D6
(shared data contracts) mandatory rather than optional.

**Rejected (for the realtime server):** Python/FastAPI (simpler and the owner's wheelhouse, but
the owner prioritized building correctly for the multiplayer future over fastest-prototype);
Godot-headless-as-backend (max code reuse but weak for DB/accounts and off-the-beaten-path to
host); Django (request/response model fights the stateful WebSocket game loop).

### D6 — Shared rules-as-data (`/shared`)
Combat formulas, item schemas, and loot tables live as **data contracts** (JSON / schema files)
consumed by **both** the Go server and the Godot client. The *rules* live in one place as data;
the client predicts display-relevant behavior with them, the server validates authoritative
outcomes with them. This reduces the "two sim implementations drift apart" risk inherent in a
two-language stack, but does not eliminate it by itself.

Required enforcement:
- Versioned JSON schemas for all shared data.
- Golden fixtures consumed by both Go and GDScript tests.
- CI validation that shared data is schema-valid and produces identical expected outcomes for
  representative formulas, item rolls, and skill definitions.
- Declarative formulas only; no hidden executable logic in client-only or server-only data files.

**Formula evaluation mechanism (direction; finalized in the shared-rules ADR).** Formulas are
expressed as a **bounded catalog of parameterized formula types** (e.g. `linear{base, per_level}`,
`weighted_table{entries[]}`, `curve{points[]}`, `clamp{min, max}`), **not** a free-form expression
language. Both Go and GDScript implement the same small evaluator set; shared data only selects a
known type and supplies parameters. This keeps cross-language equivalence tractable — a free-form
DSL or embedded script would reintroduce the very drift risk we are trying to kill, now hidden
inside two interpreters — and makes golden-fixture equivalence exhaustive over the catalog. A shared
expression DSL / embedded scripting language is explicitly deferred unless the catalog proves
insufficient in practice.

### D7 — Asset pipeline: tiered generation, glTF/`.glb`, Blender hub
- **Format:** glTF/`.glb` for all 3D models, PNG for textures. **Blender is the conversion/cleanup
  hub, driven by Python scripts agents can run.**
- **Tier A — characters + equipment:** a shared base humanoid rig with modular/swappable mesh
  parts, so gear renders on the body, animations are authored once, and visual consistency holds.
  (Mandatory for a looter — gear-on-character is half the genre's dopamine.)
- **Tier B — monsters / props / environment:** freer AI text-to-3D (Meshy / Tripo / Rodin) +
  AI textures, with scripted Blender retopo/cleanup.

**Rejected:** full AI text-to-3D for everything (gear-on-character consistency + art-style
coherence too hard); concept-art-then-manual-modeling (highest control but contradicts the
"agents build almost everything" goal).

### D8 — Observability & Agent-Playability (foundational pillar)

This is treated as a first-class architectural pillar, not a feature, because the most valuable
parts are **near-impossible to retrofit**.

**D8.1 — Deterministic authoritative simulation (Go server only).**
The authoritative sim uses **one seeded PRNG stream**, **fixed-tick** updates, **stable ordering**
(iterate by entity ID, never by Go map order), and **no wall-clock time in game logic** (use the
tick counter). The game remains **fully random to players** — each real session draws a fresh seed
from OS entropy, so loot and combat hit/miss are exactly as unpredictable as in any game.
Determinism only means *given a recorded seed + inputs, the server reproduces the same outputs*.
The discipline is **scoped to the Go server's authoritative logic** — the Godot client, rendering,
animation, and prediction layer are exempt, because the server owns every outcome that matters.

**Simulation timing (default).** **20 Hz fixed authoritative tick** (50 ms/step) — adequate feel for
a click-to-move isometric ARPG, cheap to run, and high-fidelity for replay. The client renders
**uncapped**, interpolating between server snapshots and applying local prediction. Inputs are
**tick-tagged and buffered**: the client stamps each command with its intended tick; the server
applies buffered inputs in deterministic order within the owning tick. **No rewind / "favor-the-
attacker" lag compensation in v1** — server-authoritative with prediction + reconciliation only.
Richer netcode (interpolation tuning, lag compensation, input delay budgets) is deferred to the
netcode ADR.

**Entity-ID policy (required for stable ordering).** Entities carry **monotonic, server-assigned
64-bit IDs** drawn from a **per-session counter advanced in deterministic spawn order** (spawns are
themselves driven by ordered inputs / tick events — never by wall-clock, pointer address, or map
iteration order). Authoritative iteration sorts by this ID. ID allocation is itself part of the
deterministic sim, so IDs reproduce identically under replay; without this, determinism leaks at
the spawn edges.

> Clarification (kept deliberately, because it drove the decision): "deterministic" ≠ "not random".
> All computer randomness is already a seeded PRNG; we simply own and record the seed instead of
> letting it be grabbed from the clock and discarded. Production play = fresh random seed every
> session (different loot for everyone). Debugging/testing = pin a known seed on purpose.

**D8.2 — Event-sourced session log → bit-for-bit replay.**
Every authoritative input command, seed, server build/config version, and resulting authoritative
state-change event is recorded in a per-session log. Re-simulation replays the same seed and input
stream through the authoritative Go logic, then compares the derived events/state against the
recorded output. Event playback can inspect history; input replay reproduces bugs. Together they
enable the core loop: *reproduce a bug → fix it → verify the fix by replaying the same seed and
inputs.*

**D8.3 — Structured logging, metrics + correlation IDs.**
Logs are structured JSON with a consistent schema (severity, component, `session_id`, `tick`,
`action_id`). A correlation ID spans the full causal chain client↔server per action/tick:
input → predict → send → server-receive → validate → apply rules → respond → client-reconcile.
Metrics are emitted from the first slice for server health, active sessions, tick duration,
message latency, reconciliation rate, persistence errors, and replay-test failures.

**D8.4 — Inspection / query API.**
An endpoint exposes live and historical session state as **structured data** (entity states,
recent event log, metrics) so agents inspect by querying, never by scraping pixels. In local
development it may be localhost-only; in any remote deployment it must require authentication and
authorization, and debug/inspection routes must not be publicly exposed by default. **Slice v1 ships
only the minimal live form — a read-only `GET /sessions/{id}/state`** returning current authoritative
state as structured JSON; richer historical/filtered querying is post-slice so agents don't
over-build early.

**D8.5 — Two-layer agent playability.**
1. **Headless Python protocol bot** — connects over the same JSON protocol as the real client,
   sends intents (`move`, `attack`, `pick_up`, `equip`), asserts on authoritative state. This is
   the **primary** way agents "play," and it doubles as the integration-test driver. (Agents play
   by speaking the protocol, *not* by watching a 60 fps screen — vision-driving a real-time game
   is slow, flaky, and reflex-bound.)
2. **Godot headless/automation + client debug API** — drives the real client and reads back
   *client-side* state (predicted position, reconciliation error, what is actually visible as
   data) to catch the client-only bugs the protocol bot cannot see (prediction drift, animation,
   rendering).

---

## Defaults (chosen, low-cost to revisit later)

- **Transport:** WebSocket for realtime; **JSON messages first** (fast to build, agent-trivial),
  migrate to a schema'd binary format once message shapes stabilize and perf warrants. Protobuf is
  the leading candidate, but the GDScript codec path (`godobuf`, hand-written codecs, or a
  GDExtension) must be verified for Godot 4 maintenance status before it is baked in — that decision
  belongs to the wire-protocol ADR, not here. REST/JSON for platform services.
- **Persistence:** **Postgres from day one**, run locally through dev infrastructure and remotely
  through managed/self-hosted Postgres later. A repository interface still isolates storage access,
  but SQLite is not the default because switching SQL dialects, migrations, locking behavior, IDs,
  and transaction semantics is not a config-only change.
- **Auth/session baseline:** email/dev-token auth for early development behind an interface, with
  account identity and session tokens represented in the protocol from the first slice. The exact
  production auth provider is a follow-up decision, but "anonymous local-only player" is not the
  architectural baseline.
- **Metrics baseline:** expose service health and metrics from the first server slice; wire them
  into local dashboards/CLI checks before adding feature complexity.

---

## Toolchain & pinned versions

This ADR records *intent*; the **canonical, exact pins live in repo tooling files** (the source of
truth), and CI enforces them. For an AI-led repo this matters more than usual: **no floating
"latest"** — agents must pin exact versions and treat upgrades as deliberate PRs, not silent drift.

| Component        | Baseline (verify latest stable at scaffold) | Source-of-truth pin                     |
|------------------|---------------------------------------------|-----------------------------------------|
| Godot            | 4.6.x stable (never `dev`/`rc`)             | `.godot-version` + CI export image      |
| Go               | 1.24.x                                      | `go.mod` `go` directive + CI            |
| Postgres         | 16.x                                        | `docker-compose` / migration tooling    |
| Python (tools)   | 3.12.x                                      | `.tool-versions` / `pyproject.toml`     |
| Blender (assets) | 4.2 LTS                                     | pipeline container/image                |

Exact patch versions are chosen at scaffold time and recorded in the files above, deliberately
**not** hard-coded narratively in this ADR (which would rot). The baselines are starting points to
verify against current stable when the repo is scaffolded.

---

## Repository structure

```
arpg-dev/
├── client/        # Godot 4 project (GDScript)
├── server/        # Go module — realtime game server + platform services
├── shared/        # data contracts: item schemas, loot tables, skills, damage formulas, message schemas
├── tools/         # Python agent-bot (headless protocol client) + replay/inspection CLIs
├── assets/        # source art + Blender pipeline scripts (Python)
└── docs/
    ├── adr/       # architecture decision records (this file = 0001)
    └── sdd/       # specs & plans per feature
```

---

## First vertical slice (proves every decision end-to-end)

Solo session: spawn in an isometric scene → walk → attack a monster
(**client predicts movement/attack feedback; the Go server confirms** hit, damage, and death via a
**seeded** roll) → the monster drops loot (**Go rolls from a `/shared` loot table**) → the item
enters the **backend-owned** inventory → **persists to Postgres** → survives a restart.
Before entering the session, the client authenticates, receives/resumes a session token, and the
server emits logs/metrics for the full flow.
The session is **recorded and replayable by seed**, and a **headless Python agent-bot can run the
entire flow through the same auth + protocol path and assert on outcomes**.

**Visuals are placeholder in slice v1 (scope guard).** Characters/monsters render as primitives
(capsule/box) purely to prove the architecture. Slice v1 intentionally does **not** prove D1's art
direction or D7 Tier A (modular rig + gear-on-body) — conflating architecture proof with
art-pipeline proof would bloat v1. Those are proven in **slice v2 — "equip and see it"**: import an
AI-generated glTF base rig + one modular gear piece, equip it through the backend-owned inventory,
and see it render on the character under the isometric camera. D7 specifics are owned by the
asset-pipeline ADR.

**Inspection in v1 (minimal).** The server exposes the read-only `GET /sessions/{id}/state`
(D8.4, minimal form) so agents inspect by querying structured JSON, not pixels. Historical/filtered
query is deliberately out of slice v1.

This slice exercises: D2 (authoritative state), D3 (predict/confirm), D4+D5 (Godot↔Go over the
protocol), D6 (shared loot table), D8.1–D8.3 (seeded determinism + replay + metrics), D8.4 (minimal
state endpoint), D8.5 (agent-bot). **Deferred to slice v2:** D1 (art direction) and D7 Tier A
(modular rig + gear-on-body).

---

## Consequences

### Positive
- The project starts with the real client/server, auth/session, persistence, metrics, and
  authoritative-simulation boundaries multiplayer needs.
- Two-language stack is kept honest by shared data contracts (D6).
- Seeded determinism + replay gives agents a real *reproduce → fix → verify* debugging loop —
  the linchpin of the AI-built thesis — without sacrificing any player-facing randomness.
- 3D low-poly + glTF + scriptable Blender makes AI asset generation sustainable and agent-drivable.
- All foundational choices favor text-based, agent-friendly tooling.

### Negative / costs
- Client/server plumbing tax is paid even in solo play.
- Auth, sessions, metrics, and Postgres arrive before they are strictly needed for a solo
  prototype.
- Go is not the owner's wheelhouse and adds ceremony vs. a Python prototype; slower to first
  playable.
- Determinism imposes permanent discipline on the server sim (seeded RNG only, no wall-clock in
  game logic, stable ordering, fixed tick).
- Two languages (Go, GDScript) + Python tooling = three languages to maintain.

### Risks & mitigations
- **Sim drift between client prediction and server truth** → mitigated by D6 (shared rules-as-data,
  schemas, golden fixtures, cross-language tests).
- **Auth/session work slows the first playable** → mitigated by using a minimal dev-token/email
  flow behind an interface while keeping account/session concepts in the protocol.
- **Postgres adds local setup friction** → mitigated by dev scripts/containers and treating the DB
  as part of the first vertical slice, not later infrastructure.
- **Godot 3D asset ecosystem is smaller than Unity's** → mitigated; assets are AI-generated as
  glTF anyway.
- **Determinism violated accidentally** (e.g., a stray `time.Now()` or map-order iteration in game
  logic) → mitigated by lint/review rules and replay-based regression tests in CI.
- **Protobuf/`godobuf` migration friction** → deferred until shapes stabilize; JSON first.

---

## Follow-up ADRs (anticipated)

- ADR-0002: Wire protocol & message schema (JSON now; Protobuf migration criteria; verify the
  GDScript codec path — `godobuf` vs hand-written vs GDExtension, incl. Godot 4 maintenance status).
- ADR-0003: Shared-rules contract & formula-evaluation mechanism (the parameterized evaluator
  catalog from D6; schema versioning; golden-fixture cross-language equivalence).
- ADR-0004: Determinism enforcement & replay test harness in CI.
- ADR-0005: Netcode & simulation timing (lag compensation, input-delay budgets, snapshot/
  interpolation tuning beyond the 20 Hz v1 default).
- ADR-0006: Asset pipeline specifics (base rig spec, modular attachment points, text-to-3D
  toolchain) — also owns D7 Tier A and slice v2.
- ADR-0007: Distribution & platform matrix (Godot export targets, GodotSteam integration, Linux
  Steam runtime / Proton considerations, code signing per OS).
- ADR-0008: Auth, account identity, session lifecycle, and authorization model.
- ADR-0009: Multiplayer split criteria (when realtime server and platform services separate).
