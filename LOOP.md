# packets — MVP autonomous loop

The standing protocol for the autonomous loop that ships the Packets MVP.
Every tick, the coordinator re-reads this file and executes exactly one
tick. This file is the contract; MVP.md and ROADMAP.md are the state.

## Goal

Ship the Packets MVP: a working control plane for agent-written code
changes, built from the proven experiments in this repo, wearing the
design system in `design/` — its vocabulary, tokens, and surfaces —
end to end. The maintainer granted full freedom to reshape, rename,
and rewrite the experiments in service of the MVP.

The bar: the MVP FULLY REALIZES every concept in
`design/guidelines/concepts.md` — as honest working mechanics, not
copy. Concept → feature, all of them, against a local repo:

1. **Addr + intent + packet**: point packets at a repo (`owner/repo`
   form); compose a packet from prose intent (prompt-first, no
   fix/file/line); the packet is named, revisioned, and has a
   replayable timeline from compose to delivery.
2. **Handshake**: a runnable contract authored independently of, and
   before, the code — with a strength gradient (examples →
   properties/contracts). The agent's own tests are evidence, never
   the contract.
3. **Lane**: QoS class derived from MEASURED blast radius (dependency
   graph coupling — computed, never self-reported, never vibes):
   best-effort → standard → strict → irreversible. Radius buys two
   things: more gates, and a stronger required handshake.
4. **The gauntlet — all six gates**, each lane running what its radius
   warrants: (1) intent fidelity held for the human; (2) handshake
   conformance at line rate; (3) handshake tightness = mutation vs
   SPEC (the cage/mutation machinery); (4) build · vet · lint;
   (5) test sensitivity = mutation vs the agent's TESTS (hollow-test
   detection); (6) independent check via METHOD diversity (static
   analysis / property checks — never a second LLM).
5. **Forward/hold**: a packet forwards autonomously unless inspection
   must hold it — amber advisory holds (sampled) vs red blocking holds
   (strict lane, guardrail, irreversible).
6. **Inspection — all four modes**: PULL (crack open any packet
   anytime in the Inspector — timeline, rich diff, annotations, live
   terminal — cheap enough to do for fun; looking never slows
   forwarding); PUSH (the needs-you queue, budgeted interrupts);
   STANDING (the gauntlet + watch/capture triggers carrying precision
   scores — noisy triggers lose the right to interrupt); ADVERSARIAL
   (seeded probe packets that must be caught, keeping the gates
   honest).
7. **Attention economics**: a real interrupt budget (n per week,
   counted), calibration draws (a random sample of auto-forwarded
   packets surfaced for skim), and empty-queue-is-success framing.
8. **Delivery + ACK**: delivered means ACK'd healthy (`deployed` /
   `regressed`) — distinct from merged/landed; the mark's lifecycle
   (ghost outline → delivered fill) is the state machine.
9. **Console + Inspector** as the two surfaces (per
   `design/ui_kits/console/`), wearing the design system end to end:
   tokens verbatim, the state grammar as the only state colors, mono
   voice, networking vocabulary everywhere.

Out of scope for the MVP: auth, multi-tenant, Postgres, egress
allowlists, real prod-deploy integration (ACK state transitions may be
driven by an honest local signal or explicit operator command — never
fabricated), and any marketing surface. Do not build them.

## Roles — hard model discipline

- **Coordinator (the main loop, this session) BUILDS.** As of
  2026-07-04 (user directive), the coordinator writes production
  code itself for every slice — no separate Sonnet builder subagent
  tier. The coordinator still orients, decides, sequences, runs the
  gate, updates state docs, commits, and schedules the next tick,
  but now ALSO implements each slice directly via the `tdd-rygba`
  skill (Read/Edit/Write/Bash), following the same test-first
  discipline a dispatched builder would have. The coordinator still
  NEVER explores file-by-file itself — grounding stays delegated.
- **Explorers — Haiku only.** Every read-only recon agent (grounding
  a slice, mapping symbols, checking for prior art) runs with
  `model: haiku` (the Explore agent). No other model explores. Use
  Haiku liberally for quick/simple lookups too, not just broad
  sweeps — keep the coordinator's own context for building.
