# packets — MVP roadmap (slice ledger)

State file of the autonomous MVP loop (LOOP.md is the process, MVP.md
the spec). One slice per tick unless a slice says otherwise. Statuses:
queued / in-flight / landed / dropped. Every landed slice carries one
evidence line.

NEXT: none — 17 slices landed (16 from the original MVP checklist,
plus slice 17 added by maintainer directive after the autonomous
loop stopped). MVP checklist complete (MVP.md); the autonomous loop
stops scheduling further ticks per LOOP.md's Schedule rule. Any
further work is the maintainer's call.

## Infra

- **Test suite speed** (2026-07-04, unscheduled — user directive to
  fix holistically). The old gate (`-p 1 ./...`) serialized ALL 22
  packages because `internal/cage`+`internal/sandbox` share a docker
  container label (`io.packets.sandbox=1`) that only THOSE two need
  isolated from each other. Split into `scripts/test-fast.sh`
  (local) + `test-fast`/`test-cage` CI jobs (`.github/workflows/
  ci.yml`): the other 20 packages run at full parallelism, no
  docker dependency, concurrently with the serialized cage+sandbox
  pair. Measured cold: ~45s (fast set) vs ~110s (cage+sandbox
  serialized) running concurrently → ~1m53s wall vs. the prior
  blanket-serial gate. LOOP.md's gate section updated to
  `./scripts/test-fast.sh`. NOTED, not chased: a pre-existing flaky
  `TestPruneProducerObjects_leavesOtherProducersUntouched`
  (internal/ingest, `TempDir RemoveAll: directory not empty`) —
  didn't reproduce on repeat runs, unrelated to this change.
- **Fable escalation added to LOOP.md** (2026-07-04, user directive):
  `model: fable` may be dispatched ONLY as a last resort when a
  Sonnet builder has genuinely stalled on the same slice after a
  real retry (not merely hit a session limit) — extremely rare,
  logged here whenever used.

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

- [x] **0. Bootstrap** — landed. Reorg committed (superseded docs
  archived out of tree, design/ added); LOOP.md + MVP.md + ROADMAP.md;
  `internal/refactor` deleted; gate baseline green.
  Evidence: 479d04d; gate-flake fix (real-cage test deadlines
  30s→180s, this machine runs a caged cycle in ~28s) landed after,
  full suite green.
- [x] **1. Token port** — landed. Full brand pack in style.go under
  design-system names (state grammar verbatim, IBM Plex, 3
  keyframes, marketing tokens excluded); ~300 usages rewired, zero
  layout/selector changes; tests pin exact hexes + no `--pk-`
  survivors. Evidence: full gate green.
- [x] **2. Mark + chrome** — landed. packetMark/packetMarkHeld/
  packetLockup helpers (internal/app/mark.go) per the locked spec —
  ghost TR at ≥14px cells, solid --delivered-mid below, --mark-cell
  parameterized CSS; nav home link now mark + stacked lockup with
  per-surface sub (console/inspect/settings). Full Titlebar pattern
  deferred to the Inspector shell (slice 4), where it lives.
  Evidence: full gate green.
- [x] **3. Console shell** — landed. `/` is the 3-column Console
  (360|1fr|340, internal/app/console.go): needs-you rail (open
  threads, capped, victory empty state), hero `packets verified`
  (Done count — forwarded/delivered deliberately NOT claimed yet),
  settled rail, honest dashed empties for calibration + watches;
  center column preserves the tested act-now/state sections; open-
  thread count folded into the existing SSE poll signature.
  Follow-up noted: settled-row "missed" renders amber (--held) —
  revisit the color semantics when hold states become real (slice
  10). Evidence: full gate green.
- [x] **4. Inspector shell** — landed. /review renders the 3-column
  Inspector (252|1fr|312, internal/app/inspector.go): identity
  strip (wo#/key, short base→fix rev chip omitted when unknown,
  repo folder name), file tree left (honest empty when unscoped),
  Monaco + answer form center (islands untouched), annotation rail
  right (threads as agent/question annotation cards keeping
  review-thread classes + data anchors; adjustment ✎ zone below),
  honest timeline footer. Zero pre-existing test edits needed.
  Route rename + owner/repo addr deferred (slices 15 / 5).
  Evidence: full gate green.
- [x] **5. Packet aggregate** — landed. internal/packet: Addr
  (owner/repo from git origin, ssh+https parsing, honest
  local/<dir> fallback), Lifecycle+HoldKind state machine, Packet
  aggregate w/ prompt-slug names, pure Fold(views, addr,
  openQuestions) — mapping: queued→composing, running→in-flight,
  done+caught+0q→verified, done+(!caught|q>0)→held advisory (one-
  clause reasons), failed→held blocking, unknown→held blocking
  (fail-toward-attention); Delivered pinned unreachable until ACK
  (slice 13). 93 subtests. Evidence: full gate green.
