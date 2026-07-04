package app

import (
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/go-via/via/h"

	"github.com/joaomdsg/packets/internal/packet"
)

// consoleNeedsYouCap bounds the needs-you rail to a handful of full cards
// (design/ui_kits/console/ConsoleScreen.jsx's paged queue) — the rest collapse
// into a single "and N more" line so a drowning session stays scannable.
const consoleNeedsYouCap = 4

// consoleSettledCap bounds the settled rail to the "last ~5" the slice spec
// calls for.
const consoleSettledCap = 5

// consoleIntentWords bounds the in-flight strip's intent preview to a
// scannable clause rather than a full paragraph (ROADMAP slice 6).
const consoleIntentWords = 8

// renderConsole assembles the "/" Console shell around the PRESERVED center
// content: a needs-you rail of this session's HELD packets, the center column
// (a hero stat + in-flight strip above the untouched act-now/state-history
// sections), and a settled+watches rail. Every region folds from the SAME
// packets slice (ROADMAP slice 6's packet aggregate) — one source of truth,
// never a fabricated placeholder.
func renderConsole(navKey string, packets []packet.Packet, addr packet.Addr, center h.H) h.H {
	verified := 0
	for _, p := range packets {
		if p.State == packet.Verified {
			verified++
		}
	}
	mainParts := []h.H{h.Class("console__main"), renderHeroStat(verified, addr)}
	if strip := renderInFlightStrip(packets); strip != nil {
		mainParts = append(mainParts, strip)
	}
	mainParts = append(mainParts, center)

	return h.Div(
		h.Class("console"),
		renderNeedsYouRail(navKey, packets),
		h.Div(mainParts...),
		renderSettledRail(packets),
	)
}

// heldPackets filters to packets a human must look at (Hold != HoldNone),
// blocking packets first — the most attention-worthy lead the queue. A stable
// sort keeps each hold kind's own relative (dispatch-recency) order intact.
func heldPackets(packets []packet.Packet) []packet.Packet {
	var held []packet.Packet
	for _, p := range packets {
		if p.Hold != packet.HoldNone {
			held = append(held, p)
		}
	}
	sort.SliceStable(held, func(i, j int) bool {
		return held[i].Hold == packet.HoldBlocking && held[j].Hold != packet.HoldBlocking
	})
	return held
}