- **Councils** (design forks): 3–6 short persona agents on `sonnet`,
  debated to convergence. The coordinator writes the converged
  verdict into ROADMAP.md and acts on it. Never wait on the user for
  a design fork; only a truly irreversible/destructive action pauses
  the loop.
- **Fable — extremely rare, last resort only.** `model: fable` (the
  top-tier model) may be dispatched ONLY when the coordinator itself
  has genuinely failed to make progress on the SAME slice after a
  real retry with a narrower/sharper approach (not just a session
  limit interruption — resuming after that is normal, not an
  escalation) — i.e. the complexity has grown out of control: a
  design deadlock a council round didn't resolve, a defect whose
  root cause repeated attempts couldn't isolate, or a cross-cutting
  refactor that keeps coming out subtly wrong. Fable is the
  exception, not a tier to reach for out of convenience — the user
  was explicit that it is expensive and must stay rare. Log why in
  ROADMAP.md whenever it's used.
- Any slice/agent already dispatched to a Sonnet builder before this
  directive lands should still be let to finish and reviewed
  normally — this rule governs slices from here forward.

## State files (the loop's memory)

- `MVP.md` — the MVP spec: the binding vocabulary map (old experiment
  name → design-system name), the target surfaces, the kept
  architecture, the distilled brand pack for builders. Written at
  bootstrap, amended only by council verdict.
- `ROADMAP.md` — the slice ledger: ordered slices with status
  (queued / in-flight / landed / dropped), one line of evidence per
  landed slice (commit SHA + what proves it), council verdicts, and
  a NEXT pointer. The single source of "what happens this tick".

Keep both fresh every tick. Doc drift is a defect: if a landed slice
makes a line in MVP.md, ROADMAP.md, or README.md false, fix it in the
same tick.

## Bootstrap (tick 0 — only if MVP.md does not exist)

1. Fan out Haiku explorers in parallel:
   - inventory `internal/` + `cmd/`: per package, what is proven
     (tests), what it does, what depends on it;
   - map current surfaces (`/`, `/board`, `/review`, `/settings`,
     `/stream`) against `design/ui_kits/console/` (Console + Inspector)
     and `design/components/`;
   - extract the design contract: `design/readme.md`,
     `design/guidelines/concepts.md`, `design/HANDOFF.md`, tokens.
2. Convene the naming council (sonnet): settle the vocabulary map —
   the experiments say order/board/review/session/catch/bandwidth; the
   design system says packet/console/inspector/addr/gauntlet/interrupt
   budget. `design/guidelines/concepts.md` is binding; map every
   user-facing word, and decide per internal package whether to rename
   now, rename when touched, or keep (internal names may lag, surfaces
   may not).
3. Write MVP.md (spec + vocabulary map + brand pack) and ROADMAP.md
   (sliced plan). Slices must be thin, independently testable, and
   ordered so the app is demoable after every single one. Sequence
   roughly: tokens/CSS port → Console skin → Inspector skin →
   vocabulary sweep of surfaces → compose flow polish → gauntlet
   surfacing → hold/forward mechanics → ACK/deliver states → sweep.
4. Commit the two files. Schedule the next tick.

## Every tick after bootstrap

1. **Orient** (cheap): read ROADMAP.md NEXT, `git log --oneline -5`,
   and run the gate. If the gate is red, fixing it IS the tick.
2. **Ground**: send Haiku explorers for the exact grounding the slice
   needs — files, symbols, existing tests, the design reference for
   the surface being touched. Never ground by reading files yourself;
   collect conclusions, not dumps.
3. **Fork check**: if the slice hides a real design decision, council
   it to convergence first (sonnet personas, grounded in the actual
   code + `design/readme.md`), record the verdict in ROADMAP.md.
4. **Build**: the coordinator implements the slice directly (via the
   `tdd-rygba` skill), following the build rules below. Strict
   test-first: failing test → confirm the failure message → implement
   → green → refactor with the suite green. No production code before
   its failing test. If genuinely stuck after a real retry (not a
   session-limit interruption), escalate to Fable per the Roles
   section — extremely rare.
