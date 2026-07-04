# packets — MVP roadmap (slice ledger)

State file of the autonomous MVP loop (LOOP.md is the process, MVP.md
the spec). One slice per tick unless a slice says otherwise. Statuses:
queued / in-flight / landed / dropped. Every landed slice carries one
evidence line.

NEXT: slice 5.

## Council verdicts

- **R-B1 (bootstrap, 2026-07-04, 3 personas, converged):** economy
  (spend/stock/bets/balance) retired from all surfaces; ledger stays
  as internal substrate, renamed only when touched. JetStream
  subjects + persisted JSON tags FROZEN; new concepts = new event
  kinds. One packet = one work-order identity (coarser/finer =
  stream version bump — don't). New thin `internal/packet` read-model
  aggregate; no ledger rewrite. `internal/refactor` deleted.
  Honesty gaps (lanes, precision scores, adversarial probes,
  delivered-ACK) built real or shown as honest absences — never
  faked. `/board` + `/review` + `/` fold into Console/Inspector;
  `/settings` survives as a utility page. Handshake = protected
  paths enforced in settle (deny-rule beside the secret scrub).
  Gate G3 = mutation vs handshake tests; G5 = mutation vs agent
  tests (same engine, two scopes). ACK = explicit host command
  backed by a re-checkable exit code. Probes run against a scratch
  ledger session, never the real economy.

## Slices

- [x] **0. Bootstrap** — landed. Reorg committed (old docs →
  old_stale/, design/ added); LOOP.md + MVP.md + ROADMAP.md;
  `internal/refactor` deleted; gate baseline green.
  Evidence: 479d04d; gate-flake fix (real-cage test deadlines
  30s→180s, this machine runs a caged cycle in ~28s) landed after,
  full suite green.
- [x] **1. Token port** — landed. Full brand pack in style.go under
  design-system names (state grammar verbatim, IBM Plex, 3
  keyframes, marketing tokens excluded); ~300 usages rewired, zero
  layout/selector changes; tests pin exact hexes + no `--pk-`
  survivors. Evidence: full gate green; commit below.
- [x] **2. Mark + chrome** — landed. packetMark/packetMarkHeld/
  packetLockup helpers (internal/app/mark.go) per the locked spec —
  ghost TR at ≥14px cells, solid --delivered-mid below, --mark-cell
  parameterized CSS; nav home link now mark + stacked lockup with
  per-surface sub (console/inspect/settings). Full Titlebar pattern
  deferred to the Inspector shell (slice 4), where it lives.
  Evidence: full gate green; commit below.
- [x] **3. Console shell** — landed. `/` is the 3-column Console
  (360|1fr|340, internal/app/console.go): needs-you rail (open
  threads, capped, victory empty state), hero `packets verified`
  (Done count — forwarded/delivered deliberately NOT claimed yet),
  settled rail, honest dashed empties for calibration + watches;
  center column preserves the tested act-now/state sections; open-
  thread count folded into the existing SSE poll signature.
  Follow-up noted: settled-row "missed" renders amber (--held) —
  revisit the color semantics when hold states become real (slice
  10). Evidence: full gate green; commit below.
- [x] **4. Inspector shell** — landed. /review renders the 3-column
  Inspector (252|1fr|312, internal/app/inspector.go): identity
  strip (wo#/key, short base→fix rev chip omitted when unknown,
  repo folder name), file tree left (honest empty when unscoped),
  Monaco + answer form center (islands untouched), annotation rail
  right (threads as agent/question annotation cards keeping
  review-thread classes + data anchors; adjustment ✎ zone below),
  honest timeline footer. Zero pre-existing test edits needed.
  Route rename + owner/repo addr deferred (slices 15 / 5).
  Evidence: full gate green; commit below.
- [ ] **5. Packet aggregate** — queued. `internal/packet`: Packet
  {name, addr, intent, revs, lane?, gate results?, lifecycle}
  folded from existing fabric/ledger events; addr = `owner/repo`
  binding for configured repos; lifecycle states
  composing→in-flight→verified→held→delivered mapped from existing
  events (delivered unreachable until slice 13). No new UI.
- [ ] **6. Wire surfaces to packets** — queued. Console + Inspector
  read the Packet aggregate (queue = held, in-flight, delivered
  rail; Inspector timeline from the fold). Economy meters
  (stock/balance/bets/bandwidth wallet) leave the UI.
- [ ] **7. Lanes from blast radius** — queued. Pure host-side
  function: `go list` import-graph reverse-dependency weight of a
  packet's changed packages → lane (best-effort/standard/strict/
  irreversible). Lane chip on Console/Inspector + lane health grid,
  all computed, never self-reported.
- [ ] **8. Gauntlet record** — queued. One explicit per-packet
  pipeline record of the six gates; map existing machinery (G2
  handshake run, G3 mutation-vs-spec, G4 build/vet, G5
  mutation-vs-agent-tests, G6 cage re-derivation + go vet/static
  method diversity; G1 = human intent-fidelity affordance in
  Inspector). Lane decides which gates run. New event kinds; gates
  surfaced on the Inspector timeline + Console.
- [ ] **9. Handshake mechanics** — queued. Protected `handshake/**`
  paths: settle deny-rule (agent turn cannot modify), content-hash
  check before gates, authored-before-code ordering enforced,
  strength gradient recorded (examples vs properties). Compose flow
  asks for/creates the handshake first.
- [ ] **10. Hold/forward** — queued. Forward-by-default; amber
  advisory holds (sampled) vs red blocking holds (strict lane,
  guardrail, handshake-below-lane-floor); needs-you queue driven by
  real holds with one-clause why-held strings; pk-held-pulse.
- [ ] **11. Attention economics** — queued. Interrupt budget
  (n/week counted down by real interrupts, rendered as the Console
  KPI), calibration draws (random sample of auto-forwarded packets
  in the queue rail), empty-queue-is-success copy. Bandwidth
  machinery reframed/renamed at this touch-point.
- [ ] **12. Watches + precision** — queued. Standing watch/capture
  triggers with precision scores counted from real fired-vs-useful
  history; "no history yet" until real; noisy triggers lose
  interrupt rights.
- [ ] **13. ACK/deliver** — queued. `packets deployed|regressed`
  host command backed by a re-checkable check run; delivered state
  reachable; the mark's BR cell fills only on real ACK; regressed
  routes back to held.
- [ ] **14. Adversarial probes** — queued. Seeded known-bad packet
  through the REAL gates via a scratch ledger session (never the
  real economy); probe report (caught/escaped) surfaced honestly.
- [ ] **15. Vocabulary sweep + retirement** — queued. Banned-word
  render test (MVP.md list) across all surfaces; kill remaining
  casino/PR vocabulary; retire dead routes/flags; voice QA (mono
  machine strings, lowercase, `·` counts, trailing →).
- [ ] **16. Final gauntlet sweep** — queued. Adversarial review of
  slices 1–15 (Sonnet reviewers), doc freshness (MVP.md, README),
  demo script: compose → forward → hold → inspect → deliver on a
  real local repo end to end.