- [x] **6. Wire surfaces to packets** — landed. Console reads the
  fold: needs-you = held packets (blocking-first, reasons, pulse),
  in-flight strip (pulsing + ghost composing cells), hero counts
  ONLY State==Verified, settled rail lifecycle-colored, mono addr
  line; Inspector titlebar shows addr + packet name. Meters
  (stock/balance/bandwidth/dispatch rows) deleted from UI + their
  surface helpers; caught count folded into the SSE poll signature
  (Caught-flip-only transitions fan out, pinned). Poll stays one
  ledger projection per tick; addr cached per session (sync.Once).
  Fund/bench/land controls untouched (slice 11). Evidence: full
  gate green.
- [x] **7. Lanes from blast radius** — landed. internal/packet: pure
  core (`Lane`, `ImportGraph`, `BlastRadius`, `LaneFor` — ≤10%
  best-effort, ≤40% standard, else strict, irreversible unreachable);
  exec seam `LoadImportGraph`/`ChangedPackages`/`Measure`
  (lane_measure.go) over real `go list -json` + `internal/diff`,
  erroring to `LaneUnmeasured` — never guessed. `Packet.Lane` field
  added (zero value; Fold stays pure, never computes it).
  internal/app: per-session `liveEntry.laneCache` (`laneFor` computes
  + caches on render, skips caching a rev-less packet; `cachedLane` is
  a pure map read for the poll). Inspector titlebar's order-scoped
  branch computes the lane chip on render (`lane <name>`, neutral
  pill, never a state color); Console's lane-health grid reads ONLY
  the cache (4 honest-zero buckets: best-effort/standard/strict/
  unmeasured). Proven by test that the 100ms OnConnect poll never
  populates the lane cache for an order never opened in the Inspector.
  Found + fixed along the way: a pre-existing latent fragility in
  `TestLiveServer_streamsAVerdictFromInFlightToCaughtAndLogsIt`
  (two sequential `vt.AwaitFrame` calls assumed "catch" and
  "land-clean" always land in different SSE read chunks — any Console
  content growth, including this slice's grid, could shift both into
  the SAME chunk, and the first call would consume-and-discard it on
  matching "catch" before the second call ever saw "land-clean");
  merged into one `AwaitFrame` call with both needles, immune to
  chunk boundaries. Evidence: full gate green (`./scripts/test-fast.sh`).
- [x] **8. Gauntlet record** — landed. internal/packet/Gauntlet: six
  named gates (intent fidelity, handshake conformance, handshake
  tightness, build·vet·lint, test sensitivity, independent check),
  GateStatus not-run/passed/failed/held, Forwardable() (only a hard
  Failed blocks). REAL data wired for two: G3 (handshake tightness)
  reconstructed from the order's existing catch.Outcome — never
  re-runs mutation; G4 (build·vet·lint) is a genuinely NEW exec
  seam, RunBuildVetGate — a throwaway git worktree of the fix rev,
  `go build`+`go vet`, always cleaned up. G1 (intent fidelity) is an
  honest human residual: a real ConfirmIntentFidelity action in the
  Inspector, no computed proxy. G2/G5 (undifferentiated
  handshake-vs-agent tests) and G6 (cage never wired to local
  dispatch) stay explicit "not measured" — deferred to slices 9 and
  a future cage-wiring slice, never faked. Cache/compute boundary
  mirrors slice 7's lanes exactly (compute+cache on render, poll
  reads cache only, proven by test). Inspector timeline now lists
  all six gates (previously a stub). Console gate-health summary
  deferred (out of scope this slice). Evidence: full gate green.
