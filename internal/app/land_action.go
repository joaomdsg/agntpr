package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-via/via"
	"github.com/go-via/via/h"
	"github.com/go-via/via/on"

	"github.com/joaomdsg/packets/internal/ledger"
	"github.com/joaomdsg/packets/internal/pipe"
)

// openPR is the seam the approve flow opens a PR through (process I/O — verified by
// build + manual run, like runHarness; tests swap it for a scripted reply). It pushes
// the session's branch and opens a PR, returning the PR URL.
var openPR = realOpenPR

// realOpenPR pushes the session repo's current HEAD to branch and opens a PR against
// the default branch via gh, returning the PR URL it prints. v1 pushes a direct branch
// + PR (the merge-queue path, DESIGN §29.2, is deferred). Auth: in production a
// short-lived token is injected into the environment at land time (the agntpr
// x-access-token pattern) — never a long-lived host credential.
func realOpenPR(ctx context.Context, repoDir, branch, title, body string) (string, error) {
	push := exec.CommandContext(ctx, "git", "push", "--force-with-lease", "origin", "HEAD:refs/heads/"+branch)
	push.Dir = repoDir
	if out, err := push.CombinedOutput(); err != nil {
		return "", fmt.Errorf("push: %v: %s", err, strings.TrimSpace(string(out)))
	}
	pr := exec.CommandContext(ctx, "gh", "pr", "create", "--head", branch, "--title", title, "--body", body)
	pr.Dir = repoDir
	out, err := pr.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr create: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// Approve closes the goal flow: it opens a PR from the session's work (DESIGN §16).
// First it guards (landBlocked): an unmergeable tree (rebase conflict / red checks) or
// open review threads refuses the PR unless the Lead overrides (deliberate, overridable
// friction). Cleared, it derives the branch + PR title/body and opens the PR through
// the openPR seam, caching the URL (or a calm failure / guard message) for the card.
func (c *LiveCard) Approve(ctx *via.Ctx) {
	key := c.Key
	if key == "" {
		key = defaultSessionKey
	}
	cfg, log := readLiveState(key)
	if log == nil {
		return
	}
	land := pipe.LandState("")
	if e := lookupLiveEntry(key); e != nil {
		land = pipe.LandState(e.landState())
	}
	override := landOverridden(c.LandOverride.Read(ctx))
	if blocked, reason := landBlocked(len(sessionOpenThreads(key)), land); blocked && !override {
		setLandResult(key, "blocked — "+reason)
		c.Landed.Write(ctx, "blocked")
		return
	}
	title, body := prTitleAndBody(key, latestDispatchPrompt(log))
	url, err := openPR(context.Background(), cfg.RepoDir, prBranchName(key), title, body)
	if err != nil {
		setLandResult(key, "PR failed — "+err.Error())
		c.Landed.Write(ctx, "error")
		return
	}
	setLandResult(key, url)
	c.Landed.Write(ctx, "opened")
}

// landOverridden reads the override signal tolerantly ("1" or "true").
func landOverridden(v string) bool {
	v = strings.TrimSpace(v)
	return v == "1" || v == "true"
}

// setLandResult caches the approve outcome on the session entry (no-op if unknown).
func setLandResult(key, res string) {
	if e := lookupLiveEntry(key); e != nil {
		e.setLandResult(res)
	}
}

// latestDispatchPrompt returns the most recent dispatched order's prompt — the task
// the PR title/body summarizes. Empty when there are no dispatches.
func latestDispatchPrompt(log *ledger.Log) string {
	if log == nil {
		return ""
	}
	views, err := log.RecentDispatches(1)
	if err != nil || len(views) == 0 {
		return ""
	}
	return views[0].Target.Prompt
}

// renderLandControl renders the approve-and-open-PR control: a button wired to
// Approve, an override toggle (to push past the guard deliberately), and the last
// approve outcome (the opened PR URL, a guard message, or a failure) when present.
func renderLandControl(c *LiveCard) h.H {
	key := c.Key
	if key == "" {
		key = defaultSessionKey
	}
	parts := []h.H{
		h.Class("land-control"),
		h.Button(on.Click(c.Approve), h.Class("pk-btn land-control__approve"), h.Text("Approve & open PR")),
		h.Label(h.Class("land-control__override"),
			h.Input(h.Type("checkbox"), c.LandOverride.Bind()),
			h.Span(h.Text("override guard")),
		),
	}
	if res := landResultSnapshot(key); res != "" {
		parts = append(parts, h.Span(h.Class("land-control__result"), h.Text(res)))
	}
	return h.Div(parts...)
}

// sessionHasDispatches reports whether the session has at least one dispatched order —
// the landable work the approve flow opens a PR for.
func sessionHasDispatches(log *ledger.Log) bool {
	if log == nil {
		return false
	}
	v, err := log.RecentDispatches(1)
	return err == nil && len(v) > 0
}

// landResultSnapshot returns the cached approve outcome for a session ("" if none).
func landResultSnapshot(key string) string {
	if e := lookupLiveEntry(key); e != nil {
		return e.landResultSnapshot()
	}
	return ""
}
