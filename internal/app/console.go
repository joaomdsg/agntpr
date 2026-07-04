package app

import (
	"net/url"
	"strconv"

	"github.com/go-via/via/h"

	"github.com/joaomdsg/packets/internal/ledger"
	"github.com/joaomdsg/packets/internal/review"
)

// consoleNeedsYouCap bounds the needs-you rail to a handful of full cards
// (design/ui_kits/console/ConsoleScreen.jsx's paged queue) — the rest collapse
// into a single "and N more" line so a drowning session stays scannable.
const consoleNeedsYouCap = 4

// consoleSettledCap bounds the settled rail to the "last ~5" the slice spec
// calls for; RecentDispatches(n) already caps supply to n, but a caller may
// pass a larger slice, so this filters defensively.
const consoleSettledCap = 5

// renderConsole assembles the "/" Console shell (ROADMAP slice 3) around the
// PRESERVED center content: a needs-you rail of open review threads, the
// center column (a hero stat above the untouched act-now/state-history
// sections), and a settled+watches rail. Layout only — every number renders
// from a real accessor (navKey's open threads, the ledger's dispatch tally),
// never a fabricated placeholder.
func renderConsole(navKey string, threads []review.Thread, verifiedCount int, settled []ledger.DispatchView, center h.H) h.H {
	return h.Div(
		h.Class("console"),
		renderNeedsYouRail(navKey, threads),
		h.Div(h.Class("console__main"),
			renderHeroStat(verifiedCount),
			center,
		),
		renderSettledRail(settled),
	)
}

// renderNeedsYouRail is the left rail: one card per open review thread
// (capped at consoleNeedsYouCap, "+N more" beyond that), each linking straight
// into the thread's session review — or the victory empty state when there is
// nothing open. A dashed "calibration" placeholder sits below it: the draw
// mechanic (slice 11) doesn't exist yet, so it names that honestly rather than
// faking a sample.
func renderNeedsYouRail(navKey string, threads []review.Thread) h.H {
	header := h.Div(h.Class("console__panel-header"), h.Text("needs you · "+strconv.Itoa(len(threads))))

	body := []h.H{h.Class("console__rail-body")}
	if len(threads) == 0 {
		body = append(body, h.Div(
			h.Class("console__card", "console__card--dashed"),
			h.Text("nothing needs you"),
		))
	} else {
		shown := threads
		more := 0
		if len(shown) > consoleNeedsYouCap {
			more = len(shown) - consoleNeedsYouCap
			shown = shown[:consoleNeedsYouCap]
		}
		href := "/review?key=" + url.QueryEscape(navKey)
		for _, t := range shown {
			body = append(body, renderNeedsYouCard(href, t))
		}
		if more > 0 {
			body = append(body, h.Div(h.Class("console__more"), h.Text("and "+strconv.Itoa(more)+" more")))
		}
	}
	body = append(body, h.Div(
		h.Class("console__card", "console__card--dashed"),
		h.Div(h.Class("console__empty-kicker"), h.Text("calibration")),
		h.Div(h.Text("no calibration draws yet")),
	))

	return h.Aside(h.Class("console__rail", "console__rail--needs-you"),
		header,
		h.Div(body...),
	)
}

// renderNeedsYouCard renders one open review thread as a card: its body text
// as the title, its file:line as the anchor (when the thread carries one),
// and a trailing "inspect →" naming the destination per the house voice. The
// whole card is the link.
func renderNeedsYouCard(href string, t review.Thread) h.H {
	title := t.Body
	if title == "" {
		title = t.Render()
	}
	card := []h.H{
		h.Href(href),
		h.Class("console__card"),
		h.Div(h.Class("console__thread-title"), h.Text(title)),
	}
	if t.File != "" && t.StartLine > 0 {
		card = append(card, h.Div(h.Class("console__thread-loc"), h.Text(t.File+":"+strconv.Itoa(t.StartLine))))
	}
	card = append(card, h.Div(h.Class("console__thread-arrow"), h.Text("inspect →")))
	return h.A(card...)
}

// renderHeroStat is the center column's header strip: the tabular DONE count
// from the ledger's dispatch tally, labelled "packets verified" — the only
// forward state the gauntlet actually proves today (MVP.md vocabulary map:
// "forwarded"/"delivered" aren't mechanized yet, so neither word belongs here).
func renderHeroStat(verifiedCount int) h.H {
	return h.Div(h.Class("console__hero"),
		h.Span(h.Class("console__hero-stat"), h.Text(strconv.Itoa(verifiedCount))),
		h.Span(h.Class("console__hero-label"), h.Text("packets verified")),
	)
}

// renderSettledRail is the right rail: the honest replacement for a fake
// "recently delivered" (delivered isn't real until ACK — slice 13). It lists
// the session's recent DONE dispatches, each a state square (verified when the
// order minted its own catch, held otherwise) + id + one-word outcome, or the
// dashed empty state when nothing has settled. Below it, "your watches" is a
// dashed empty placeholder — the mechanic (slice 12) doesn't exist yet.
func renderSettledRail(dispatches []ledger.DispatchView) h.H {
	done := make([]ledger.DispatchView, 0, len(dispatches))
	for _, d := range dispatches {
		if d.Status == "done" {
			done = append(done, d)
		}
	}
	// The header counts every settled order — captured BEFORE the display cap
	// below, so a caller that ever hands this more than consoleSettledCap
	// dispatches still gets an honest count, not a silently truncated one.
	count := len(done)
	if len(done) > consoleSettledCap {
		done = done[:consoleSettledCap]
	}

	header := h.Div(h.Class("console__panel-header"), h.Text("settled · "+strconv.Itoa(count)))
	body := []h.H{h.Class("console__rail-body")}
	if len(done) == 0 {
		body = append(body, h.Div(h.Class("console__card", "console__card--dashed"), h.Text("nothing settled yet")))
	} else {
		for _, d := range done {
			body = append(body, renderSettledRow(d))
		}
	}

	watchesHeader := h.Div(h.Class("console__panel-header"), h.Text("your watches"))
	watchesBody := h.Div(h.Class("console__rail-body"),
		h.Div(h.Class("console__card", "console__card--dashed"), h.Text("no watches yet")),
	)

	return h.Aside(h.Class("console__rail", "console__rail--settled"),
		header, h.Div(body...),
		watchesHeader, watchesBody,
	)
}

// renderSettledRow renders one settled dispatch: an 8px state square
// (verified when it minted its own catch, held otherwise), its id, and a
// one-word outcome — mirroring the vocabulary the preserved center column
// already uses for the same dispatch (caught/missed), never a second name for
// the same fact.
func renderSettledRow(d ledger.DispatchView) h.H {
	state := "held"
	outcome := "missed"
	if d.Caught {
		state = "verified"
		outcome = "caught"
	}
	return h.Div(h.Class("console__card", "console__settled-row"),
		h.Span(h.Class("console__cell"), h.Data("state", state)),
		h.Span(h.Class("console__settled-id"), h.Text("wo#"+strconv.Itoa(d.ID))),
		h.Span(h.Class("console__settled-outcome"), h.Text(outcome)),
	)
}
