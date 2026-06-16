# CURR-ROADMAP — shortest path to the end-to-end review flow

> **Goal flow.** Start a session on a GitHub repo → coauthor a work-order in
> realtime with a real Claude Code harness in a producer → watch it get filled
> → inspect the work → leave adjustments → watch the adjustments get addressed
> → see tests pass → open a PR.
>
> This doc is the *near-term* roadmap that turns the built spine into that
> demoable flow. The full design is `DESIGN.md`; the round-by-round build log
> is `DESIGN.md` §0.2. Risk register: `RISKS.md`. Vision/economy: `VISION.md`.

## Where we are (grounded against the wired surface)

The goal flow maps onto two loops in the design. Only one is built today:

- **The catch-economy loop** (`DESIGN.md` §0.2, `internal/harness` + `internal/ledger`)
  — live order → real `claude` harness fills it → oracle catch → answer-with-a-test
  to kill a surviving mutant. **Built and wired end-to-end.**
- **The review-thread loop** (`DESIGN.md` §4, §11 P2/P3, §12.3, §14, §28) —
  inline comment on a line → fed to the harness as a turn → reply + fixup
  revision → resolve/outdated. **This is the flow the goal describes, and it is
  largely unbuilt.**

The build deliberately front-loaded the novel integrity spine (`DESIGN.md` §10:
"the build order de-risks the pipe and the integrity, not the review UX"). The
review-thread UX was sequenced last.

### Step-by-step status

| Step | Status | Evidence |
|------|--------|----------|
| (a) start session on a **gh repo** | **WIRED** | `-repo <url>` clones on boot (`app.CloneOnBoot`, `cmd/packets/main.go`); board create clones a URL pick (`resolveOrCloneRepo`, `internal/app/board.go`, `internal/app/repo_clone.go`) |
| (b) coauthor work-order realtime w/ real harness | **WIRED** | Monaco authoring (`internal/app/authoring.go`) → `PlaceOrder` → `runLiveOrder` → `harness.RunProcess`/`RunContainer` (`internal/app/live.go`) |
| (c) see it filled | **WIRED** | live activity beats over SSE to the card (`internal/app/live.go`, `formatActivity`) |
| (d) inspect work | **WIRED** | Monaco base→fix diff island + cached verdict (`internal/app/review_surface.go`, `internal/app/live.go`) |
| (e) do adjustments | **WIRED** | anchored comment → `assist.ReviewTurnPrompt` (§12.3) → `AddAdjustment` dispatches a harness turn (`internal/app/review_adjust.go`); entry-point form on the review surface |
| (f) see adjustments addressed | **WIRED** | the adjustment reuses the live-order pipe — the agent re-edits and the fix settles a new revision (`AddAdjustment` → `drainQueuedOrders` → `runLiveOrder`) |
| (g) see tests pass | **WIRED** | catch cycle runs `go test ./...`, verdict resolves on the card (`internal/app/live.go`, `pipe.RunCatchCycle`) |
| (h) open a PR | **WIRED** | `Approve` guards (`landBlocked`) → opens a PR via the `openPR` seam (push + `gh pr create`), surfaced in a land control (`internal/app/land.go`, `internal/app/land_action.go`) |

**Net (branch `roadmap-auto`):** all of (a)–(h) wired. The three additive slices
below are built. Remaining is *refinement*, not the spine: the deferred
§12.4/§14/§28 thread/outdated/re-anchor machinery, the §29.2 merge-queue
lifecycle, and the live `openPR`/clone subprocess paths exercised against a real
remote (the orchestration is unit-tested via swappable seams; the git/gh/clone
I/O is verified by build + manual run).

*Done since (autonomous refinement loop, branch `roadmap-auto`):*
- **real squash-to-one on land** — `squashToCommit` (`internal/app/land_action.go`)
  collapses `BaseRev`..HEAD into one detached commit via `git commit-tree` (no ref
  move, no working-tree touch) and `realOpenPR` pushes that single SHA, so the
  opened PR is one clean squashed commit instead of every session revision.
- **pre-push secret-scan gate** — `settle.ScanCommitRange(base, rev)` scans only what
  the pushed range adds (two-dot diff); `realOpenPR` runs it between squash and push
  and refuses rather than leak, closing the blind spot that the squashed commit was
  built from HEAD's tree and never re-scanned (RISKS.md secret-leak is CRITICAL).

