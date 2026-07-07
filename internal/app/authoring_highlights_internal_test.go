package app

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joaomdsg/packets/internal/assist"
)

// The producer's flagged spans surface as readable cards, not only as inline
// editor decorations a Lead must hover to read — each card carries the real note
// and its severity, so what the producer flagged is legible at a glance.
func TestRenderHighlightCards_showsEachFlagsNoteAndSeverity(t *testing.T) {
	cards := renderHighlightCards([]assist.Highlight{
		{Start: 0, End: 6, Note: "vague: what is a legitimate burst?", Severity: "question"},
		{Start: 10, End: 20, Note: "no recovery path specified", Severity: "gap"},
	})
	body := renderHTMLNodes(t, cards)
	assert.Equal(t, 2, strings.Count(body, "analysis__flag\""), "one card per flagged span")
	assert.Contains(t, body, "vague: what is a legitimate burst?")
	assert.Contains(t, body, "question", "the flag's severity tags it")
	assert.Contains(t, body, "no recovery path specified")
	assert.Contains(t, body, "gap")
}

// A flag with an empty OR whitespace-only note is nothing to read — skipped
// rather than shown as a blank card (matching how the rest of the assist layer
// treats emptiness, via trim).
func TestRenderHighlightCards_skipsAFlagWithNoReadableNote(t *testing.T) {
	cards := renderHighlightCards([]assist.Highlight{
		{Start: 0, End: 6, Note: "", Severity: "note"},
		{Start: 6, End: 8, Note: "   ", Severity: "note"},
		{Start: 8, End: 9, Note: "real note", Severity: "note"},
	})
	body := renderHTMLNodes(t, cards)
	assert.Equal(t, 1, strings.Count(body, "analysis__flag\""), "only the flag with something to say renders")
	assert.Contains(t, body, "real note")
}

// A flag with no severity still shows its note — the tag is optional, the note is
// the point — and NO empty severity tag is emitted.
func TestRenderHighlightCards_rendersANoteWithoutAStraySeverityTag(t *testing.T) {
	cards := renderHighlightCards([]assist.Highlight{{Start: 0, End: 3, Note: "untagged but real"}})
	body := renderHTMLNodes(t, cards)
	assert.Equal(t, 1, strings.Count(body, "analysis__flag\""))
	assert.Contains(t, body, "untagged but real")
	assert.NotContains(t, body, "analysis__flag-severity", "no severity → no empty tag element")
}

// No flags → no cards; a clean draft's panel shows none.
func TestRenderHighlightCards_isEmptyForNoFlags(t *testing.T) {
	assert.Empty(t, renderHighlightCards(nil))
}

// The analysis panel surfaces the producer's flagged spans as cards beside its
// summary and questions — the built composer's take on the exploration's
// harness-pair panel, from real producer output only.
func TestAnalysisPanel_surfacesFlaggedSpansAsCards(t *testing.T) {
	da := &draftAnalysis{Draft: "d", Result: &assist.Analysis{
		Summary:    "s",
		Highlights: []assist.Highlight{{Start: 0, End: 6, Note: "ambiguous goal", Severity: "question"}},
	}}
	body := renderHTML(t, renderAnalysisPanel(da))
	assert.Contains(t, body, "analysis__flag", "flagged spans render as cards in the panel")
	assert.Contains(t, body, "ambiguous goal")
}
