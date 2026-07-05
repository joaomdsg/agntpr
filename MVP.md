# packets — MVP specification

The binding spec for the MVP loop (process: LOOP.md; ledger:
ROADMAP.md). Written by the bootstrap council 2026-07-04; amended only
by council verdict.

Packets is a control plane for agent-written code changes. Agents
compose, verify, and forward changes ("packets") at line rate against
machine-checkable contracts ("handshakes"); the few with real blast
radius are held for a human. Networking automated forwarding but never
automated away inspection — Packets does the same for code.
`design/guidelines/concepts.md` is the binding conceptual model; this
file maps it onto the codebase.

## Concept → feature (the MVP checklist)

Every item ships as an honest working mechanic on a local repo:

1. **Addr** — the repo in `owner/repo` form. New: parse/bind an addr
   identity for each configured repo (today: bare RepoDir paths).
2. **Intent + packet** — a packet is one agent-written change: named,
   revisioned, addressed, with a replayable timeline. Composed from
   prose intent (`Target.Prompt` exists). One packet = one of today's
   work orders (identity 1:1 — the systems council froze this).
3. **Handshake** — a runnable contract authored independently of, and
   before, the agent's code. Concretely: protected paths (e.g.
   `handshake/**` test files) the agent's turn CANNOT touch —
   enforced in `settle` as a deny-rule beside the secret scrub, and
   content-hash-checked before gates run. Strength gradient recorded
   (examples → properties/contracts).
4. **Lane** — best-effort → standard → strict → irreversible, derived
   from MEASURED blast radius: a host-side pure function over the
   `go list` import graph (reverse-dependency weight of changed
   packages). Never self-reported, never vibes. Radius buys more
   gates and a stronger required handshake.
5. **The gauntlet — six gates**, orchestrated as one explicit
   pipeline record per packet, each lane running what its radius
   warrants:
   - G1 intent fidelity — the human residual (Inspector affordance).
   - G2 handshake conformance — run the handshake, line rate.
   - G3 handshake tightness — mutation vs SPEC (existing
     mutation+catch machinery scoped to handshake tests).
   - G4 build · vet · lint — deterministic (exists).
   - G5 test sensitivity — mutation vs the AGENT'S tests (same
     mutation engine, agent-test scope; hollow-test detection).
   - G6 independent check — METHOD diversity: the cage re-derivation
     (evidence, never self-report) + static analysis. Never a second
     LLM.
6. **Forward/hold** — a packet forwards autonomously unless
   inspection must hold it. Amber = advisory hold (sampled); red =
   blocking hold (strict lane, guardrail, irreversible). Holds feed
   the needs-you queue and debit the interrupt budget.
7. **Inspection — four modes**:
   - PULL: the Inspector — crack open any packet anytime (timeline,
     rich diff, annotations, live terminal activity). Looking never
     slows forwarding.
   - PUSH: the needs-you queue, budgeted interrupts.
   - STANDING: the gauntlet + watches — triggers carrying PRECISION
     SCORES counted from real fired-vs-useful history; a watch with
     no history shows "no history yet", never a fake number.
   - ADVERSARIAL: `packets probe` seeds a known-bad revision in a
     self-contained throwaway git repo (never touching any session's
     ledger or repo) and runs it through the real gates; the gates
     must catch it; results reported honestly (caught/escaped).
8. **Attention economics** — a real interrupt budget (n/week,
   counted down as holds interrupt), calibration draws (a random
   sample of auto-forwarded packets surfaced for skim), and
   empty-queue-is-success framing.
9. **Delivery + ACK** — delivered = ACK'd healthy, distinct from
   merged. Driven by an explicit host-invoked CLI command
   (`packets deployed` / `packets regressed`) backed by an optional
   re-checkable command whose exit code must agree with the
   asserted verb — never an agent's self-report. Until a packet has
   a real ACK, its settled-rail cell never renders delivered.