- [x] **9. Handshake mechanics** — landed. Protected `handshake/**`
  path-deny in settle (scanStagedPaths, blocks add/modify/delete —
  distinct from the added-line secret scan, both checked, either
  blocks); packet.Handshake (WriteHandshake/VerifyHandshake, hash
  via reanchor.HashLines, self-declared Strength
  examples/properties — never scored); ledger.Target +
  HandshakePath/Hash/Strength (wire-safe, omitempty). G2 becomes
  REAL: RunHandshakeGate runs `go test ./handshake/...` in a
  throwaway worktree of the fix rev (mirrors G4's pattern exactly);
  VerifyHandshake mismatch force-fails the gate regardless of the
  test result (integrity over a stale pass). Compose flow requires
  AuthorHandshake before PlaceOrder for live (prompt-bearing)
  orders; legacy pre-funded orders unaffected. DEFERRED (explicit,
  next slice or later): splitting mutation into a handshake-scoped
  run (G3) vs an agent-test-scoped run (G5) — both remain exactly
  as slice 8 left them. Evidence: full gate green.
- [x] **10. Hold/forward** — landed. internal/packet/hold.go:
  LaneFloor (lane→minimum required HandshakeStrength) +
  ReconcileHold — a pure, escalate-only composition (called by the
  app after Lane+Gauntlet are cache-attached, never inside Fold):
  a handshake below its lane's floor forces HoldBlocking with the
  exact design-voice phrase "handshake below lane floor"; a hard
  gate failure forces HoldBlocking naming that gate's own Detail
  (reused, never reinvented); otherwise Fold's lifecycle-hold
  baseline is untouched. Fixed a real gap along the way:
  `Target.HandshakeStrength` was never copied onto `Packet` — now
  is. Wired into Console's needs-you/settled rails via cache-only
  reads (poll-never-computes invariant intact). DEFERRED (explicit,
  matches MVP.md's own sequencing): advisory-hold SAMPLING/
  calibration draws are slice 11; "guardrail" is not modeled as a
  separate concept from gate-failure — gate-failure IS the
  guardrail trip for MVP purposes. Evidence: full gate green.
- [x] **11. Attention economics** — landed. Reused the existing
  block/unblock timestamps (real interrupt ground truth, no new
  schema) via `ledger.InterruptsSince` — Console KPI renders the
  locked `N/10 interrupts` format, live via the SSE poll. Calibration
  draw: a pure, stable (non-flickering) random pick over Verified
  packets, replacing the dashed placeholder when one exists. Added
  the concepts.md dry-aside line ("an empty queue is success, not
  idleness") to the victory empty state. Console-only; /board
  /settings /inspect untouched. Noted, not fixed: the SSE poll's
  composite-int signature assumes each folded count stays <1000 —
  pre-existing pattern, extended one digit deeper, immaterial at
  MVP session scale. Evidence: full gate green.
- [x] **12. Watches + precision** — landed (built directly by the
  coordinator, per the 2026-07-04 model-discipline revision — no
  Sonnet builder subagent). internal/packet/watch.go: 3 canonical
  pre-defined triggers (strict-lane, gate-failure, blocking-hold —
  not an author-your-own DSL), EvaluateWatch (fail-closed on an
  unrecognized kind), Precision (marked-fires-only sample, "no
  history yet" when empty), IsNoisy (sampled≥5 && score<0.5). App
  wiring: fires recorded once per (kind,packet) idempotently on the
  same render pass as reconcileHolds, a real MarkWatchFire human
  action (mirrors ConfirmIntentFidelity), Console's "your watches"
  rail replaced with real precision + an inline mark prompt
  suppressed once a watch turns noisy. Full RYGBA (Explore Yellow +
  general-purpose Blue) on both the pure unit and the app wiring;
  Blue caught only a doc-comment overclaim. Evidence: full gate
  green.
- [x] **13. ACK/deliver** — landed (built directly by the
  coordinator). `internal/packet`: lifecycleFor gains "deployed"→
  Delivered/no-hold and "regressed"→Held/blocking/"deployment
  regression"; Deliverable() is now real (State==Delivered);
  ReconcileHold exempts Delivered from lane-floor/gate-failure
  escalation (a real gap the Blue audit caught: those rules were
  state-agnostic and could have shown a delivered packet as
  blocking-held). Console's settled rail + per-packet state cell
  render "delivered" (--delivered fill) only on a real ACK — never
  counted as "verified" in the hero stat. New `packets deployed`/
  `packets regressed` CLI subcommand (cmd/packets): an optional
  --check command's exit code must AGREE with the asserted verb or
  the command refuses (never appends a status that contradicts what
  it just observed); reopens the same durable ledger a running
  server uses. Manually smoke-tested end to end (bare assert,
  confirming check, contradicting check → refusal). Evidence: full
  gate green.
- [x] **14. Adversarial probes** — landed (built directly by the
  coordinator). internal/packet/probe.go: RunAdversarialProbe
  materializes its own throwaway git repo with a deliberate syntax
  error and runs it through the real RunBuildVetGate (G4) — no
  repoDir/ledger parameter anywhere in the call chain, so it cannot
  touch any real session's economy by construction (simpler and
  safer than routing through a scratch ledger session). ProbeReport
  distinguishes "caught" from "ESCAPED". New `packets probe` CLI
  subcommand prints the report and exits non-zero on an escape,
  doubling as a gate-health check. Manually smoke-tested (real run:
  caught, exit 0). Evidence: full gate green.
- [x] **15. Vocabulary sweep + retirement** — landed (built directly
  by the coordinator). Renamed ~35 rendered-copy violations found by
  an exhaustive scan: land/PR/merge → forward/held (land_action.go,
  land.go, board.go); draft → intent, Place order → Compose packet
  (authoring.go); spend/balance/stock/order → compose/catches/
  target (onboarding.go, fund_work.go, supply.go); session/board →
  addr/fleet (board.go, live.go, nav.go); bench → "queued targets:";
  oracle → gate, including surface/card.go's verdict headlines used
  by BOTH Console and the fleet board. New
  `TestSurfaces_neverRenderRetiredVocabulary` +
  `..._perPacketInspectorNeverRendersRetiredVocabulary`
  (internal/app/vocabulary_internal_test.go): a real regex-based
  regression guard exercising every packet lifecycle state across
  `/`, `/board`, `/review` (both branches), `/settings` — it caught
  one real violation on first run (`/settings`'s "live orders"). A
  Blue audit of the guard itself then found a genuine coverage gap
  (the land-lifecycle badges need a real Approve action, which the
  fixture can't drive) and, through that gap, a live "merge" bug in
  `internal/surface/land.go` no test had ever exercised — both
  fixed. Routes/flags checked: none dead, nothing to retire.
  Evidence: full gate green.
- [x] **16. Final gauntlet sweep** — landed. Adversarial review of
  slices 1–15 (app-layer + gauntlet-layer) surfaced four real gaps,
  all fixed: (1) a stale-cached Lane/Gauntlet could force a
  Verified packet's State to Held without flipping State itself —
  `ReconcileHold` now forces State=Held whenever either escalation
  rule fires, closing a settled-rail/needs-you contradiction. (2) a
  timed-out `go build`/`go vet`/handshake-test exec was reported as
  a fabricated GateFailed instead of an honest GateNotRun — added
  `packet.GateForExecError` (extract-first, tested via a pre-
  cancelled context, no real-timing flakiness) and wired it into
  both `RunBuildVetGate` and `RunHandshakeGate`, including their
  worktree-checkout and handshake-test-output paths. (3) `Approve`
  had no concurrency guard, unlike every sibling mutating flow — two
  concurrent lands could race the shared repo's push/PR ops and
  corrupt the cached result; added `beginLand`/`endLand` mirroring
  `beginAnswer`/`endAnswer` (same `findingsMu`), proven by a test
  that holds one `Approve` in flight and confirms a concurrent
  second is dropped while a later one still runs once the slot
  frees. (4) `orderTarget` capped dispatch reads at 50 while
  `packetForOrder` folded ALL of them — a session's oldest packet
  became a named-but-uninspectable ghost past 50 later orders;
  unified both to read unbounded. Doc freshness: MVP.md's Console
  layout/adversarial-probe/vocabulary-map fixed to match what
  actually shipped; wrote README.md (previously absent) documenting
  the built MVP, its concepts, CLI surface, and architecture; added
  `scripts/demo.sh`, a real runnable walkthrough building a scratch
  repo and driving compose → forward → hold → inspect → deliver
  (the fully scriptable gate/adversarial mechanics run automatically
  via `verify-catch`/`probe`; the interactive Console/Inspector
  steps are a printed browser walkthrough, since composing/
  inspecting is a genuine UI action). Evidence: full gate green.
- [x] **17. Learning/convergence** — landed (maintainer directive,
  post-MVP; scoped via an Opus consult after the user asked for
  guidance on signal/process/UI/scope-unit tradeoffs). New MVP.md
  concept 11: a repo new to Packets is "learning" until it
  accumulates real settled history, then "converged" — never a
  fabricated verdict. internal/packet/learning.go (pure, mirrors
  watch.go's honest-history shape exactly): `LearningThreshold` (5,
  same minimum-sample bar as IsNoisy), `SettledCount`
  (verified/held/delivered — identical set to the console's existing
  settled rail, cross-checked to never diverge), `Converged`
  (SettledCount >= threshold). Wired into Console's settled rail as
  a new `renderLearningCard`: honest "N/5 settled" running count,
  flips to "converged" once real history clears the bar — no cache,
  computed fresh from the same folded packets every render, matching
  the settled rail's own live-computation pattern. No new
  persistence/repo-identity layer; scoped per session like every
  other cache (laneCache, watch fires). Full RYGBA (Explore Yellow +
  general-purpose Blue) on both the pure unit and the app wiring;
  Yellow caught a real test gap (a naive `len(packets)>=5` could
  have slipped past the first draft of the convergence test) before
  Green landed. Evidence: full gate green.
