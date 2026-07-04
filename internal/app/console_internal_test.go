package app

import (
	"context"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-via/via/vt"

	"github.com/joaomdsg/packets/internal/catch"
	"github.com/joaomdsg/packets/internal/fabric"
	"github.com/joaomdsg/packets/internal/ledger"
	"github.com/joaomdsg/packets/internal/pipe"
	"github.com/joaomdsg/packets/internal/reanchor"
)

// gitInitNoRemote initializes a bare local repo with no origin remote, so
// packet.ParseAddr on it deterministically falls back to the honest
// "local/<dir>" identity instead of resolving a real remote.
func gitInitNoRemote(t testing.TB, dir string) {
	t.Helper()
	require.NoError(t, exec.Command("git", "init", "-q", dir).Run())
}

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

// ROADMAP slice 6 (packet 6.MVP): the needs-you rail is now driven by the
// packet fold, not the session's raw connect-cycle findings — a HELD packet
// (blocking or advisory) surfaces as a card the Lead can click straight into
// its own /review?wo=<id>. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_needsYouRailShowsAHeldPacketCardWithItsReason(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "needsyou", "i")

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendStatus(1, "failed")) // a run failure — held, blocking
	registerSession("needsyou", LiveConfig{BaseRev: "own-b-needsyou", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=needsyou").HTML())
	require.Contains(t, body, "needs you · 1", "the panel header counts the held packet")
	require.Contains(t, body, "run failed", "the card names the one-clause hold reason")
	require.Contains(t, body, `href="/review?key=needsyou&amp;wo=1"`, "the card links straight into the packet's own review")
	require.Contains(t, body, "inspect →", "the link carries a trailing arrow naming its destination, per the house voice")
}

// Blocking holds are the most attention-worthy — they must render ahead of
// advisory holds regardless of dispatch recency. NOT parallel (shared
// liveReg/liveFabric).
func TestLiveCard_needsYouRailOrdersBlockingPacketsBeforeAdvisoryOnes(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "needsyouorder", "i")
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}

	// Order 1 (oldest — RecentDispatches is newest-first, so without the
	// blocking-before-advisory sort this would render SECOND): a run failure.
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 1}, own))
	require.NoError(t, log.AppendStatus(1, "failed"))
	// Order 2 (newest): done with no catch — an advisory gap, not a failure.
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 101, ReasonTag: "catch"}))
	require.NoError(t, log.AppendDispatch("d2", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "beta.go", Line: 2}, own))
	require.NoError(t, log.AppendStatus(2, "done"))
	registerSession("needsyouorder", LiveConfig{BaseRev: "own-b-needsyouorder", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=needsyouorder").HTML())
	blockingIdx := strings.Index(body, "run failed")
	advisoryIdx := strings.Index(body, "gap found")
	require.True(t, blockingIdx >= 0 && advisoryIdx >= 0, "both hold reasons render")
	require.Less(t, blockingIdx, advisoryIdx, "the blocking packet renders before the advisory one — most attention-worthy first")
}

// A session drowning in open questions must stay scannable: the rail shows
// only the first few full cards and collapses the rest into a single "and N
// more" line, rather than growing the rail unboundedly. NOT parallel (shared
// liveReg/liveFabric).
func TestLiveCard_needsYouRailCapsFullCardsAndCollapsesTheRestIntoAMoreLine(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "needsyoucap", "i")

	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	for i := 1; i <= 5; i++ {
		require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100 + i, ReasonTag: "catch"}))
		require.NoError(t, log.AppendDispatch("d", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: i}, own))
		require.NoError(t, log.AppendStatus(i, "failed"))
	}
	registerSession("needsyoucap", LiveConfig{BaseRev: "own-b-needsyoucap", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=needsyoucap").HTML())
	require.Contains(t, body, "needs you · 5", "the header honestly counts every held packet, not just the shown ones")
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

// ROADMAP slice 6: the settled rail now shows lifecycle-colored rows straight
// from the packet fold — verified, held-advisory, and held-blocking (a run
// failure IS settled-red) all count; only composing/in-flight are excluded.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_settledRailShowsLifecycleColoredRows(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "settledc", "i")

	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendStatus(1, "done"))
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "alpha.go", Line: 7, ReasonTag: "catch", Producer: "wo:1"}))

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 101, ReasonTag: "catch"}))
	require.NoError(t, log.AppendDispatch("d2", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "beta.go", Line: 9}, own))
	require.NoError(t, log.AppendStatus(2, "done")) // left with no matching catch — a settled advisory gap

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 102, ReasonTag: "catch"}))
	require.NoError(t, log.AppendDispatch("d3", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "gamma.go", Line: 11}, own))
	require.NoError(t, log.AppendStatus(3, "failed")) // a run failure — settled, blocking

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
	require.Contains(t, body, "settled · 3", "verified, held-advisory, and held-blocking orders all count as settled")
	require.Contains(t, body, `data-state="verified"`, "the caught order's state square is verified")
	require.Contains(t, body, `data-state="held-blocking"`, "the failed order's state square is the blocking-red hue")
	require.Contains(t, body, `data-state="held"`, "the missed order's state square is the advisory-amber hue")
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

