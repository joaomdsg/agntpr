package app

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-via/via/vt"

	"github.com/joaomdsg/packets/internal/catch"
	"github.com/joaomdsg/packets/internal/fabric"
	"github.com/joaomdsg/packets/internal/ledger"
	"github.com/joaomdsg/packets/internal/mutation"
	"github.com/joaomdsg/packets/internal/pipe"
	"github.com/joaomdsg/packets/internal/reanchor"
)

// ROADMAP slice 3: "/" grows a 3-column Console shell around the untouched
// live-card content — a needs-you rail, the preserved center column, and a
// settled+watches rail. Without the grid the Lead has nowhere honest to see
// open questions or settled work at a glance. NOT parallel (shared
// liveReg/liveFabric).
func TestLiveCard_rendersConsoleGridWithThreeRegions(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.Contains(t, body, `class="console"`, "the card renders the console grid shell")
	require.Contains(t, body, "console__rail--needs-you", "the left rail is the needs-you region")
	require.Contains(t, body, "console__main", "the center column carries the preserved content")
	require.Contains(t, body, "console__rail--settled", "the right rail is the settled+watches region")
}

// The needs-you rail is the whole point of the slice: an open review thread
// (a surviving mutant from the last cycle) must surface as a card the Lead can
// click straight into /review, instead of being buried behind a full navigation.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_needsYouRailShowsAThreadCardForAnOpenFinding(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	e := lookupLiveEntry(defaultSessionKey)
	require.NotNil(t, e)
	e.setFindings([]mutation.Finding{
		{File: "alpha.go", Line: 7, Outcome: mutation.Survived, Message: "mutated >= to >"},
	})

	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.Contains(t, body, "needs you · 1", "the panel header counts the open thread")
	require.Contains(t, body, "alpha.go:7", "the card anchors the thread to its file:line")
	require.Contains(t, body, "mutated &gt;= to &gt;", "the card names the finding")
	require.Contains(t, body, `href="/review?key=default"`, "the whole card links into the session's review")
	require.Contains(t, body, "inspect →", "the link carries a trailing arrow naming its destination, per the house voice")
}

// A session drowning in open questions must stay scannable: the rail shows
// only the first few full cards and collapses the rest into a single "and N
// more" line, rather than growing the rail unboundedly. NOT parallel (shared
// liveReg/liveFabric).
func TestLiveCard_needsYouRailCapsFullCardsAndCollapsesTheRestIntoAMoreLine(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	e := lookupLiveEntry(defaultSessionKey)
	require.NotNil(t, e)
	var findings []mutation.Finding
	for i := 1; i <= 5; i++ {
		findings = append(findings, mutation.Finding{File: "alpha.go", Line: i, Outcome: mutation.Survived, Message: "finding"})
	}
	e.setFindings(findings)

	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.Contains(t, body, "needs you · 5", "the header honestly counts every open thread, not just the shown ones")
	require.Equal(t, 4, strings.Count(body, "console__thread-title"),
		"only the capped number of full cards render")
	require.Contains(t, body, "and 1 more", "the rest collapse into a single more-line")
}

// An empty needs-you queue is a VICTORY, not a dead screen — the honest empty
// state must say so calmly, never with an alarm or a fabricated placeholder.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_needsYouRailShowsVictoryEmptyStateWhenNoOpenThreads(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.Contains(t, body, "needs you · 0", "the panel header honestly counts zero open threads")
	require.Contains(t, body, "nothing needs you", "the empty state reads as a victory, not a dead end")
	require.Contains(t, body, "console__card--dashed", "the empty state uses the dashed empty-card treatment")
}

// The calibration draw mechanic (slice 11) does not exist yet — the rail must
// show the honest "no draws yet" empty state rather than a fabricated sample.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_calibrationRailShowsHonestEmptyStateBeforeItsMechanicExists(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.Contains(t, body, "calibration", "the calibration region is present")
	require.Contains(t, body, "no calibration draws yet", "no fake sample before the mechanic ships")
}