- **legible land-result** — `classifyLandResult` (`internal/app/land_action.go`) maps the
  cached outcome to opened/blocked/error; `renderLandControl` renders an opened PR as a
  clickable `target=_blank` link in confirmed-green, a guard block as a calm dim notice,
  a failure in the miss hue — so the finish line reads at a glance instead of one mono blob.

- **adjustment re-anchor (thin §28 slice)** — `reanchorAdjustment` (`internal/app/review_adjust.go`)
  relocates a left adjustment's commented line against the settled revision by exact
  content match (same / moved / outdated); the last anchor is cached on the liveEntry
  (`setAdjAnchor`) and the review surface renders a badge ("still on line N" / "addressed —
  moved to line M" / "addressed — line edited") — so "leave an adjustment → watch it
  addressed" has a visible payoff. Exact-match only; git-hunk rebase + rename tracking
  stay deferred.

- **leased land push (`--force-with-lease` correctness)** — `pushRefspec`/`pushSquash`
  (`internal/app/land_action.go`) replace the vacuous bare `--force-with-lease` (which on a
  no-tracking-ref session branch degraded to an unguarded clobber) with an explicit lease:
  `--force-with-lease=refs/heads/<branch>:<expected>` — empty on first land (must-not-exist),
  the cached last-pushed SHA (`liveEntry.lastPushedSHA`) on a re-land. A hermetic bare-repo
  integration test verifies git's semantics (empty=must-not-exist, stale lease bites, legit
  re-land succeeds). Approve threads + caches the SHA; fails CLOSED on partial failure.

- **re-land updates the open PR** — `isPRAlreadyExists`/`ghPRViewURL` (`internal/app/land_action.go`):
  on a re-land the leased push already updated the open PR but `gh pr create` fails
  "already exists"; realOpenPR now recognizes that benign signal (both "pull request" &
  "already exists", case-insensitive — not over-matching) and surfaces the existing PR's
  URL via `gh pr view` instead of a spurious failure.

- **catch economy: no phantom catch from an incomplete oracle** — `catch.LineState` gains an
  `Undetermined` set; `LineStateAt` records timed-out (never-killed) mutants instead of
  silently dropping them; `Detect` fails CLOSED (`NoOracleSignal`) when the after-run is
  incomplete, so a fix that merely makes the suite hang can no longer mint a phantom Catch
  (a "mechanical-fact-as-guarantee" hole in the novel economy core, RISKS.md dominant
  family). `pipe.CatchAcross` attributes the quiet verdict to a new `ReasonOracleIncomplete`
  vs `ReasonNoMutableOperator`. Backward-compatible JSON (old transcripts → nil → old behavior).

- **catch economy: symmetric fail-closed on incomplete runs** — `Detect`'s guard now fires
  on `before.Undetermined` OR `after.Undetermined` (was after-only): a catch is minted only
  from two COMPLETE oracle runs, so a timed-out before-mutant (which understates the
  baseline survivor-set) can no longer mint a spurious catch. Intentional contract change
  (the old before-ignored test was rewritten). `CatchAcross` attributes either-side
  incompleteness to `ReasonOracleIncomplete`. Safe: ledger mints only on `Catch`, so
  flipped cases simply don't record (no orphan/balance risk; verified across cage+ledger).

- **reanchor: no wrong-line Moved on ambiguous content** — the `Moved` branch
  (`internal/reanchor/reanchor.go`) validated the relocated anchor with a single positional
  hash check, so a duplicated line (`}`, `return nil`, blank) could mint `Moved` onto a
  coincidentally-identical line — and the economy then confirms a catch on the WRONG line.
  Now fails closed: `rangeHashMatchCount` requires the anchor's content to match EXACTLY ONE
  window; >1 → `Outdated`. Hardens the trust anchor the catch economy keys off (the layer
  below `Detect`). Holds for multi-line blocks too. Fail-closed only reduces catches (safe).

*Investigated & dropped:*
- ~~orchestrator `registerSession` double-drain~~ — **not a reachable bug.** `CreateSession`
  (board.go:228) already guards re-registration (`liveReg.Load(key)` → no-op) and is the only
  multi-call path; boot seeds `default` once. The proposed fixes are net-negative here:
  `e.cfg`/`e.log` are read UNLOCKED in hot paths, so in-place mutation would *introduce* races;
  and replace-semantics are load-bearing for the test harness (re-seeding `default` across
  `NewServer`). Left as-is.

