package app

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-via/via"
	"github.com/go-via/via/h"
	"github.com/go-via/via/on"

	"github.com/joaomdsg/packets/internal/assist"
	"github.com/joaomdsg/packets/internal/ledger"
)

// AddAdjustment is the keystone of the review-thread loop (DESIGN §12.3): leaving an
// anchored comment on a line dispatches a harness TURN that tells the agent what to
// fix, run against the session's live HEAD — so "leave an adjustment" becomes "watch
// the agent fix it". It is an additive reuse of the live-order pipe: the comment turn
// is dispatched and drained exactly like a placed order (funded by attention
// bandwidth), with the agent's edits folding into the next settled revision. An empty
// comment, a treeless session, or an over-budget meter is a silent no-op.
func (c *ReviewCard) AddAdjustment(ctx *via.Ctx) {
	key := c.Key
	if key == "" {
		key = defaultSessionKey
	}
	cfg, log := readLiveState(key)
	if log == nil {
		return
	}
	text := strings.TrimSpace(c.AdjText.Read(ctx))
	if text == "" {
		return // an empty comment is nothing to address
	}
	head, ok := repoHead(cfg.RepoDir)
	if !ok {
		return // no resolvable tree to route the comment against
	}
	file := strings.TrimSpace(c.AdjFile.Read(ctx))
	line, _ := strconv.Atoi(strings.TrimSpace(c.AdjLine.Read(ctx)))
	// Quote the line under review when we can read it (best-effort; a missing/unreadable
	// line just omits the quote — the file:line and the comment still carry the turn).
	codeLine := readSourceLine(cfg.RepoDir, file, line)
	tgt := ledger.Target{BaseRev: head, Prompt: assist.ReviewTurnPrompt(file, line, codeLine, text)}
	// Funded by attention bandwidth like any UI-authored live order; an over-budget
	// meter is refused by the ledger and is a silent no-op.
	if err := log.AppendLiveDispatch("liveorder", tgt, ownTargetOf(cfg)); err != nil {
		return
	}
	go drainQueuedOrders(key)
}

// renderAdjustmentForm renders the review surface's adjustment entry point: file +
// line + comment inputs (bound to the adjustment signals) and a button wired to
// AddAdjustment. This is the UI half of the comment→harness round-trip — the Lead
// tells the agent what to change and the live harness re-edits in place.
func renderAdjustmentForm(c *ReviewCard) h.H {
	return h.Div(h.Class("review-adjust"),
		h.P(h.Class("review-adjust__label"), h.Text("Leave an adjustment — tell the agent what to change:")),
		h.Input(h.Type("text"), c.AdjFile.Bind(), h.Class("pk-input review-adjust__file"), h.Placeholder("file (e.g. main.go)")),
		h.Input(h.Type("text"), c.AdjLine.Bind(), h.Class("pk-input review-adjust__line"), h.Placeholder("line"), h.Attr("aria-label", "line number")),
		h.Input(h.Type("text"), c.AdjText.Bind(), h.Class("pk-input review-adjust__text"), h.Placeholder("what should change?")),
		h.Button(on.Click(c.AddAdjustment), h.Class("pk-btn review-adjust__submit"), h.Text("Leave adjustment")),
	)
}

// readSourceLine returns the 1-indexed content of file's line within repoDir, so the
// review turn can quote the line under comment. Best-effort: a missing file, an
// unreadable tree, or an out-of-range/non-positive line degrades to "" (no quote)
// rather than erroring — the adjustment still dispatches, it just isn't quoted.
func readSourceLine(repoDir, file string, line int) string {
	if file == "" || line < 1 {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(repoDir, file))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	if line > len(lines) {
		return ""
	}
	return strings.TrimRight(lines[line-1], "\r")
}
