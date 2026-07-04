package app

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-via/via/vt"

	"github.com/joaomdsg/packets/internal/fabric"
	"github.com/joaomdsg/packets/internal/ledger"
)

// A fresh session has interrupted the Lead zero times this week — the KPI
// must show the honest zero against the LOCKED weekly cap of 10
// (design/ui_kits/console/ConsoleScreen.jsx's "5/10 interrupts" header),
// never omit the stat or invent a different cap. NOT parallel (shared
// liveReg/liveFabric).
func TestLiveCard_interruptKPIShowsHonestZeroOnAFreshSession(t *testing.T) {
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
	require.Contains(t, body, "0/10 interrupts", "a session that has raised no interrupts shows the honest zero against the fixed cap")
}

// The KPI must reflect a REAL count of raised interrupts (ledger blocks), not
// a fabricated number — it changes as real blocks are logged. NOT parallel
// (shared liveReg/liveFabric).
func TestLiveCard_interruptKPIReflectsRealLoggedInterrupts(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "interruptkpi", "i")
	require.NoError(t, log.AppendBlock("q:1", time.Now()))
	require.NoError(t, log.AppendBlock("q:2", time.Now()))
	registerSession("interruptkpi", LiveConfig{BaseRev: "own-b-interruptkpi", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=interruptkpi").HTML())
	require.Contains(t, body, "2/10 interrupts", "the KPI counts the two real interrupts this session logged, against the fixed cap")
}

// The KPI must show the REAL count even past the fixed cap, never clamp it —
// clamping would hide from the Lead exactly how far over budget they ran.
// NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_interruptKPIShowsTheRealCountUncappedWhenOverBudget(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "interruptover", "i")
	for i := 0; i < 11; i++ {
		require.NoError(t, log.AppendBlock("q:"+string(rune('a'+i)), time.Now()))
	}
	registerSession("interruptover", LiveConfig{BaseRev: "own-b-interruptover", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	body := bodyOf(vt.NewClient(t, server, "/?key=interruptover").HTML())
	require.Contains(t, body, "11/10 interrupts", "the real over-budget count is shown honestly, never clamped at the cap")
}

// A new interrupt raised while a tab is connected must update the KPI over
// the LIVE SSE stream, not just on the next full reload — the same pattern
// TestLiveCard_heroStatRefreshesLiveWhenACatchMintsWithoutAStatusOrQuestionChange
// already proves for the hero stat's dispatch-tally signature. NOT parallel
// (shared liveReg/liveFabric).
func TestLiveCard_interruptKPIRefreshesLiveWhenANewInterruptIsRaised(t *testing.T) {
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "interruptlive", "i")
	registerSession("interruptlive", LiveConfig{BaseRev: "own-b-interruptlive", FixRev: "own-f", Anchor: anchorForCap()}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = defLog.Close() })

	tc := vt.NewClient(t, server, "/?key=interruptlive")
	frames, cancel := tc.SSE()
	defer cancel()
	vt.AwaitFrame(t, frames, 10*time.Second, "0/10 interrupts")

	require.NoError(t, log.AppendBlock("q:1", time.Now()))

	vt.AwaitFrame(t, frames, 10*time.Second, "1/10 interrupts")
}