- **multi-hit secret-refusal message** — `formatSecretRefusal` (`internal/app/land_action.go`)
  names EVERY detected secret (scan order, singular/plural count) instead of just the first,
  so a Lead fixes them in one pass; `.land-control__result` gets `white-space: pre-wrap` so the
  per-secret lines stack legibly.

- **surface honesty: incomplete-oracle no longer lies** — `ReasonOracleIncomplete` (added in the
  catch-economy fix) had no surface case, so `PresentVerdict` fell through and the card rendered
  the FALSE "This line has no mutable operator" for a line that *has* operators (the run just
  timed out). Added a distinct `OracleIncomplete` verdict + honest render ("Oracle incomplete —
  the line has mutable operators, but a mutation run did not finish"). A self-introduced
  regression (the tick-7 reason wasn't wired through the presentation layer); verified end-to-end.

- **secret gate: `--text` defeats `.gitattributes` binary-coercion** — a `*.env -diff` attribute
  made `git diff` render a PLAINTEXT secret file as "Binary files differ" (zero `+` lines), so the
  added-line scanner missed it and the secret entered history — defeating the §12.2.1 CRITICAL
  invariant, and the same hole sat on `ScanCommitRange` (the pre-push land gate). `--text` is now
  pinned on all three scan diffs (`Settle`, `ScanHistory`, `ScanCommitRange`) forcing a textual diff;
  artifact surfacing (`--numstat`) is unaffected so genuine binaries still surface. Verified no
  false-positive on high-entropy binary (high-confidence structured rules).

- **economy: bandwidth meter floors at zero** — `Projection.Bandwidth()` returned `total - bwSpent`
  unfloored, so a forged/over-published `bwspend` (past the `AppendBandwidthSpend` overdraft gate,
  the same forge path the balance test exercises) drove the meter negative — a corrupt projection
  the `< cost` gates and board would misread. Floored at zero (strict-safe: only lowers the reported
  value, never lets an overdraft through). Mirrors the project's "projection holds against forged
  stream data" invariant.

- **economy: balance floors at zero too** — `Projection.Balance()` now floors at zero (twin of the
  bandwidth floor), so a forged positive over-spend published past the `AppendSpend` gate can't drive
  the balance negative. Both economy projections now hold against forged stream data in both
  directions (negative-spend can't mint credit; positive-over-spend can't go negative). Strict-safe:
  only lowers the reported value, gate stays correct, no raw-field reader bypasses it.

- **harness: errored turn keeps prior revision** — `Supervisor.Run` settled on EVERY `turn.ended`,
  so a crashed/out-of-turns agent (`result subtype:"error_max_turns"`/`error_during_execution`)
  minted its PARTIAL working tree as a real revision and threaded fabricated work into the catch
  cycle — violating DESIGN §1183 ("error subtype → keep prior revision"). Added `translate.TurnErrored`
  (fail-closed: only "success" mints) + a reducer guard that records an errored turn as non-minted
  (base doesn't advance, `lastMintedSHA` skips it).

- **harness: errored turn discards its partial edits** — completing the errored-turn fix:
  `orchestrator.DiscardWorkingTree` (`git reset --hard rev` + `git clean -fd`) rolls the working tree
  back to `baseRev` on an errored turn, so the residue can't be swept into a LATER turn's whole-tree
  `git add -A` mint (§1183 "partial edits discarded"). Wired into the harness errored branch (fails
  the run if rollback fails); the container path inherits it via the same reducer. Verified by an
  errored-then-successful end-to-end leak test.

- **cage verifier well-formedness gate** — `DeriveCatch` (`internal/cage/derive.go`) now refuses a
  transcript whose Before/After `Survivors`/`Undetermined` contain an operator absent from that
  side's `Inventory` (`escapesInventory`), BEFORE `catch.Detect` — so a malformed/forged transcript
  from an untrusted cage (a phantom out-of-alphabet operator that flips the verdict) is refused
  wholesale rather than trusted. Enforces the `Survivors ⊆ Inventory` invariant catch.go only
  declared. Correct asymmetry: the host's trusted pipe path builds `LineState` via `catch.LineStateAt`
  (well-formed by construction), so only the untrusted cage path needed the gate.

- **claims: producer can't forge a verdict** — the cross-process producer grant allowed publishing
  its whole `claim.>` subtree, including the host-reserved `verdict` kind — so a producer could
  publish a forged `ClaimVerdict{Rejected:true}` to self-resolve its own bets (drop from in-flight,
  count a verified-loss), corrupting the fleet board's loss half WITHOUT the cage verifier running.
  Added `fabric.ClaimVerdictKind` + a `Deny` on that one kind in the producer grant (NATS Deny
  overrides Allow); `ledger.subjectKindVerdict` now references the shared constant (no drift). The
  host verifier's verdict publish is unaffected; producers keep every other claim kind. Verified the
  fabric authz is the complete chokepoint (in-process = host; no non-host path to inject a verdict).

- **supply: a split can't re-open an already-funded parent** — `fundableBacklog` (`internal/app/supply.go`)
  expanded a split's sub-targets without checking the parent's consumed state, so a split refinement
  on an already-funded parent re-opened its sub-regions (a latent double-draw, gated only at the
  action layer). Now `ok && !consumed[t]`: a split replaces a parent only while it's still fundable;
  once bought, neither the parent nor its sub-regions re-draw. Makes the pure fn correct in isolation.

*Investigated & found sound:*
- supply / backlog draw-down — no double-draw/off-by-one/re-draw in the live paths: `fundableBacklog`
  keys `consumed` on the full Target struct (pure projection of persisted orders, debit+order under
  one lock). The one pure-fn gap (split-on-consumed-parent) is now hardened (above).

- **board renders human verdict labels** — the fleet board's "why" tag rendered the raw
  snake_case verdict token (`lost_via_rename`, `oracle_incomplete`); `surface.VerdictLabel`
  (delegates to `present()`, raw fallback for unknown/forward tokens) now gives the same human
  headline the review card shows ("Anchor lost: file renamed"). Completes the surface-legibility theme.

- **land: cache the pushed SHA on a post-push failure (unwedge re-land)** — Approve cached the
  pushed SHA only on full success, so a `gh` failure AFTER a successful push left the cache stale
  while the remote branch advanced — wedging the next re-land's `--force-with-lease` (it leased the
  old SHA against a branch that moved). Approve now caches whenever `pushedSHA != ""` (before the
  err check); `realOpenPR` returns the pushed sha on its two post-push failure paths. Verified the
  lease-loop unwedges (cached SHA matches the remote at every step).

- **security regression guards** — added two characterization tests locking shipped security
  properties against regression: an errored turn's secret-bearing file is scrubbed by
  `DiscardWorkingTree` (`git clean -fd`, so it can't leak into a later mint), and `ScanHistory`'s
  `--text` catches a secret in a `.gitattributes`-coerced "binary" file. Each fails if its guard
  is removed.

*Hardening queue — EXHAUSTED (as of the autonomous loop's final sweep).* Every major subsystem
has been swept and hardened with a shipped, tested fix; a final council micro-sweep of the last
unexplored corners — authoring/analyze (incl. the analysisCancel supersede race), gc/prune,
bundle-guard accounting, tokenstore (no secret-logging), ingest pruning — found them all SOUND
(no race, swallowed-error, leak, miscount, or lying projection). The economy firewall, the
trust spine (catch/reanchor/cage), the secret gates, the harness revision discipline, the
cross-process auth, and the surface honesty layer all hold against forged/incomplete/malformed
input in the directions checked.

*Feature progress (beyond hardening):*
- **review loop remembers MULTIPLE adjustments** — `liveEntry.adjAnchors` (was a single
  overwriting `adjAnchor`) now appends every adjustment; `relocateAdjustments` (pure, injected
  reader) relocates each against its file's current content and `renderAdjustmentStatuses` shows
  one badge per adjustment with the Lead's comment + addressed/moved/outdated status. The goal
  flow is "leave adjustmentS → watch them addressed"; a Lead leaving several no longer sees only
  the last. Reuses the `reanchorAdjustment` core unchanged. (tick-4 council demo-fidelity slice.)

- **review loop: resolve/dismiss an adjustment** — closes (e)/(f) symmetrically with the
  answer-vanish flow: `ReviewCard.ResolveAdjustment` (off-ledger, no harness — just forgets the
  anchor via `removeAnchor`/`removeAdjAnchor`) lets the Lead clear an addressed adjustment, with a
  per-badge "resolve" button wired through the datastar inline-assign bridge (keyed on the original
  anchor file:line, escaped via `jsStr`). The adjustment list is no longer accumulate-only.

- **review loop: re-commenting a line replaces, not stacks** — `upsertAnchor` gives `addAdjAnchor`
  last-writer-per-`file:line` semantics (mirroring `splitRefinements`): re-commenting a line updates
  its single badge with the latest comment+anchor instead of stacking a duplicate, and bounds
  `adjAnchors` to one entry per commented line (closing the prior unbounded-append note). Distinct
  lines still each get their own badge.

*Deferred work — STARTED (council scope-approved, off-ledger only):*
- **§29.2 "Landed ≠ Merged" — first slice** — the land control now surfaces that opening a PR is NOT
  merging it: `classifyLandLifecycle` (gh state → landed/merged/bounced, fail-closed — never a false
  Merged), a swappable `mergeState` gh seam, an ephemeral `liveEntry.landLifecycle` cache (set to
  `landed` on Approve's open, cleared on blocked/error so no stale badge lingers), a `CheckMergeState`
  action (no-op unless a PR was opened; gh error = calm no-op), and a land-control badge ("Landed —
  not yet merged" / "Merged" / "PR closed unmerged") + check button. RENDER/STATE-ONLY: touches NO
  economy stream / ledger event kind (verified) — that constraint is why it was safe to start
  autonomously. The full merge-queue/bounce-retry machinery and any DURABLE lifecycle record (§29.3
  `landing_outcomes` table) stay GATED for the Lead (they touch the authoritative substrate).
- **§29.2 lifecycle on the FLEET BOARD** — the merge outcome now surfaces per board row, not just on
  the open card: pure `boardLifecycle(lc)` (merged→"merged"/bounced→"closed unmerged"/show=true;
  landed/""/unknown→show=false), `CardRow.LandLifecycle` set from `e.landLifecycleSnapshot()` next to
  `row.Land`, and a `board-row__lifecycle` render span shown ONLY for terminal merged/bounced — the
  routine "landed, not yet merged" transient stays off the board so it stays CALM (mirrors how
  `boardLand` only surfaces BLOCKED). RENDER/STATE-ONLY: reads ephemeral `landLifecycle`, no ledger /
  economy-stream change (Blue-verified). `fleetFingerprint` (board.go:181) now folds the DISPLAYED
  lifecycle state (gated on `boardLifecycle` show), so a terminal merged/bounced auto-live-refreshes
  a board tab — while the hidden landed/""/unknown transients do NOT move the fingerprint, preserving
  the idle-no-flood invariant (a `""→landed` change pushes no look-identical frame). Delimiter scheme
  audited collision-free; tests assert both the live-refresh and the idle-calm halves.

*AUTONOMOUS LOOP — CONVERGED & STOPPED (council verdict, code-verified).* A full-design-doc council
sweep verified against actual code (not the self-report) that the safe off-ledger surface is genuinely
swept: goal flow (a)–(h) all wired & green; §29.2 "Landed ≠ Merged" built on BOTH the card and the
fleet board (badge + fingerprint-fold, idle-calm preserved) and reachable end-to-end via the Lead's
"Check merge state" button; the "watch a real worker" surface complete (bounded scrolling
`activityTranscript`, every `liveEntry.*Snapshot()` accessor wired to a render — no dead ephemeral
field left to harvest). The ONE remaining off-ledger imperfection (the fill-buffer Stream signature
at `live.go:1557` keys on the latest beat, not transcript length, so two identical consecutive beats
lag one frame until the next distinct beat) was flagged as marginal — ~2 lines, self-corrects
sub-second, near-zero payoff — i.e. manufacturing busywork; deliberately NOT built. Everything else of
real value is Lead-gated by the council discriminator: §29.3 `landing_outcomes` durable records, the
merge-queue/bounce-retry machinery, `CheckMergeState` on an automatic polling cadence (an
external-I/O-on-a-cadence POLICY decision — today's explicit Lead-clicked button is correct), §14
thread/message relational projection, fan-out, the trust-economy. These are Lead decisions, not ticks.
The loop stopped itself here rather than churn idle re-assessments or manufacture marginal work; re-run
`/loop` to resume once the Lead un-gates work or adds a new session/repo.

*Demo-fidelity feature stream — also EXHAUSTED (council-confirmed).* The review-thread loop is now
at good fidelity (multiple adjustments tracked, each with its comment + addressed/moved/outdated
badge; re-comment replaces; resolve clears). A focused council verdict confirmed the resolve slice
was the LAST genuine tick-sized demo-fidelity item — everything else is cosmetic (e.g. a diff
snippet in the badge — noise, the Monaco diff is one click away) or an explicitly-deferred large
subsystem (§14 thread projection, §29.2 merge-queue), not tick-sized.

Deliberately NOT built (net-negative or out of scope, not gaps):
- `scanStagedDiff` giant-line bound — a size cap before the regex could MISS a secret in a large
  line; capping the secret scanner is the wrong trade. Left uncapped (the cost is bounded ~1.3x).
- `realOpenPR` post-push sha-return test — would need a `gh`-on-PATH stub, against the project's
  "I/O seam verified by build + manual run" convention; the contract is already stub-asserted at
  the `Approve` boundary.
- nested/recursive split expansion — single-level by design; whether multi-level sharpening is a
  real workflow is a product-scope question for the Lead, not a bug.

## The plan — three additive slices

Do **not** build the full `DESIGN.md` §14 thread/message projection or the §28
re-anchor machinery first. Reuse the live-harness pipe that already exists
(`runLiveOrder`) and add three thin slices. TDD per `CLAUDE.md` / `CONVENTIONS.md`
(`tdd-rygba`).

### Slice 1 — Comment→harness round-trip (keystone, the real work)

The one piece that makes it feel like reviewing a teammate. Converts (e) and (f)
from "submit a test" to "tell the agent what to fix, watch it fix."

- Add an anchored-comment entry point: `{file, line, text}` composes the
  `DESIGN.md` §12.3 turn template and dispatches it to the **same**
  `runLiveOrder` against the existing session HEAD, settling a new revision.
- Render adjustments as flat session-attached comments first. Full
  thread/outdated/re-anchor state (`DESIGN.md` §12.4, §28; schema §14) is
  deferred polish — **not** a prerequisite.
- This is an *additive reuse* of `runLiveOrder` (`internal/app/live.go`), not new
  architecture. It is the only architecturally non-trivial piece.

Refs: `DESIGN.md` §4 (PR⇄harness mapping), §12.3 (routing a comment back),
§11 P2.

### Slice 2 — `gh pr create` on approve (small, ~1–2 days)

Closes (h). Reuses the agntpr `x-access-token` push pattern.

- `landMode=pr` action: guard (open threads / red checks, overridable) → squash
  session revisions → push branch with a short-lived token → `gh pr create`.
- Mechanical; no new architecture.

Refs: `DESIGN.md` §16 (approve & land), §9.2 (landMode), §29.2 (merge queue /
Landed non-terminal — can be deferred; v1 can push direct branch + PR).

### Slice 3 — Repo-from-URL on session create (small, ~1–2 days)

Closes (a).

- Extend `-repo` (and the board "create session" path) to accept a URL: clone
  on create, checkout a fresh branch off `base_ref`.

Refs: `DESIGN.md` §15.2 (lifecycle: starting → clone + checkout branch),
`cmd/packets/main.go`, `internal/app/board.go`.

## Sequencing & estimate

1. **Slice 1** — substantial (the keystone). Build first; everything else is
   leverage on it.
2. **Slices 2 & 3** — small, independent, ~1–2 days each; parallelizable.

**Bottom line:** ~one substantial slice + two small ones from a coherent
end-to-end demo of the goal flow — *provided* adjustments render as
comment→harness turns rather than full GitHub-grade outdated-thread machinery.

## Explicitly deferred (not on this roadmap)

- Full thread/message relational projection + outdated/re-anchor (`DESIGN.md`
  §12.4, §14, §28).
- Merge-queue delivery + Landed→Merged lifecycle (`DESIGN.md` §29.2).
- Cross-process external-producer claims for *live* orders (today the claim path
  serves pre-baked backlog targets only; `DESIGN.md` §0.2).
- The rest of the trust economy: catch-weight, risk tiers, trust half-life,
  earned concurrency, Delegation Tiers (`DESIGN.md` §0.2, `VISION.md`).
