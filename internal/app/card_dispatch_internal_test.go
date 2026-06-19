package app

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-via/via/vt"

	"github.com/joaomdsg/packets/internal/catch"
	"github.com/joaomdsg/packets/internal/fabric"
	"github.com/joaomdsg/packets/internal/ledger"
	"github.com/joaomdsg/packets/internal/mutation"
)

// After a Spend funds a work-order, the Lead on the session card needs to watch
// THAT order resolve caught-or-missed — the payoff of the spend. Today the card
// shows only aggregate dispatch counts (queued/running/done); the per-order
// round-trip lives only on the fleet board, forcing a context-switch off the card
// the Lead is acting on. The live card must surface this session's recent
// work-orders with their caught/missed outcome, closing spend → dispatch → watch
// it resolve on one surface. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_showsThisSessionsDispatchRoundTripOutcomes(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "cardc", "i")

	// Fund two work-orders (two catches → balance 2), run both: WO#1 mints (caught),
	// WO#2 does not (missed) — the same round-trip the board already surfaces.
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 101, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendDispatch("d2", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "beta.go", Line: 9}, own))
	require.NoError(t, log.AppendStatus(1, "done"))
	require.NoError(t, log.AppendStatus(2, "done"))
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "alpha.go", Line: 7, ReasonTag: "catch", Producer: "wo:1"}))
	registerSession("cardc", LiveConfig{BaseRev: "own-b-cardc", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := vt.NewClient(t, server, "/?key=cardc").HTML()
	require.Contains(t, body, "WO#1", "the card shows the caught order by id")
	require.Contains(t, body, "alpha.go:7", "with its target line")
	require.Contains(t, body, "caught", "WO#1 minted → caught, shown on the card")
	require.Contains(t, body, "WO#2", "the card shows the missed order too")
	require.Contains(t, body, "missed", "WO#2 ran but minted nothing → missed, shown on the card")
	// Each resolved order carries a per-outcome hook so the calm palette can color
	// caught vs missed — the round-trip outcome legible at a glance, not
	// undifferentiated dim text (extends R45's per-state honesty to dispatches).
	require.Contains(t, body, `data-outcome="caught"`, "the caught order carries a per-outcome hook")
	require.Contains(t, body, `data-outcome="missed"`, "the missed order carries a per-outcome hook")
	// The stylesheet colors both outcomes in the honest palette (selectors live in
	// the head), never an alarm red/green.
	require.Contains(t, body, `[data-outcome="caught"]`, "the stylesheet colors a caught order")
	require.Contains(t, body, `[data-outcome="missed"]`, "the stylesheet colors a missed order")
}

// A session that has funded no work-orders must NOT render an empty dispatch
// cluster — an empty round-trip block is visual noise that implies activity where
// there is none. The cluster is omitted entirely until there is an order to show.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_omitsTheDispatchClusterWhenNoOrdersFunded(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	// bodyOf scopes past the <head>: the stylesheet contains "board-row__dispatches"
	// as a CSS selector, so a whole-page check would always match.
	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.NotContains(t, body, "board-row__dispatches",
		"a session with no funded orders renders no dispatch cluster, not an empty block")
}

