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

*Council-queued next slices (not yet built):*
- cage verifier well-formedness gate: `DeriveCatch` trusts cage-reported `LineState` without checking
  `Survivors/Undetermined ⊆ Inventory` (an invariant the code declares but doesn't enforce at the
  untrusted boundary) — a phantom out-of-alphabet operator can flip NoCatch→PartialCatch (a
  positive-looking, non-recordable surface verdict). Refuse malformed evidence. (integrity, medium)
- `scanStagedDiff` giant-line bound: with `--text`, a large genuinely-binary file emits its bytes as
  one huge `+` line — bounded (~1.3x) but uncapped memory/CPU per regex. A size cap / binary-skip
  before regex would bound it. (hardening, low)
- `ScanHistory` coercion test: add the attribute-coercion characterization test when ScanHistory is
  wired into a live gate (currently unwired). (test-debt, low)
- board "why" tag renders RAW verdict tokens (`oracle_incomplete`, `lost_via_rename`) instead of
  human copy (board.go:640) — cosmetic, NOT a lie; a shared token→label map would align board with
  the card headlines. (polish, low)
- cache-after-success recoverability: if push succeeds but `gh` (create OR view) fails, the
  un-cached SHA wedges the next re-land (fails closed, not data loss) — cache the pushed SHA
  as soon as the push succeeds, independent of PR-URL resolution. (low)

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