// renderNeedsYouRail is the left rail: one card per held packet (capped at
// consoleNeedsYouCap, "+N more" beyond that), each linking straight into the
// packet's own review — or the victory empty state when nothing is held. A
// dashed "calibration" placeholder sits below it: the draw mechanic (slice 11)
// doesn't exist yet, so it names that honestly rather than faking a sample.
func renderNeedsYouRail(navKey string, packets []packet.Packet) h.H {
	held := heldPackets(packets)
	header := h.Div(h.Class("console__panel-header"), h.Text("needs you · "+strconv.Itoa(len(held))))

	body := []h.H{h.Class("console__rail-body")}
	if len(held) == 0 {
		body = append(body, h.Div(
			h.Class("console__card", "console__card--dashed"),
			h.Text("nothing needs you"),
		))
	} else {
		shown := held
		more := 0
		if len(shown) > consoleNeedsYouCap {
			more = len(shown) - consoleNeedsYouCap
			shown = shown[:consoleNeedsYouCap]
		}
		for _, p := range shown {
			body = append(body, renderNeedsYouCard(navKey, p))
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

// renderNeedsYouCard renders one held packet as a card: an 8px state cell
// (blocking/advisory hued), the packet's Name, its one-clause HoldReason in
// mono microtype, and a trailing "inspect →" naming the destination. The
// whole card is the link into the packet's own review.
func renderNeedsYouCard(navKey string, p packet.Packet) h.H {
	href := "/review?key=" + url.QueryEscape(navKey) + "&wo=" + strconv.Itoa(p.ID)
	state := "held"
	if p.Hold == packet.HoldBlocking {
		state = "held-blocking"
	}
	return h.A(
		h.Href(href),
		h.Class("console__card"),
		h.Span(h.Class("console__cell"), h.Data("state", state)),
		h.Div(h.Class("console__thread-title"), h.Text(p.Name)),
		h.Div(h.Class("console__thread-loc"), h.Text(p.HoldReason)),
		h.Div(h.Class("console__thread-arrow"), h.Text("inspect →")),
	)
}

// renderHeroStat is the center column's header strip: the count of packets
// whose State is Verified (done, the order's own catch minted, no open
// questions) — the only forward state the gauntlet actually proves today
// (MVP.md vocabulary map: "forwarded"/"delivered" aren't mechanized yet, so
// neither word belongs here). The addr line beside it is the honest repo
// identity (packet.ParseAddr), never a fabricated owner.
func renderHeroStat(verifiedCount int, addr packet.Addr) h.H {
	return h.Div(h.Class("console__hero"),
		h.Span(h.Class("console__hero-stat"), h.Text(strconv.Itoa(verifiedCount))),
		h.Span(h.Class("console__hero-label"), h.Text("packets verified")),
		h.Span(h.Class("console__hero-addr"), h.Text("addr "+addr.String())),
	)
}

// renderInFlightStrip is the center column's "in flight · N" strip: one
// pulsing --signal cell per InFlight packet (with a first-words preview of its
// intent) and one ghost-outline cell per Composing packet — the mark's
// composing idiom reused for a queued order. Returns nil (omitted entirely)
// when nothing is in flight or composing, rather than an empty header.
func renderInFlightStrip(packets []packet.Packet) h.H {
	var inFlight, composing []packet.Packet
	for _, p := range packets {
		switch p.State {
		case packet.InFlight:
			inFlight = append(inFlight, p)
		case packet.Composing:
			composing = append(composing, p)
		}
	}
	n := len(inFlight) + len(composing)
	if n == 0 {
		return nil
	}

	rows := []h.H{h.Div(h.Class("console__panel-header"), h.Text("in flight · "+strconv.Itoa(n)))}
	for _, p := range inFlight {
		rows = append(rows, h.Div(h.Class("console__inflight-row"),
			h.Span(h.Class("console__cell"), h.Data("state", "in-flight")),
			h.Span(h.Class("console__inflight-name"), h.Text(p.Name)),
			h.Span(h.Class("console__inflight-intent"), h.Text(firstWords(p.Intent, consoleIntentWords))),
		))
	}
	for _, p := range composing {
		rows = append(rows, h.Div(h.Class("console__inflight-row"),
			h.Span(h.Class("console__cell"), h.Data("state", "composing")),
			h.Span(h.Class("console__inflight-name"), h.Text(p.Name)),
		))
	}
	return h.Div(append([]h.H{h.Class("console__inflight")}, rows...)...)
}

// firstWords returns the first n whitespace-separated words of s — a
// scannable clause rather than a full paragraph. No ellipsis marker; a
// truncated clause reads fine without one in this dense mono row.
func firstWords(s string, n int) string {
	words := strings.Fields(s)
	if len(words) > n {
		words = words[:n]
	}
	return strings.Join(words, " ")
}

// renderSettledRail is the right rail: the honest replacement for a fake
// "recently delivered" (delivered isn't real until ACK — slice 13). It lists
// verified and held packets — a run failure IS settled-red, only
// composing/in-flight packets are excluded — each a lifecycle-colored state
// square + Name + one-word state, or the dashed empty state when nothing has
// settled. Below it, "your watches" is a dashed empty placeholder — the
// mechanic (slice 12) doesn't exist yet.
func renderSettledRail(packets []packet.Packet) h.H {
	var settled []packet.Packet
	for _, p := range packets {
		if p.State == packet.Verified || p.State == packet.Held {
			settled = append(settled, p)
		}
	}
	// The header counts every settled packet — captured BEFORE the display cap
	// below, so a caller that ever folds more than consoleSettledCap packets
	// still gets an honest count, not a silently truncated one.
	count := len(settled)
	if len(settled) > consoleSettledCap {
		settled = settled[:consoleSettledCap]
	}

	header := h.Div(h.Class("console__panel-header"), h.Text("settled · "+strconv.Itoa(count)))
	body := []h.H{h.Class("console__rail-body")}
	if len(settled) == 0 {
		body = append(body, h.Div(h.Class("console__card", "console__card--dashed"), h.Text("nothing settled yet")))
	} else {
		for _, p := range settled {
			body = append(body, renderSettledRow(p))
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

// renderSettledRow renders one settled packet: a state square colored by
// State/Hold (verified solid, held advisory amber, held blocking red), its
// Name, and its lifecycle State.String() ("verified" or "held") — the same
// one-word vocabulary the needs-you rail and the packet aggregate already use,
// never a second name for the same fact.
func renderSettledRow(p packet.Packet) h.H {
	state := "verified"
	if p.State == packet.Held {
		state = "held"
		if p.Hold == packet.HoldBlocking {
			state = "held-blocking"
		}
	}
	return h.Div(h.Class("console__card", "console__settled-row"),
		h.Span(h.Class("console__cell"), h.Data("state", state)),
		h.Span(h.Class("console__settled-id"), h.Text(p.Name)),
		h.Span(h.Class("console__settled-outcome"), h.Text(p.State.String())),
	)
}
