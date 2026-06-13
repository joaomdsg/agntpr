package app

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-via/via"
	"github.com/go-via/via/vt"

	"github.com/joaomdsg/packets/internal/fabric"
	"github.com/joaomdsg/packets/internal/harness"
	"github.com/joaomdsg/packets/internal/ledger"
	"github.com/joaomdsg/packets/internal/translate"
)

// The review turn quotes the line under comment, so the orchestrator must read that
// exact source line from the working tree. readSourceLine returns the 1-indexed line's
// content, and degrades to "" (no quote) for a missing file or an out-of-range line —
// never an error that would block dispatching the adjustment.
func TestReadSourceLine_returnsTheOneIndexedLineOrEmpty(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("line one\nline two\nline three\n"), 0o644))

	assert.Equal(t, "line two", readSourceLine(dir, "a.go", 2), "the 1-indexed line content is returned")
	assert.Equal(t, "", readSourceLine(dir, "a.go", 99), "an out-of-range line degrades to no quote")
	assert.Equal(t, "", readSourceLine(dir, "missing.go", 1), "a missing file degrades to no quote")
	assert.Equal(t, "", readSourceLine(dir, "a.go", 0), "a non-positive line degrades to no quote")
}

// The keystone of the review-thread loop: leaving an adjustment on a line dispatches a
// harness turn (DESIGN §12.3) that tells the agent what to fix, against the session's
// live HEAD — reusing the live-order pipe. The dispatched order must carry the composed
// review turn (the comment + the address-it instruction), so the agent re-edits in
// place. NOT parallel (shared liveReg/liveFabric).
func TestLiveCard_addAdjustmentDispatchesAReviewTurnToTheHarness(t *testing.T) {
	resetConsumersForTest()
	repo := initGitRepoForOrder(t)
	head := gitOrder(t, repo, "rev-parse", "HEAD")

	restoreHarness := runHarness
	t.Cleanup(func() { runHarness = restoreHarness })
	runHarness = func(_ context.Context, _, _ string, _ func([]translate.UIEvent)) ([]harness.Turn, error) {
		return nil, nil
	}

	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "adjust", "i")
	bbase := time.Unix(1_700_000_000, 0)
	require.NoError(t, log.AppendBlock("q1", bbase))
	require.NoError(t, log.AppendUnblock("q1", bbase.Add(30*time.Second))) // +3 bandwidth funds the turn
	registerSession("adjust", LiveConfig{RepoDir: repo, BaseRev: head, Anchor: anchorForCap(), TestCmd: []string{"true"}}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	_, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	}, via.WithTestServer(&server))
	require.NoError(t, err)
	t.Cleanup(func() { _ = defLog.Close() })

	tc := vt.NewClient(t, server, "/?key=adjust")
	require.Equal(t, 200, tc.Action((&LiveCard{Key: "adjust"}).AddAdjustment).
		WithSignal("adjfile", "main.go").
		WithSignal("adjline", "3").
		WithSignal("adjtext", "handle the nil case here").Fire(),
		"leaving an adjustment is a calm, valid action")

	got := orderRecordFor(t, log, 1)
	assert.Contains(t, got.Target.Prompt, "handle the nil case here", "the dispatched turn carries the comment")
	assert.Contains(t, got.Target.Prompt, "main.go", "the turn names the commented file")
	assert.Contains(t, got.Target.Prompt, "Address this specifically", "the turn instructs the agent to address it")
	assert.Equal(t, head, got.Target.BaseRev, "the turn runs against the session's live HEAD")
}

// An empty adjustment comment is nothing to address — a silent no-op, never a
// dispatched turn. NOT parallel (shared globals).
func TestLiveCard_addAdjustmentIsANoOpOnEmptyComment(t *testing.T) {
	resetConsumersForTest()
	repo := initGitRepoForOrder(t)
	head := gitOrder(t, repo, "rev-parse", "HEAD")
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, "adjnoop", "i")
	registerSession("adjnoop", LiveConfig{RepoDir: repo, BaseRev: head, Anchor: anchorForCap(), TestCmd: []string{"true"}}, log)

	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	_, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	}, via.WithTestServer(&server))
	require.NoError(t, err)
	t.Cleanup(func() { _ = defLog.Close() })

	tc := vt.NewClient(t, server, "/?key=adjnoop")
	require.Equal(t, 200, tc.Action((&LiveCard{Key: "adjnoop"}).AddAdjustment).
		WithSignal("adjfile", "main.go").WithSignal("adjline", "3").WithSignal("adjtext", "   ").Fire())

	got := orderRecordFor(t, log, 1)
	assert.Equal(t, "", got.Target.Prompt, "an empty comment never dispatches a turn")
}
