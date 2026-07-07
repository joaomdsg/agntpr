package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joaomdsg/packets/internal/assist"
)

// lineOfOffset maps a byte offset in the draft to its 1-based line, so a flag can
// name WHERE it is — the "where" the Lead reads to find the span.
func TestLineOfOffset_countsNewlinesBeforeTheOffset(t *testing.T) {
	draft := "line one\nline two\nline three"
	assert.Equal(t, 1, lineOfOffset(draft, 0), "offset 0 is line 1")
	assert.Equal(t, 1, lineOfOffset(draft, 3), "before the first newline is still line 1")
	assert.Equal(t, 2, lineOfOffset(draft, 9), "just past the first newline is line 2")
	assert.Equal(t, 3, lineOfOffset(draft, 18), "past the second newline is line 3")
}

// A negative or past-the-end offset clamps to a real line — never line 0, never a
// crash, so a stale/garbled offset can't produce a nonsense location.
func TestLineOfOffset_clampsOutOfRangeToARealLine(t *testing.T) {
	draft := "a\nb\nc"
	assert.Equal(t, 1, lineOfOffset(draft, -5), "a negative offset is line 1, not line 0")
	assert.Equal(t, 3, lineOfOffset(draft, 999), "past the end clamps to the last line")
}

const flagDraft = "package main\n\nfunc main() {}\n\n// more lines here for offsets\n"

// The producer's flagged spans surface as readable cards, not only as inline
// editor decorations a Lead must hover to read — each card carries the real note,
// its severity, and its line location.
func TestRenderHighlightCards_showsEachFlagsNoteSeverityAndLine(t *testing.T) {
	cards := renderHighlightCards(flagDraft, []assist.Highlight{
		{Start: 0, End: 6, Note: "vague: what is a legitimate burst?", Severity: "question"},
		{Start: 14, End: 20, Note: "no recovery path specified", Severity: "gap"},
	})
	body := renderHTMLNodes(t, cards)
	assert.Equal(t, 2, strings.Count(body, "analysis__flag\""), "one card per flagged span")
	assert.Contains(t, body, "vague: what is a legitimate burst?")
	assert.Contains(t, body, "question", "the flag's severity tags it")
	assert.Contains(t, body, "no recovery path specified")
	assert.Contains(t, body, "gap")
	assert.Contains(t, body, "line 1", "the first flag names its line (offset 0)")
	assert.Contains(t, body, "line 3", "the second flag names its line (offset 14 is past two newlines)")
}

// A flag with an empty OR whitespace-only note is nothing to read — skipped
// rather than shown as a blank card (matching how the rest of the assist layer
// treats emptiness, via trim).
func TestRenderHighlightCards_skipsAFlagWithNoReadableNote(t *testing.T) {
	cards := renderHighlightCards(flagDraft, []assist.Highlight{
		{Start: 0, End: 6, Note: "", Severity: "note"},
		{Start: 6, End: 8, Note: "   ", Severity: "note"},
		{Start: 8, End: 9, Note: "real note", Severity: "note"},
	})
	body := renderHTMLNodes(t, cards)
	assert.Equal(t, 1, strings.Count(body, "analysis__flag\""), "only the flag with something to say renders")
	assert.Contains(t, body, "real note")
}

// A flag with no severity still shows its note and line — the tag is optional —
// and NO empty severity tag is emitted.
func TestRenderHighlightCards_rendersANoteWithoutAStraySeverityTag(t *testing.T) {
	cards := renderHighlightCards(flagDraft, []assist.Highlight{{Start: 0, End: 3, Note: "untagged but real"}})
	body := renderHTMLNodes(t, cards)
	assert.Equal(t, 1, strings.Count(body, "analysis__flag\""))
	assert.Contains(t, body, "untagged but real")
	assert.Contains(t, body, "line 1", "even an untagged flag names where it is")
	assert.NotContains(t, body, "analysis__flag-severity", "no severity → no empty tag element")
}

// No flags → no cards; a clean draft's panel shows none.
func TestRenderHighlightCards_isEmptyForNoFlags(t *testing.T) {
	assert.Empty(t, renderHighlightCards(flagDraft, nil))
}

// The analysis panel surfaces the producer's flagged spans as cards beside its
// summary and questions — the built composer's take on the exploration's
// harness-pair panel, from real producer output only.
func TestAnalysisPanel_surfacesFlaggedSpansAsCards(t *testing.T) {
	da := &draftAnalysis{Draft: flagDraft, Result: &assist.Analysis{
		Summary:    "s",
		Highlights: []assist.Highlight{{Start: 0, End: 6, Note: "ambiguous goal", Severity: "question"}},
	}}
	body := renderHTML(t, renderAnalysisPanel(da))
	assert.Contains(t, body, "analysis__flag", "flagged spans render as cards in the panel")
	assert.Contains(t, body, "ambiguous goal")
	assert.Contains(t, body, "line 1", "the flag names its location, computed from the real draft")
}
