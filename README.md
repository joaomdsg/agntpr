# packets

A control plane for agent-written code changes. Agents compose, verify, and
forward changes ("packets") at line rate against machine-checkable contracts
("handshakes"); the few with real blast radius are held for a human.
Networking automated forwarding but never automated away inspection — packets
does the same for code.

The full conceptual model lives in [design/guidelines/concepts.md](design/guidelines/concepts.md);
[MVP.md](MVP.md) maps it onto what's actually built.

## Concepts

- **addr** — the repo a session tracks, `owner/repo`.
- **packet** — one agent-written change: named, revisioned, addressed, with
  a replayable timeline. Composed from a prose **intent**.
- **handshake** — a runnable contract authored independently of, and before,
  the agent's code (protected test paths the agent's turn cannot touch).
  Carries a strength gradient: examples → properties.
- **lane** — best-effort → standard → strict → irreversible, derived from a
  MEASURED blast radius (a pure function over the `go list` import graph of
  changed packages) — never self-reported. A stronger lane requires a
  stronger handshake and runs more gates.
- **the gauntlet** — six gates run per packet, scoped to what the lane
  warrants: G1 intent fidelity, G2 handshake conformance, G3 handshake
  tightness (mutation vs spec), G4 build·vet·lint, G5 test sensitivity
  (mutation vs the agent's own tests), G6 independent check (host
  re-derivation in an isolated cage, never a second LLM's opinion).
- **forward / hold** — a packet forwards on its own unless inspection must
  hold it: advisory (sampled, amber) or blocking (strict lane, gate failure,
  red). Holds feed the needs-you queue and debit the interrupt budget.
- **inspection, four modes** — pull (open any packet's diff/timeline
  anytime), push (the needs-you queue), standing (the gauntlet plus watches,
  each carrying a precision score counted from real fired-vs-useful history
  — "no history yet" until there is one), adversarial (`packets probe` runs
  a known-bad revision through the real gates in a throwaway repo and
  reports caught/escaped).
- **attention economics** — a real interrupt budget that counts down as
  holds interrupt, plus calibration draws (a random sample of
  auto-forwarded packets surfaced for a skim).
- **delivery + ACK** — delivered means ACK'd healthy, distinct from
  forwarded. Set only by an explicit host command (`packets deployed` /
  `packets regressed`), never an agent's self-report.

## Quick start

Requires **Go 1.26+** and **git**.

```bash
git clone <this repo>
cd packets
go build ./cmd/packets
```

Point it at a repo and a revision pair to review, then open the console:

```bash
./packets -repo . -base <weakSHA> -fix <fixSHA> -file adult.go -line 4
open http://localhost:3000
```

- `/` — Console: the needs-you queue and calibration draw, the packets
  hero stat and in-flight strip, the settled rail, your watches.
- `/review` — Inspector: a packet's changed-file tree, diff, inline
  annotations, and gate timeline.
- `/board` — the cross-repo fleet listing.
- `/settings` — configure the Anthropic API key a live order's harness uses.

Seed a **live order** — a real Claude Code harness runs your task to produce
the fix. Needs the `claude` binary on `PATH` plus `ANTHROPIC_API_KEY`:

```bash
./packets -repo . -base <sha> -fix <sha> -file a.go -line 4 \
  -live 'file=pkg/y.go,line=12,base=<sha>,prompt=fix the off-by-one in the loop bound'
```

By default a live order runs the harness as a host subprocess. To run it in
a hardened, egress-allowed container instead:

```bash
docker build -f internal/harness/Dockerfile -t packets-agent .
ANTHROPIC_API_KEY=... ./packets -container -repo . -base <sha> -file a.go -line 4 \
  -live 'file=pkg/y.go,line=12,base=<sha>,prompt=fix the off-by-one'
```

Track several addrs from one server, each its own isolated console:

```bash
./packets -repo . -base <sha> -fix <sha> -file a.go -line 4 \
  -session 'key=rate-limiter,base=<sha>,fix=<sha>,file=rl.go,line=12'
# default console at /  ·  keyed console at /?key=rate-limiter
```

> If `go` errors with a `GOROOT` version mismatch, prefix commands with
> `env -u GOROOT`.

Or run the scripted walkthrough, which builds a real scratch repo and drives
compose → forward → hold → inspect → deliver end to end:

```bash
./scripts/demo.sh
```

### CLI commands

| Command | Does |
|---|---|
| `packets` | starts the server (flags above) |
| `packets verify-catch -repo -base -fix -file -line [-tip]` | runs the mutation/catch oracle over a revision pair, prints the verdict as JSON |
| `packets deployed -ledger -session -wo [-check]` | ACKs a packet as delivered (healthy); an optional check command's exit code must agree |
| `packets regressed -ledger -session -wo [-check]` | ACKs a packet as regressed |
| `packets probe` | adversarial mode: seeds a known-bad revision in a throwaway repo, runs it through the real gates, reports caught/escaped |

## Architecture

Server-rendered Go via [go-via/via](https://github.com/go-via/via) `h.*` +
Datastar SSE — one hand-rolled CSS system, Monaco as JS islands. No
Tailwind/React/shadcn/WebSockets.

| Package | Responsibility |
|---|---|
| `internal/fabric` | embedded JetStream — the event-sourced source of truth |
| `internal/ledger` | append-only economy log folded from fabric events |
| `internal/packet` | the packet read-model (lane, handshake, gauntlet, hold) folded from the ledger |
| `internal/mutation` | diff-scoped mutation oracle — deep payload inspection |
| `internal/catch` | the confirmed-catch oracle: the pure base→fix differential over mutation |
| `internal/diff` | structured git diff (changed files, hunks, line ranges) |
| `internal/reanchor` | maps a comment's line anchor across revisions |
| `internal/review` | anchored annotation threads, surviving-mutant questions |
| `internal/settle` | turns a harness turn into a git revision (no-edit guard, secret scrub) |
| `internal/orchestrator` | host-side coordinator: composes settle + diff into a minted revision |
| `internal/surface` | projects gate verdicts into Console/Inspector view models |
| `internal/harness` | the Claude Code agent runner (host subprocess or hardened container) |
| `internal/cage` | the hardened, no-egress verification sandbox behind G6 |
| `internal/app` | the Console/Inspector server, live sessions, fleet board |
| `cmd/packets` | CLI entrypoint and flag/session wiring |

## Testing

```bash
./scripts/test-fast.sh
```

Runs the full suite split across two concurrent groups: everything except
`internal/cage`/`internal/sandbox` at full parallelism, and that pair
serialized (`-p 1`) since both assert on a shared docker container label.

## Documents

- **[MVP.md](MVP.md)** — the binding MVP spec: concept → feature checklist, vocabulary map, integrity invariants, brand pack.
- **[ROADMAP.md](ROADMAP.md)** — the slice-by-slice build ledger.
- **[LOOP.md](LOOP.md)** — the autonomous build loop's process contract.
- **[design/](design/)** — the design system: tokens, components, guidelines, the concept model.
- **[CONVENTIONS.md](CONVENTIONS.md)** — coding conventions.

## License

[MIT](LICENSE) © João Gonçalves
