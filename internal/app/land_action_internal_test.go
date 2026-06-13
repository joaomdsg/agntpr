package app

import (
	"context"
	"errors"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-via/via"
	"github.com/go-via/via/vt"

	"github.com/joaomdsg/packets/internal/fabric"
	"github.com/joaomdsg/packets/internal/ledger"
)

// approveServer wires a funded session whose live HEAD is a real repo, returning the
// log + a /-mounted client, so Approve tests drive the action over HTTP.
func approveServer(t *testing.T, key string) (*ledger.Log, *httptest.Server) {
	t.Helper()
	resetConsumersForTest()
	repo := initGitRepoForOrder(t)
	head := gitOrder(t, repo, "rev-parse", "HEAD")
	ctx := context.Background()
	f, err := fabric.Start(ctx, t.TempDir())
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	log := ledger.Bind(f, key, "i")
	bbase := time.Unix(1_700_000_000, 0)
	require.NoError(t, log.AppendBlock("q1", bbase))
	require.NoError(t, log.AppendUnblock("q1", bbase.Add(30*time.Second)))
	registerSession(key, LiveConfig{RepoDir: repo, BaseRev: head, Anchor: anchorForCap(), TestCmd: []string{"true"}}, log)
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	_, defLog, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	}, via.WithTestServer(&server))
	require.NoError(t, err)
	t.Cleanup(func() { _ = defLog.Close() })
	return log, server
}

// Approve must REFUSE to open a PR when the tree can't land (red checks) and the Lead
// hasn't overridden — mirroring PR etiquette (you don't merge red CI). The PR
// subprocess is never invoked. NOT parallel (shared globals + openPR seam).
func TestApprove_refusesToOpenPRWhenBlocked(t *testing.T) {
	restore := openPR
	t.Cleanup(func() { openPR = restore })
	called := false
	openPR = func(_ context.Context, _, _, _, _ string) (string, error) { called = true; return "", nil }

	log, server := approveServer(t, "appblk")
	lookupLiveEntry("appblk").setLand("checks_red") // the integrated tree fails its checks
	_ = log

	tc := vt.NewClient(t, server, "/?key=appblk")
	require.Equal(t, 200, tc.Action((&LiveCard{Key: "appblk"}).Approve).Fire(),
		"a blocked approve is a calm no-op, never a crash")
	assert.False(t, called, "a blocked tree never opens a PR")
}

// With an override, Approve opens the PR despite a block — the guard is deliberate
// friction, not a hard gate (DESIGN §16). NOT parallel (shared globals).
func TestApprove_overrideOpensThePRDespiteABlock(t *testing.T) {
	restore := openPR
	t.Cleanup(func() { openPR = restore })
	var gotBranch, gotTitle string
	openPR = func(_ context.Context, _, branch, title, _ string) (string, error) {
		gotBranch, gotTitle = branch, title
		return "https://example/pr/1", nil
	}

	log, server := approveServer(t, "appovr")
	lookupLiveEntry("appovr").setLand("checks_red")
	require.NoError(t, log.AppendLiveDispatch("liveorder", ledger.Target{BaseRev: "h", Prompt: "Add the widget."}, ledger.Target{}))

	tc := vt.NewClient(t, server, "/?key=appovr")
	require.Equal(t, 200, tc.Action((&LiveCard{Key: "appovr"}).Approve).
		WithSignal("landoverride", "1").Fire())
	assert.Equal(t, prBranchName("appovr"), gotBranch, "the PR pushes the session's stable branch")
	assert.Equal(t, "Add the widget.", gotTitle, "the PR title comes from the session's task")
}

// A clean tree with no open threads opens the PR directly and surfaces the returned
// URL on the card, so the Lead sees their work landed. NOT parallel (shared globals).
func TestApprove_cleanTreeOpensPRAndSurfacesTheURL(t *testing.T) {
	restore := openPR
	t.Cleanup(func() { openPR = restore })
	openPR = func(_ context.Context, _, _, _, _ string) (string, error) {
		return "https://example/pr/42", nil
	}

	log, server := approveServer(t, "appok")
	require.NoError(t, log.AppendLiveDispatch("liveorder", ledger.Target{BaseRev: "h", Prompt: "Do the task."}, ledger.Target{}))

	tc := vt.NewClient(t, server, "/?key=appok")
	require.Equal(t, 200, tc.Action((&LiveCard{Key: "appok"}).Approve).Fire())

	body := bodyOf(vt.NewClient(t, server, "/?key=appok").HTML())
	assert.Contains(t, body, "https://example/pr/42", "the opened PR URL surfaces on the card")
}

// A push/PR failure must surface calmly on the card, never crash the action. NOT
// parallel (shared globals).
func TestApprove_pushFailureSurfacesCalmly(t *testing.T) {
	restore := openPR
	t.Cleanup(func() { openPR = restore })
	openPR = func(_ context.Context, _, _, _, _ string) (string, error) {
		return "", errors.New("push rejected")
	}

	_, server := approveServer(t, "appfail")
	tc := vt.NewClient(t, server, "/?key=appfail")
	require.Equal(t, 200, tc.Action((&LiveCard{Key: "appfail"}).Approve).Fire(),
		"a failed push is still a calm 200")

	body := bodyOf(vt.NewClient(t, server, "/?key=appfail").HTML())
	assert.Contains(t, body, "push rejected", "the failure reason surfaces calmly on the card")
}
