# ADR-0007: Client animation state model

Status: Accepted (2026-06-05)
Context: ADR-0001 (tech stack), ADR-0006 (asset pipeline), slice v3
`animate-and-react`.

## Decision

Animation is **client-side presentation state**, never authored on the wire.

- The local player's `idle/walk/attack` states are derived from signals already
  present in the client: movement input/prediction and the local attack input.
- The monster's `hit/death` states are driven by the **authoritative
  `monster_damaged` / `monster_killed` events** that the server already emits in
  `state_delta.events`. The client begins reading the `events` array; no new
  message type, schema, or sim change is introduced.
- States are **discrete clips** managed by a small priority state machine
  (`terminal > one-shot > locomotion`) in an injected `AnimationController`. No
  `AnimationTree`/blend spaces in this slice.
- The event→clip mapping is a client-only constant (`main.gd`), deliberately not
  in `shared/`, which is reserved for cross-language server/client contracts.

## Consequences

- Adding entity reactions later (e.g. player damage) requires the server to
  emit the authoritative trigger first; the client mapping then extends trivially.
- Because animation never crosses the wire, server tests and the protocol remain
  untouched (acceptance: empty `server/` diff).
