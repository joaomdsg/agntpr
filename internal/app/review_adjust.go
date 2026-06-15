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
	// Remember the anchor so a later render can relocate it against the settled revision
	// and show whether the agent addressed it (DESIGN §28 thin slice). Only when we could
	// read the line — an unquotable anchor has nothing to relocate by content.
	if codeLine != "" {
		if e := lookupLiveEntry(key); e != nil {
			e.setAdjAnchor(file, line, codeLine)
		}
	}
	tgt := ledger.Target{BaseRev: head, Prompt: assist.ReviewTurnPrompt(file, line, codeLine, text)}
	// Funded by attention bandwidth like any UI-authored live order; an over-budget
	// meter is refused by the ledger and is a silent no-op.
	if err := log.AppendLiveDispatch("liveorder", tgt, ownTargetOf(cfg)); err != nil {
		return
	}
	go drainQueuedOrders(key)
}

// adjAnchorState is where an adjustment's commented line ended up after the agent
// settled a new revision — the thin content-relocation slice of DESIGN §28.
type adjAnchorState int

const (
	adjSame     adjAnchorState = iota // still on the line the Lead commented
	adjMoved                          // the same line survived but shifted to a new number
	adjOutdated                       // the line was edited/deleted — the agent addressed it
)

// adjReanchor is the relocation outcome: the state plus the 1-based line the anchor now
// sits on (0 when outdated — there is no surviving line to point at).
type adjReanchor struct {
	State adjAnchorState
	Line  int
}

// adjAnchorRecord is the LAST adjustment's anchor cached on a liveEntry: the file:line
// the Lead commented and the content of that line at comment time, against which a later
// revision is relocated (reanchorAdjustment).
type adjAnchorRecord struct {
	file    string
	line    int
	content string
}

// reanchorAdjustment relocates an adjustment's commented line against a new revision's
// file content by EXACT line-content match (DESIGN §28, thin slice — no git-hunk rebase
// or rename tracking). An empty anchor has nothing to relocate. The original line still
// at its old number is Same; the same content found elsewhere is Moved to the first such
// line; content gone entirely is Outdated. A trailing "\r" is trimmed on both sides so a
// \r\n revision doesn't spuriously read as edited.
func reanchorAdjustment(origLine string, origLineNum int, newContent string) adjReanchor {
	if origLine == "" {
		return adjReanchor{State: adjOutdated}
	}
	want := strings.TrimSuffix(origLine, "\r")
	lines := strings.Split(newContent, "\n")
	if origLineNum >= 1 && origLineNum <= len(lines) &&
		strings.TrimSuffix(lines[origLineNum-1], "\r") == want {
		return adjReanchor{State: adjSame, Line: origLineNum}
	}
	for i, ln := range lines {
		if strings.TrimSuffix(ln, "\r") == want {
			return adjReanchor{State: adjMoved, Line: i + 1}
		}
	}
	return adjReanchor{State: adjOutdated}
}

// renderAdjustmentForm renders the review surface's adjustment entry point: file +
// line + comment inputs (bound to the adjustment signals) and a button wired to
// AddAdjustment. This is the UI half of the comment→harness round-trip — the Lead
// tells the agent what to change and the live harness re-edits in place.
func renderAdjustmentForm(c *ReviewCard) h.H {
	parts := []h.H{
		h.Class("review-adjust"),
		h.P(h.Class("review-adjust__label"), h.Text("Leave an adjustment — tell the agent what to change:")),
		h.Input(h.Type("text"), c.AdjFile.Bind(), h.Class("pk-input review-adjust__file"), h.Placeholder("file (e.g. main.go)")),
		h.Input(h.Type("text"), c.AdjLine.Bind(), h.Class("pk-input review-adjust__line"), h.Placeholder("line"), h.Attr("aria-label", "line number")),
		h.Input(h.Type("text"), c.AdjText.Bind(), h.Class("pk-input review-adjust__text"), h.Placeholder("what should change?")),
		h.Button(on.Click(c.AddAdjustment), h.Class("pk-btn review-adjust__submit"), h.Text("Leave adjustment")),
	}
	if badge := renderAdjustmentStatus(c); badge != nil {
		parts = append(parts, badge)
	}
	return h.Div(parts...)
}

// renderAdjustmentStatus shows whether the LAST adjustment was addressed: it relocates
// the cached anchor against the file's CURRENT content (the settled revision on disk) and
// renders a badge — still-here / moved / line-edited — so "leave an adjustment → watch it
// addressed" has a visible payoff (DESIGN §28 thin slice). Returns nil when no adjustment
// was left this session (nothing to report).
func renderAdjustmentStatus(c *ReviewCard) h.H {
	key := c.Key
	if key == "" {
		key = defaultSessionKey
	}
	e := lookupLiveEntry(key)
	if e == nil {
		return nil
	}
	anchor := e.adjAnchorSnapshot()
	if anchor == nil {
		return nil
	}
	cfg, _ := readLiveState(key)
	content := ""
	if data, err := os.ReadFile(filepath.Join(cfg.RepoDir, anchor.file)); err == nil {
		content = string(data)
	}
	r := reanchorAdjustment(anchor.content, anchor.line, content)
	var cls, text string
	switch r.State {
	case adjSame:
		cls, text = "review-adjust__status--same", "still on line "+strconv.Itoa(r.Line)
	case adjMoved:
		cls, text = "review-adjust__status--moved", "addressed — moved to line "+strconv.Itoa(r.Line)
	default: // adjOutdated
		cls, text = "review-adjust__status--outdated", "addressed — line edited"
	}
	return h.Span(h.Class("review-adjust__status "+cls), h.Text(text))
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