// ROADMAP slice 6: the hero stat counts ONLY packets whose State is Verified
// (done AND the order's own catch minted AND no open questions) — a
// done-but-missed order no longer counts, unlike the old raw Done tally.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_heroStatRendersTheVerifiedDoneCount(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "heroc", "i")

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 101, ReasonTag: "catch"}))
	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendDispatch("d2", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "beta.go", Line: 9}, own))
	require.NoError(t, log.AppendStatus(1, "done"))
	require.NoError(t, log.AppendStatus(2, "done")) // done but left with no matching catch — NOT verified
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "alpha.go", Line: 7, ReasonTag: "catch", Producer: "wo:1"}))
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
	require.Contains(t, body, `console__hero-stat">1<`, "only the genuinely verified order counts — the done-but-missed one does not")
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

// ROADMAP slice 6: the center column gains an "in flight · N" strip under the
// hero — one pulsing signal cell per running order, one ghost-outline cell
// per queued (composing) order. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_inFlightStripShowsPulsingRunningAndGhostQueuedRows(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "inflight", "i")

	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 100, ReasonTag: "catch"}))
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendStatus(1, "running"))
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 101, ReasonTag: "catch"}))
	require.NoError(t, log.AppendDispatch("d2", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "beta.go", Line: 9}, own))
	// order 2 gets no status appended — it defaults to "queued" (composing).
	registerSession("inflight", LiveConfig{BaseRev: "own-b-inflight", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=inflight").HTML())
	require.Contains(t, body, "in flight · 2", "the strip counts both the running and the queued order")
	require.Contains(t, body, `data-state="in-flight"`, "the running order renders the pulsing signal cell")
	require.Contains(t, body, `data-state="composing"`, "the queued order renders the ghost-outline cell")
}

// The strip is a live-activity affordance only — it must be entirely absent
// when nothing is composing or in flight, not an empty header. NOT parallel
// (shared liveReg/liveFabric).
func TestLiveCard_inFlightStripAbsentWhenNothingInFlightOrComposing(t *testing.T) {
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
	require.NotContains(t, body, "console__inflight", "the strip is omitted entirely, not shown empty")
}

// ROADMAP slice 6: the console hero region carries a mono addr line, using
// the honest repo identity (packet.ParseAddr) rather than a raw folder name —
// a remote-less repo falls back to "local/<dir>", never a fabricated owner.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_addrLineRendersTheLocalFallbackForARemotelessRepo(t *testing.T) {
	dir := t.TempDir()
	gitInitNoRemote(t, dir)

	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "addrline", "i")
	registerSession("addrline", LiveConfig{RepoDir: dir}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=addrline").HTML())
	require.Contains(t, body, "addr local/"+filepath.Base(dir), "a remote-less repo falls back to the honest local/<dir> identity")
}

// ROADMAP slice 6: the four retired economy meter rows (stock/balance/
// bandwidth/dispatch) must never render on the console — MVP.md retires them
// from the UI outright. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_economyMeterRowsAreGoneFromTheConsole(t *testing.T) {
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
	for _, hook := range []string{
		`data-state="stock"`, `data-state="balance"`, `data-state="bandwidth"`, `data-state="dispatch"`,
	} {
		require.NotContainsf(t, body, hook, "the retired economy meter %q must not render on the console", hook)
	}
}

// The needs-you/hero/settled regions must stay live even when the fold-
// visible fact that changes them is a catch minting AFTER an order's status
// already settled to "done" — no status or question change accompanies it,
// so a signature keyed only on those facts would miss the transition. NOT
// parallel (shared liveReg/liveFabric).
func TestLiveCard_heroStatRefreshesLiveWhenACatchMintsWithoutAStatusOrQuestionChange(t *testing.T) {
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

	own := ledger.Target{BaseRev: "ob", FixRev: "of", TipRev: "of", Path: "own.go", Line: 1}
	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "c.go", Line: 1, ReasonTag: "catch"}))
	require.NoError(t, log.AppendDispatch("d1", ledger.Target{BaseRev: "b", FixRev: "f", TipRev: "f", Path: "alpha.go", Line: 7}, own))
	require.NoError(t, log.AppendStatus(1, "done")) // done, not yet caught — held, not verified

	tc := vt.NewClient(t, server, "/")
	frames, cancel := tc.SSE()
	defer cancel()
	vt.AwaitFrame(t, frames, 10*time.Second, `console__hero-stat">0<`)

	require.NoError(t, log.Append(ledger.CatchRecord{Outcome: catch.Catch, Path: "alpha.go", Line: 7, ReasonTag: "catch", Producer: "wo:1"}))

	vt.AwaitFrame(t, frames, 10*time.Second, `console__hero-stat">1<`)
}