// The right rail's "settled" region is the honest replacement for a fake
// "recently delivered" — delivered isn't real until ACK. It lists the session's
// recent DONE dispatches, each with its caught/held outcome legible at a glance.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_settledRailListsADoneDispatch(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "settledc", "i")

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 101, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendDispatch("d2", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "beta.go", Line: 9}, own))
	require.NoError(t, log.AppendStatus(1, "done"))
	require.NoError(t, log.AppendStatus(2, "done")) // left with no matching catch — a settled MISS
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "alpha.go", Line: 7, ReasonTag: "catch", Producer: "wo:1"}))
	registerSession("settledc", LiveConfig{BaseRev: "own-b-settledc", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=settledc").HTML())
	require.Contains(t, body, "settled · 2", "the settled rail counts both done dispatches")
	require.Contains(t, body, "wo#1", "the row names the caught order's id")
	require.Contains(t, body, "wo#2", "the row names the missed order's id")
	require.Contains(t, body, `data-state="verified"`, "a caught order's state square is verified")
	require.Contains(t, body, `data-state="held"`, "a missed order's state square is held, distinct from verified")
}

// A session with nothing settled yet must show the honest dashed empty state,
// not an empty region masquerading as content. NOT parallel (shared
// liveReg/liveFabric).
func TestLiveCard_settledRailShowsDashedEmptyWhenNothingSettled(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.Contains(t, body, "settled · 0", "the settled rail honestly counts zero")
	require.Contains(t, body, "nothing settled yet", "the dashed empty state names the honest absence")
}

// Watches (slice 12) do not exist yet — the rail must show the honest empty
// state, never a fabricated watch. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_watchesRailShowsHonestEmptyStateBeforeItsMechanicExists(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.Contains(t, body, "your watches", "the watches region is present")
	require.Contains(t, body, "no watches yet", "no fabricated watch before the mechanic ships")
}

// The center column's header strip is the hero stat: the real Done count from
// the ledger, labelled "packets verified" — never "forwarded"/"delivered",
// which aren't mechanized yet. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_heroStatRendersTheVerifiedDoneCount(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "heroc", "i")

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendStatus(1, "done"))
	registerSession("heroc", LiveConfig{BaseRev: "own-b-heroc", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=heroc").HTML())
	require.Contains(t, body, `console__hero-stat">1<`, "the hero stat shows the real done count")
	require.Contains(t, body, "packets verified", "the hero stat is labelled with the honest mechanized state")
}

// A fresh session with zero done orders must show an honest zero, never omit
// the hero stat — a missing stat would read as a bug, not calm silence.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_heroStatShowsHonestZeroOnAFreshSession(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	body := bodyOf(vt.NewClient(t, server, "/").HTML())
	require.Contains(t, body, `console__hero-stat">0<`, "a fresh session honestly shows zero verified, not a hidden stat")
	require.Contains(t, body, "packets verified", "the label is present even at zero")
}

// The needs-you rail must stay live: a thread appearing off the connect-cycle
// path (e.g. a dispatched order's own findings, or a /review resolution
// arriving while "/" is open in another tab) is otherwise invisible to an open
// SSE connection until the Lead reloads. The Stream poll must fold the open-
// thread count into its existing signature so the rail re-renders live, without
// standing up a whole new SSE channel. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_needsYouRailRefreshesLiveWhenOpenThreadsChangeOffTheConnectCyclePath(t *testing.T) {
	restore := resolveCycle
	t.Cleanup(func() { resolveCycle = restore })
	resolveCycle = func(_ context.Context, _, _, _, _ string, _ reanchor.Anchor, _ []string, _, _ bool, _ chan<- pipe.TraceEvent) (Resolution, error) {
		return Resolution{Verdict: string(catch.NoCatch)}, nil
	}

	logPath := filepath.Join(t.TempDir(), "catches.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: logPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	tc := vt.NewClient(t, server, "/")
	frames, cancel := tc.SSE()
	defer cancel()
	vt.AwaitFrame(t, frames, 10*time.Second, "needs you · 0")

	e := lookupLiveEntry(defaultSessionKey)
	require.NotNil(t, e)
	e.setFindings([]mutation.Finding{
		{File: "alpha.go", Line: 7, Outcome: mutation.Survived, Message: "mutated >= to >"},
	})

	vt.AwaitFrame(t, frames, 10*time.Second, "needs you · 1")
}
