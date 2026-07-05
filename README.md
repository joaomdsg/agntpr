# packets

<p><img src="design/assets/branding/badge-experimental.svg" alt="status: experimental" height="28"></p>

A control plane for agent-written code changes. Agents compose, verify, and
forward changes ("packets") at line rate against machine-checkable contracts
("handshakes"); the few with real blast radius are held for a human.
Networking automated forwarding but never automated away inspection — packets
does the same for code.

Nothing self-reports. A lane is measured, not declared; a catch is a
differential a mutation survived, not a claim; a delivery is an ACK, not an
agent saying "done". Where a number can't be earned honestly, the surface
says so — `no history yet` — rather than inventing one.

The fuller conceptual model lives in
[design/guidelines/concepts.md](design/guidelines/concepts.md);
[MVP.md](MVP.md) maps it onto what's actually built.

## Concepts

- **addr** — the repo a session tracks, in `owner/repo` form.
- **packet** — one agent-written change: named, revisioned, addressed, with a
  replayable timeline. Composed from a prose **intent**.
- **handshake** — a runnable contract authored independently of, and before,
  the agent's code: protected test paths the agent's turn cannot touch.
  Carries a strength gradient, examples → properties.
- **lane** — best-effort → standard → strict → irreversible, derived from a
  MEASURED blast radius (a pure function over the `go list` import graph of
  the changed packages), never self-reported. A stronger lane requires a
  stronger handshake and runs more gates.