10. **Console + Inspector** — the only two primary surfaces
    (design/ui_kits/console/ was the starting reference; the built
    shape covers the same regions with an honest subset — no 24h
    bars or digest, since neither has a real data source yet):
    - Console `/` (360 | 1fr | 340): needs-you queue + calibration
      draw | `packets verified` hero stat, in-flight strip, lane
      health | settled rail (verified/held/delivered), your watches.
    - Inspector (252 | 1fr | 312): changed-files tree | rich diff +
      inline annotations | annotation rail; timeline footer (the six
      gates).
    `/settings` survives as a plain utility page (API key). `/board`
    survives as the cross-repo fleet listing (a capability Console,
    scoped to one addr, doesn't provide) — vocabulary-clean per
    slice 15, not folded in; `/review` is the Inspector's route.
11. **Learning/convergence** — a repo new to Packets has produced no
    real judgment yet; the Console's "learning" card shows the
    honest running count of settled packets (verified, held, OR
    delivered — same set as the settled rail) against a real
    minimum-sample threshold (5, mirroring watches' own
    sampled≥5 bar), flipping to "converged" once real history
    clears it. Scoped per session (matches every other per-session
    cache) — never a fabricated verdict, never per-repo persistence
    that doesn't exist elsewhere in the app.

## Vocabulary map (user-facing; binding)

| built term | becomes |
|---|---|
| order / work-order | packet |
| session / card | addr (the repo identity); "packet" in-flow |
| board (as a surface concept) | Console (`/`); the `/board` ROUTE itself survives unchanged as the cross-repo fleet listing (a capability Console doesn't provide), vocabulary-clean |
| review (surface) | Inspector; "inspect" as the verb; `/review` is its route |
| oracle | gate (e.g. "recheck the gate", "Gate running…", "Gate incomplete") |
| catch / no-catch / partial | gate output: tightened / gap found |
| verdict | gate results / packet state |
| land (clean/conflict/checks_red) | forward → verify → deliver; red/conflict = held (blocking) |
| PR / Approve & open PR / merged / merge blocked | forward / "forward →" / forwarded / held: \<reason\> |
| bounced | held |
| spend a catch / Place order | compose a packet |
| balance / stock | catches (a real count, never a currency word) |
| bets | claims (in-flight producer claims) |
| draft | intent (Analyze/Update intent) |
| bench | "queued targets:" |
| bandwidth | interrupt budget (a countdown, not a wallet) |
| dispatch | forward |
| question threads | annotations (Inspector rail) |
| hit-rate | substrate for precision scores; never shown as-is |

Banned on any surface (add a render test): PR, merge, approve, review
(noun), order, session, board, oracle, verdict, land, bounced, draft,
bench, spend, stock, balance, bet, LGTM. Voice: mono = machine's
voice; lowercase operational copy; counts with `·`; no exclamation
marks; no emoji; links end with a trailing →; empty states are
victories.

## Architecture (settled — do not relitigate)

- Server-rendered Go via `go-via/via` `h.*` + Datastar SSE; ONE
  hand-rolled CSS system in `internal/app/style.go`; Monaco as JS
  islands; NO Tailwind/React/shadcn/WebSockets.
- `internal/fabric` (JetStream) is the event source of truth. Event
  SUBJECTS and persisted JSON TAGS are FROZEN — rename Go
  types/fields freely, but wire names stay pinned; new concepts get
  NEW event kinds (gate results, hold, deliver, ack), never renames
  of existing ones.
- New `internal/packet`: a THIN read-model aggregate — Packet
  {name, addr, intent, revs, lane, handshake ref, gate results,
  lifecycle} folded from ledger/fabric events, plus new lifecycle
  events. The ledger stays the data-only append log; do not rewrite
  it.
- Proven machinery to reuse as-is: fabric, ledger, mutation, catch,
  pipe, harness (subprocess + container), settle, cage, ingest,
  diff, reanchor, translate, sandbox. `internal/refactor` is deleted.

## Integrity invariants (testable; survive any reshape)

1. Single mint path: one production writer of catch/economy records;
   every gate tier terminates in the existing append path — no
   parallel minter, ever.
2. Pre-specification: the agent never names its own denominator —
   anchors host-constructed; the handshake lives in paths the agent's
   turn cannot modify (settle deny-rule + content hash).
3. Secret-scrub on settle is pre-gate-zero: lane-independent, never
   skipped by any fast lane.
4. Cage stays no-egress (`--network=none`, cap-drop, read-only);
   probe/verification work never gets network. The agent container
   (egress-allowed) remains a separate runner.
5. Evidence over self-report: no state transition minted from an
   exit code or claim the untrusted side controls; hosts re-derive
   from structured evidence (the ACK signal included).
6. No fabricated metrics: every number a human sees is a pure fold
   over the append-only log or a real computation (import graph,
   trigger history). No lane badge, precision score, or delivered
   fill before its honest mechanic exists.

## Brand pack (port verbatim; source: design/tokens/)

Fonts: IBM Plex Sans + IBM Plex Mono, weights 400/500/600/700
(Google Fonts import; self-host later is fine).

Colors: --ground:#0a0d13; --surface-deep:#080a0f;
--surface-panel:#0f131b; --surface-card:#141924;
--surface-raised:#1b212e; --hairline:#1b212e;
--border-faint:rgba(255,255,255,.08); --border-mid:#2a3242;
--border-dashed:#3a4150; --ink:#e8ebf1; --text-body:#c3cad6;
--text-muted:#98a1b2; --text-faint:#616b7c; --text-ghost:#4e5666;
--text-disabled:#3a4150; --signal:#4cc4d4; --on-signal:#08222a;
--delivered:#2a7683; --delivered-mid:#357f8a; --verified:#46c08a;
--held:#e6b23e; --risk:#f0666b; --risk-deep:#d1585c;
--risk-muted:#cf9a97; --agent:#a78bfa; --you/--accent = signal;
--add = verified; --del = risk. (--heading-cream #F6EBCD is
marketing-only — never in-app.)

State grammar (the ONLY state colors, used everywhere): composing =
outline (delivered stroke @18% of cell, 8% tint); in-flight/you =
signal; verified/forwarded = verified; held advisory = held; held
blocking = risk; agent authorship = agent; delivered = delivered
fill. Outline → fill = promised → landed. Live pulses; settled is
matte.

Effects: --glow-signal:0 0 8px var(--signal); --glow-risk:0 0 10px
rgba(240,102,107,.55); --glow-btn:0 6px 18px rgba(76,196,212,.24);
--shadow-frame:0 30px 70px rgba(0,0,0,.45); --shadow-tooltip:0 12px
30px rgba(0,0,0,.5); --wash-signal:radial-gradient(90% 70% at 82%
-12%, rgba(76,196,212,.08), transparent 55%). Keyframes: pk-pulse
{0%,100%{opacity:1}50%{opacity:.35}} (1.8–2s); pk-held-pulse
{0%,100%{box-shadow:0 0 6px rgba(240,102,107,.35)}50%{box-shadow:0 0
16px rgba(240,102,107,.75)}}; pk-flow
{0%{transform:translateX(-140%)}100%{transform:translateX(240%)}}.
Transitions ~.2s interactive, .4s state recolor. No bounces/slides.
No blur, no transparency surfaces, no gradients beyond the wash.

Spacing: --sp-1..8: 4,7,9,12,14,16,22,32px. Radii: glyph 4, btn-sm 7,
card-sm 8, btn 8, ann 9, card 10, cta 11, stat 12, frame 14, shell
16, pill 999. Heights: chip 15/17/20; btn 26/28/34; cta 50.

Type: --fs-micro 9, tiny 9.5, small 10, label 10.5, body-mono 11,
body 12.5, emph 13, hero-stat 36 (h2 38 / h1 64 marketing-only).
Tracking: kicker .2em, label .1em, chip .05em, word .04em. Tabular
numerals always. Dense mono 9–13px in-app.

Layouts: Console grid `360px 1fr 340px`; Inspector `252px 1fr 312px`;
panel headers 12×16; page gutters 22–32px; card paddings 7–14px;
every region bounded by a 1px hairline; cards NEVER drop-shadow
(depth = app-frame mocks + tooltips only). Progress: 5–6px tracks,
3px radius. Bars: 1px-gapped verified bars; held markers = 4px amber
ticks. Empty states: dashed 1px border, centered mono microtype.

PacketMark (locked; code-built, never an asset): 2×2 grid; TL/BL
signal fill (BL risk when held-variant); TR ghost composing (stroke
--delivered @18% of cell, 8% tint, scale 1.035); BR --delivered fill.
Gap 27% of cell (min 1.5px), radius 23%. Below ~14px cell the ghost
outline is FORBIDDEN — solid --delivered-mid. Wordmark: Plex Mono 700
lowercase, +.04em, lockup gap ≈1.2×cell, size ≈1.7×cell. In-app: the
mark alone or the stacked `packets / LABEL` lockup — never mark +
wordmark + breadcrumb together.

Icons: unicode glyphs in Plex Mono only — ✎ ⚑ ● ◂ ▸ ⧉ ✱ ⇄ ＋ ✕ → ⌁ ✓.
The 15px rounded square with a letter (M/A/⚑) is the file-status
glyph. Never an icon library, never hand-drawn SVG icons.

Components (reference JSX in design/components/): Button
(primary/secondary/outline/pager; one primary per screen), Chip
(pill mono microtype, state color-mix 13–16% tint, border 40–45%
mix; neutral = flat raised), Card (row/dashed/tile variants; accent =
3px left border, authorship colors only), AnnotationCard (author
chip + severity blocking/self-flag/question/nit + location; sans
prose body; 3px left border in author color), PacketCell (8px state
square; live pulses; round = dot), Titlebar (locked: mark → stacked
lockup → hairline → name + rev chip + addr → status packet w/
tooltip → annotations legend), Timeline (packet-life dots:
plan grey, edit/comment cyan, test amber, mutation/flag purple, held
red big + glow), Terminal (flat header dots, dark paper, colored
author lines).