5. **Review**: before landing, re-read the full diff against the
   design non-negotiables, CONVENTIONS.md, and the firewall rules;
   fix anything that violates them. Then run the gate.
6. **Land**: update ROADMAP.md (slice → landed, evidence line) and any
   drifted doc; commit with a plain message (`Slice N: <what>`), no
   attribution/Co-Authored-By trailers; push to the loop branch
   `roadmap-auto` (the maintainer merges to main on their own terms).
7. **Schedule**: ScheduleWakeup with this same loop prompt. Active
   building: 60s. Waiting only on CI or long agents: ~270s. MVP
   checklist complete: do a final sweep tick, then stop scheduling and
   report.

## The gate (must be green to land)

```
./scripts/test-fast.sh
```

Runs build+vet, then the ~20 non-container packages at full package
parallelism concurrently with `internal/cage`+`internal/sandbox`
under `-p 1` (those two — and only those two — must stay serialized
against each other: both launch docker containers labeled
`io.packets.sandbox=1` and assert on that shared label). Wall time is
`max(fast, cage)`, not the sum — ~2min locally, vs. the old blanket
`-p 1 ./...` gate. `TMPDIR` defaults to `/home/jgonc/tmp` (off tmpfs —
cage tests fill `/tmp`; override with `TMPDIR_OVERRIDE`). CI mirrors
this split as two concurrent jobs (`test-fast` / `test-cage` in
`.github/workflows/ci.yml`). Check the previous tick's run each orient
step (`gh run list -L 1`). A cage-image pull flake is re-run, not
debugged.

## Build rules (binding on the coordinator for every slice)

- CONVENTIONS.md is binding: test names as behavioral claims,
  test-first, outside-in through the public API, real > stub > mock,
  table-driven, `t.Parallel()`, testify, comments explain why.
- Doc comments in implementation and test code must not cite a
  ROADMAP slice number or name another test function — both go
  stale (slices get renumbered/renamed, tests get renamed/deleted).
  Explain the WHY in terms of the code's own behavior/invariants
  instead (user directive, 2026-07-05).
- Stack is settled — do not relitigate: server-rendered Go via
  `go-via/via` `h.*` builders + Datastar SSE; ONE hand-rolled CSS
  system; NO Tailwind/React/shadcn/WebSockets; Monaco stays a JS
  island; JetStream (`internal/fabric`) is the event source of truth.
- Design non-negotiables: port `design/tokens/*.css` values verbatim;
  the state grammar is the only state color set (composing outline,
  in-flight cyan #4cc4d4, verified green #46c08a, held amber #e6b23e,
  blocking red #f0666b, agent purple #a78bfa, delivered #2a7683) —
  never invent a state color; IBM Plex Mono for everything
  operational, sans only for prose; unicode glyphs, never icon libs or
  SVG icons; the 2×2 mark is built in code (PacketMark spec in
  design/readme.md), never an asset; lowercase operational copy, no
  exclamation marks, no emoji; networking vocabulary always — forward
  / hold / inspect / deliver / ACK / addr / lane / packet — never PR /
  merge / approve / review-as-a-noun.
- Firewall (never weaken): only host settle mints revisions — the
  harness and any agent output mint NOTHING; a live packet's catch
  anchor is pre-specified, never derived from the agent's own diff;
  no fabricated metrics (no invented ranks, scores, or cross-session
  numbers — render only what the ledger proves).
- Tests must not need an API key, network egress, or a running Docker
  daemon unless the existing suite already gates that path (cage tests
  skip cleanly when the image is absent).

## Standing orders

- Never wait on the user. Councils decide forks. Pause only for
  irreversible/destructive actions or a hard external blocker (and
  say so in the final message instead of scheduling).
- Quality over motion: if no queued slice clears "makes the MVP
  demonstrably better", run an adversarial review sweep of recent
  slices instead — and if that finds nothing, tighten ROADMAP.md and
  lengthen cadence rather than manufacture work.
- Keep every tick's final message a two-line status: what landed
  (with SHA), what's next.