// The whole point of a filled order is to inspect the changes the producer made —
// but the only drill-link into a settled order's diff was gated on it leaving open
// questions (surviving mutants). An order that fills cleanly, OR a miss that left
// zero questions (no mutable operator, lost-via-rename, oracle-incomplete), then
// surfaced NO clickable path to its own base→fix diff — the changes were
// unreachable by click. Every settled order must link into its /review diff.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_settledOrderAlwaysLinksToItsDiffEvenWithNoOpenQuestions(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "inspectc", "i")

	// One catch funds one order; it runs to done but mints nothing (missed) and
	// leaves zero findings → zero open questions: exactly the no-link blind spot.
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendStatus(1, "done"))
	registerSession("inspectc", LiveConfig{BaseRev: "own-b-inspectc", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=inspectc").HTML())
	require.Contains(t, body, "missed", "the order ran to done and minted nothing")
	require.Contains(t, body, `href="/review?key=inspectc&amp;wo=1"`,
		"a settled order with zero open questions still links into its base→fix diff")
	require.Contains(t, body, "inspect diffs",
		"the drill link reads as a diff inspection, not an open-questions count")
}

// The inspect link is about reaching the producer's edits, NOT about the catch
// outcome — so a CAUGHT order that left zero open questions must get it just the
// same as a missed one. Without this the link would read as "only failures are
// inspectable", which is wrong: a clean catch's diff is exactly what a Lead wants
// to review before landing. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_caughtOrderAlsoLinksToItsDiffWithZeroQuestions(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "caughtc", "i")

	// One catch funds the order; it runs to done AND mints its own catch (caught),
	// leaving zero findings → zero open questions.
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendStatus(1, "done"))
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "alpha.go", Line: 7, ReasonTag: "catch", Producer: "wo:1"}))
	registerSession("caughtc", LiveConfig{BaseRev: "own-b-caughtc", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=caughtc").HTML())
	require.Contains(t, body, "caught", "the order minted its own catch")
	require.Contains(t, body, `href="/review?key=caughtc&amp;wo=1"`,
		"a caught order links into its diff just like a missed one — inspection is outcome-independent")
	require.Contains(t, body, "inspect diffs")
}

// An order that DID leave open questions keeps its "N open questions" drill link
// (the test-debt affordance) and must NOT also sprout an "inspect diffs" link —
// the two are one link with different framing, never both. This locks the
// existing Questions>0 path against the new settled-link change.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_orderWithOpenQuestionsKeepsTheQuestionsLinkNotInspect(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "qlinkc", "i")

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendStatus(1, "done"))
	registerSession("qlinkc", LiveConfig{BaseRev: "own-b-qlinkc", FixRev: "own-f", Anchor: anchorForCap()}, log)
	e := lookupLiveEntry("qlinkc")
	require.NotNil(t, e)
	// The filled order left a surviving mutant → one open question for WO#1.
	e.setOrderFindings(1, []mutation.Finding{{File: "alpha.go", Line: 7, Outcome: mutation.Survived, Message: "mutated >= to >"}})

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=qlinkc").HTML())
	require.Contains(t, body, `href="/review?key=qlinkc&amp;wo=1"`, "the order still links into its review")
	require.Contains(t, body, "open questions", "a questioned order keeps the test-debt framing")
	require.NotContains(t, body, "inspect diffs",
		"an order with open questions shows ONE link (open questions), never also an inspect link")
}

// A queued/running order has not produced any edits yet, so it carries no diff to
// inspect — the inspect link must be gated on a SETTLED (done) order, never shown
// while work is still queued/running (which would 404-equivalent into an empty
// diff). NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_unsettledOrderShowsNoInspectLink(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "pendingc", "i")

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	registerSession("pendingc", LiveConfig{BaseRev: "own-b-pendingc", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=pendingc").HTML())
	require.Contains(t, body, "queued", "the order has not resolved")
	require.NotContains(t, body, "inspect diffs",
		"a queued order has produced no diff yet — no inspect link before it settles")
}

// A queued/running order has NOT resolved, so it must carry NO outcome hook — it
// stays neutral, never colored caught or missed before it has an outcome. Without
// this guard a bug could paint unresolved work as a confirmed catch or a loss.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_aQueuedOrderCarriesNoOutcomeHook(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "queuedc", "i")

	// One catch funds one order, left QUEUED (no done status, no outcome yet).
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	registerSession("queuedc", LiveConfig{BaseRev: "own-b-queuedc", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	// bodyOf scopes past the head — the stylesheet's [data-outcome=...] selectors
	// live there; we assert the rendered queued ORDER carries no outcome attribute.
	body := bodyOf(vt.NewClient(t, server, "/?key=queuedc").HTML())
	require.Contains(t, body, "WO#1", "the queued order is shown")
	require.Contains(t, body, "queued", "with its unresolved status")
	require.NotContains(t, body, "data-outcome=",
		"a queued order carries no outcome hook — it is never colored before it resolves")
}
