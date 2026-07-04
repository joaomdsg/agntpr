# packets — MVP roadmap (slice ledger)

State file of the autonomous MVP loop (LOOP.md is the process, MVP.md
the spec). One slice per tick unless a slice says otherwise. Statuses:
queued / in-flight / landed / dropped. Every landed slice carries one
evidence line.

NEXT: slice 1.

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
  Evidence: this commit.
- [ ] **1. Token port** — queued. Replace the `--pk-*` values in
  internal/app/style.go with the MVP.md brand pack verbatim (new
  token names, IBM Plex fonts, keyframes, state-grammar colors);
  map existing classes onto the new tokens so every current surface
  re-skins without layout changes. Render tests pin token presence.
- [ ] **2. Mark + chrome** — queued. PacketMark built in code
  (locked spec incl. small-size rule) + the Titlebar pattern +
  nav rebrand (stacked lockup). Server-render tested.
- [ ] **3. Console shell** — queued. `/` becomes the 3-column
  Console (360|1fr|340): regions in place — needs-you rail,
  forwarded hero, in-flight, recently delivered, watches — fed by
  the data that exists today under the NEW vocabulary; dashed
  honest empty states where the mechanic doesn't exist yet (no lane
  health, no watches numbers). Old `/` card content retired.
- [ ] **4. Inspector shell** — queued. `/inspect/<packet>` becomes
  the 3-column Inspector (252|1fr|312): changed-files tree, Monaco
  rich diff, annotation rail (today's question threads reframed),
  Titlebar; timeline footer stubbed honestly. `/review` folds in
  and dies.
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