- **the gauntlet** — six gates per packet, scoped to what the lane warrants:
  G1 intent fidelity, G2 handshake conformance, G3 handshake tightness
  (mutation vs the spec), G4 build·vet·lint, G5 test sensitivity (mutation vs
  the agent's own tests), G6 independent check (host re-derivation in an
  isolated cage — never a second LLM's opinion).
- **forward / hold** — a packet forwards on its own unless inspection must
  hold it: advisory (sampled, amber) or blocking (strict lane, gate failure,
  red). Holds feed the needs-you queue and debit the interrupt budget.
- **inspection, four modes** — pull (open any packet's diff and timeline
  anytime), push (the needs-you queue), standing (the gauntlet plus watches,
  each carrying a precision score counted from real fired-vs-useful history —
  `no history yet` until there is one), adversarial (`packets probe` runs a
  known-bad revision through the real gates and reports caught / escaped).
- **attention economics** — a real interrupt budget that counts down as holds
  interrupt, plus calibration draws: a random sample of auto-forwarded packets
  surfaced for a skim. An empty queue is success, not idleness.
- **learning / convergence** — a repo new to packets is *learning* until it
  has accumulated enough real settled history; the console shows the honest
  count against the threshold, then *converged*. Never a fabricated result.
- **delivery + ACK** — delivered means ACK'd healthy, distinct from forwarded.
  Set only by an explicit host command (`packets deployed` / `packets
  regressed`), never an agent's self-report.

## Quick start

Requires **Go 1.26+** and **git**.

```bash
git clone https://github.com/joaomdsg/packets.git
cd packets
go build ./cmd/packets
```

**See it work.** The scripted walkthrough builds a throwaway repo with three
real revisions — an under-tested boundary, a fix that catches it, and a fix
with a broken build — and drives the whole pipeline end to end: compose →
forward → hold → inspect → deliver.

```bash
./scripts/demo.sh
```

**Review a real revision pair.** Point packets at a repo and a base→fix pair,
anchor a comment at a file and line, then browse to the console:

```bash
./packets -repo . -base <baseSHA> -fix <fixSHA> -file pkg/auth/session.go -line 42
# console at http://localhost:3000
```

- `/` — **Console**: the needs-you queue and calibration draw; the packets
  hero stat, in-flight strip, and lane health; the settled rail, the learning
  card (a repo's convergence progress), and your watches.
- `/review` — **Inspector**: a packet's changed-file tree, diff, inline
  annotations, and the six-gate timeline.
- `/board` — the cross-repo fleet listing, one row per addr.
- `/settings` — the Anthropic API key a live order's harness uses.

**Seed a live order** — a real Claude Code harness runs your task to produce
the fix. Needs the `claude` binary on `PATH` and `ANTHROPIC_API_KEY`:

```bash
./packets -repo . -base <baseSHA> -fix <fixSHA> -file pkg/auth/session.go -line 42 \
  -live 'file=pkg/ratelimit/window.go,line=12,base=<baseSHA>,prompt=fix the off-by-one in the loop bound'
```

By default the harness runs as a host subprocess. To run it in a hardened,
egress-allowed container instead:

```bash
docker build -f internal/harness/Dockerfile -t packets-agent .
ANTHROPIC_API_KEY=... ./packets -container -repo . -base <baseSHA> -file pkg/auth/session.go -line 42 \
  -live 'file=pkg/ratelimit/window.go,line=12,base=<baseSHA>,prompt=fix the off-by-one'
```

**Track several addrs** from one server, each its own isolated console:

```bash
./packets -repo . -base <baseSHA> -fix <fixSHA> -file pkg/auth/session.go -line 42 \
  -session 'key=rate-limiter,base=<baseSHA>,fix=<fixSHA>,file=pkg/ratelimit/window.go,line=12'
# default console at /  ·  keyed console at /?key=rate-limiter
```

> If `go` errors with a `GOROOT` version mismatch, prefix commands with
> `env -u GOROOT`.

### CLI commands

| Command | Does |
|---|---|
| `packets` | starts the server (flags above) |
| `packets verify-catch -repo -base -fix -file -line [-tip]` | runs the mutation/catch oracle over a revision pair; prints the outcome as JSON |
| `packets deployed -ledger -session -wo [-check]` | ACKs a packet as delivered (healthy); an optional `-check` command's exit code must agree |
| `packets regressed -ledger -session -wo [-check]` | ACKs a packet as regressed |
| `packets probe` | adversarial mode: seeds a known-bad revision in a throwaway repo, runs it through the real gates, reports caught / escaped |

## Architecture

Server-rendered Go — [go-via/via](https://github.com/go-via/via) `h.*`
compositions over Datastar SSE, one hand-rolled CSS system, Monaco as JS
islands. No Tailwind, React, shadcn, or WebSockets.

The packages read as the signal path: events land on the spine, fold into
read-models, get put through the gauntlet, and surface to the operator — with
a producer path that turns an agent's turn into a revision worth folding.

**the spine · event source of truth**
- `internal/fabric` — embedded JetStream; the event-sourced source of truth
- `internal/ledger` — append-only log of confirmed catches, folded from fabric events
- `internal/bridge` — feeds a session's economy from that log

**read-models · folded from the ledger**
- `internal/packet` — the packet aggregate: lane, handshake, gauntlet, hold, lifecycle
- `internal/review` — anchored annotation threads and surviving-mutant questions
- `internal/reanchor` — maps a comment's line anchor across revisions

**the gauntlet · verification machinery**
- `internal/diff` — structured git diff (changed files, hunks, line ranges)
- `internal/mutation` — the diff-scoped mutation oracle
- `internal/catch` — the confirmed-catch oracle: the pure base→fix differential over mutation
- `internal/cage` — the hardened, no-egress verification sandbox behind G6
- `internal/sandbox` — runs untrusted work in a one-shot, locked-down container

**producing a packet · an agent's turn → a minted revision**
- `internal/harness` — the Claude Code agent runner (host subprocess or hardened container)
- `internal/translate` — maps the harness's stream-json events into the pipeline's own
- `internal/settle` — turns a harness turn into a git revision (no-edit guard, secret scrub)
- `internal/orchestrator` — host-side coordinator: composes settle + diff into a minted revision
- `internal/pipe` — wires settle, diff, mutation, and reanchor into one catch cycle
- `internal/ingest` — trusted host-side intake of an untrusted producer's git objects
- `internal/assist` — the producer's live analysis of a work-order draft

**surfaces · what the operator sees**
- `internal/surface` — projects gate results into Console/Inspector view models
- `internal/app` — the Console/Inspector server, live sessions, and fleet board
- `internal/tokenstore` — persists the single Anthropic API key the live server uses
- `cmd/packets` — the CLI entrypoint and flag/session wiring

## Testing

```bash
./scripts/test-fast.sh
```

Runs the full suite as two concurrent groups: everything except
`internal/cage` and `internal/sandbox` at full parallelism, and that pair
serialized (`-p 1`) because both assert on a shared docker container label.

## Documents

- **[MVP.md](MVP.md)** — the binding MVP spec: concept → feature checklist, vocabulary map, integrity invariants, brand pack.
- **[ROADMAP.md](ROADMAP.md)** — the slice-by-slice build ledger.
- **[LOOP.md](LOOP.md)** — the autonomous build loop's process contract.
- **[CONVENTIONS.md](CONVENTIONS.md)** — coding and test conventions.
- **[design/](design/)** — the design system: concept model, tokens, voice, and the component catalog.

## License

[MIT](LICENSE) © João Gonçalves

<p>
  <img src="design/assets/branding/badge-built-with-via.svg" alt="built with via" height="28">
  <img src="design/assets/branding/badge-rights-joaomdsg.svg" alt="© joaomdsg" height="28">
</p>
