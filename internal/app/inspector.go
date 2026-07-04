package app

import (
	"strconv"

	"github.com/go-via/via/h"

	"github.com/joaomdsg/packets/internal/packet"
	"github.com/joaomdsg/packets/internal/review"
)

// shortRev truncates rev to the short-SHA convention (7 characters); a rev at
// or under that length renders as-is — never padded or fabricated.
func shortRev(rev string) string {
	if len(rev) > 7 {
		return rev[:7]
	}
	return rev
}

// renderInspectorTitlebar renders the Inspector's identity strip: the scope
// name (wo#<id> or the raw session key — honest, never a display alias), the
// packet's own Name beside it when the scope folds to one (packetName ""
// omits it), a base→fix rev chip in short SHAs (omitted entirely when either
// rev is unknown — the Inspector never fabricates a revision), a neutral lane
// chip (ROADMAP slice 7 — lane is QoS, never a state color; LaneUnmeasured
// when the scope folds to no single packet or nothing has been measured yet,
// rendered honestly rather than omitted), and the repo addr (packet.ParseAddr
// — owner/name, or the honest local/<dir> fallback). navHeader already
// carries the brand mark + lockup, so this strip adds no second mark (MVP.md
// Titlebar pattern).
func renderInspectorTitlebar(scope, baseRev, fixRev string, addr packet.Addr, packetName string, lane packet.Lane) h.H {
	parts := []h.H{
		h.Class("inspector__titlebar"),
		h.Span(h.Class("inspector__name"), h.Text(scope)),
	}
	if packetName != "" {
		parts = append(parts, h.Span(h.Class("inspector__packet-name"), h.Text(packetName)))
	}
	if baseRev != "" && fixRev != "" {
		parts = append(parts, h.Span(h.Class("inspector__rev"),
			h.Text(shortRev(baseRev)+"→"+shortRev(fixRev))))
	}
	parts = append(parts, h.Span(h.Class("inspector__lane"), h.Text("lane "+lane.String())))
	if addr.Name != "" {
		parts = append(parts, h.Span(h.Class("inspector__addr"), h.Text(addr.String())))
	}
	return h.Div(parts...)
}

// renderInspectorGrid assembles the 3-column Inspector body (252px|1fr|312px,
// MVP.md): the changed-files tree, the Monaco island + answer form, and the
// annotation rail — each hairline-bounded, mirroring the Console shell's
// region approach.
func renderInspectorGrid(left, main h.H, rail []h.H) h.H {
	return h.Div(
		h.Class("inspector"),
		h.Div(h.Class("inspector__tree"), left),
		h.Div(h.Class("inspector__main"), main),
		h.Aside(append([]h.H{h.Class("inspector__rail")}, rail...)...),
	)
}

// renderInspectorEmptyTree is the honest left-rail empty state for a review
// that isn't scoped to one work-order: there is no single base→fix diff to
// scope a tree to, so the Inspector says so rather than showing an arbitrary
// tree — the open threads still list on the annotation rail.
func renderInspectorEmptyTree() h.H {
	return h.Div(h.Class("inspector__tree-empty"),
		h.Text("pick a packet to scope the file tree"))
}

// renderInspectorTimeline is the Inspector's full-width footer: an honest
// dashed empty. The replayable packet-life timeline (MVP.md concept 2, 10)
// is folded from the packet aggregate that lands in slice 5 — never faked
// here.
func renderInspectorTimeline() h.H {
	return h.Div(h.Class("inspector__timeline"),
		h.Div(h.Class("inspector__timeline-kicker"), h.Text("timeline")),
		h.Div(h.Text("no replayable timeline yet")),
	)
}

// annotationRailHeader renders the rail's kicker in the house voice
// ("label · N" counts, MVP.md vocabulary map).
func annotationRailHeader(n int) h.H {
	return h.Div(h.Class("inspector__rail-header"), h.Text("annotations · "+strconv.Itoa(n)))
}

// renderAnnotationCard renders one open thread as an Inspector annotation
// card (MVP.md AnnotationCard spec): an "agent" author chip (every open
// thread here is an oracle finding, never user-authored), a severity chip
// carrying the thread's Conventional-Comment tag, the file:line anchor in
// mono, and the body as prose. It KEEPS the "review-thread" class and the
// data-file/data-line attributes the answer-form anchor flow and the editor
// island payload depend on — the annotation-card classes are additive, never
// a replacement.
func renderAnnotationCard(t review.Thread) h.H {
	return h.Div(
		h.Class("review-thread annotation-card"),
		h.Data("file", t.File),
		h.Data("line", strconv.Itoa(t.StartLine)),
		h.Div(h.Class("annotation-card__head"),
			h.Span(h.Class("annotation-card__chip annotation-card__chip--author"), h.Text("agent")),
			h.Span(h.Class("annotation-card__chip annotation-card__chip--sev"), h.Text(t.Tag)),
			h.Span(h.Class("review-thread__anchor annotation-card__where"),
				h.Text(t.File+":"+strconv.Itoa(t.StartLine))),
		),
		h.Span(h.Class("review-thread__body annotation-card__body"), h.Text(t.Render())),
	)
}

// renderAnnotationCards renders every thread as an annotation card, in order.
func renderAnnotationCards(threads []review.Thread) []h.H {
	out := make([]h.H, 0, len(threads))
	for _, t := range threads {
		out = append(out, renderAnnotationCard(t))
	}
	return out
}
