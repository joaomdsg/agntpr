package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-via/via/vt"

	"github.com/joaomdsg/packets/internal/app"
)

// An empty adjustment comment is nothing to address — a silent no-op, never a
// dispatched turn. NOT parallel (shared globals).
func TestLiveCard_addAdjustmentIsANoOpOnEmptyComment(t *testing.T) {
	repo := initGitRepoForOrder(t)
	head := gitOrder(t, repo, "rev-parse", "HEAD")

	server, _ := bootDefaultServer(t, defaultBootCfg)
	log := addFundedSession(t, "adjnoop", app.LiveConfig{RepoDir: repo, BaseRev: head, Anchor: anchorForCap(), TestCmd: []string{"true"}})

	tc := vt.NewClient(t, server, "/review?key=adjnoop")
	require.Equal(t, 200, tc.Action((&app.ReviewCard{Key: "adjnoop"}).AddAdjustment).
		WithSignal("adjfile", "main.go").WithSignal("adjline", "3").WithSignal("adjtext", "   ").Fire())

	got := orderRecordFor(t, log, 1)
	assert.Equal(t, "", got.Target.Prompt, "an empty comment never dispatches a turn")
}

// The review surface must render the adjustment entry point — inputs bound to the
// adjustment signals and a button wired to AddAdjustment — else the Lead has no way to
// leave an adjustment (the comment→harness round-trip would be unreachable from the
// UI). NOT parallel (shared globals).
func TestReviewCard_rendersTheAdjustmentEntryPoint(t *testing.T) {
	repo := initGitRepoForOrder(t)
	head := gitOrder(t, repo, "rev-parse", "HEAD")

	server, _ := bootDefaultServer(t, defaultBootCfg)
	_ = addFundedSession(t, "adjui", app.LiveConfig{RepoDir: repo, BaseRev: head, Anchor: anchorForCap(), TestCmd: []string{"true"}})

	body := bodyOf(vt.NewClient(t, server, "/review?key=adjui").HTML())
	assert.Contains(t, body, "/_action/AddAdjustment", "the review surface renders the leave-adjustment action")
	assert.Contains(t, body, `data-bind="adjtext"`, "with an input bound to the adjustment comment signal")
}
