package packet

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/joaomdsg/packets/internal/ledger"
)

// Packet is the read-model aggregate: one work order, made legible for the
// design concept model. Every field derives from a real input (a
// ledger.DispatchView, the repo Addr, and an injected open-questions count) —
// nothing here is fabricated.
type Packet struct {
	ID            int
	Name          string
	Addr          Addr
	Intent        string
	BaseRev       string
	FixRev        string
	State         Lifecycle
	Hold          HoldKind
	HoldReason    string
	Caught        bool
	Verdict       string
	OpenQuestions int
	// Lane is the packet's measured QoS class. Fold NEVER computes it — Fold
	// is a pure data->data projection over ledger views, while a lane needs a
	// `go list` exec against the repo (Measure, lane_measure.go) — so every
	// folded Packet starts at the honest zero value, LaneUnmeasured, until
	// the app layer measures and attaches one (ROADMAP slice 7's laneFor).
	Lane Lane
}

// Deliverable always reports false: delivered is UNREACHABLE until a real
// ACK mechanic exists (slice 13, "packets deployed"/"packets regressed").
// Nothing in Fold can produce State==Delivered today — see the never-Delivered
// test pinning that invariant.
func (p Packet) Deliverable() bool {
	return false
}

// slugName derives a packet's Name from the order's own prompt, never
// invented: a lowercase hyphen slug of the first 3 whitespace-separated words,
// keeping only letters/digits within each word. A word that cleans to empty
// (pure punctuation/symbols) is SKIPPED, not fabricated — the slug still uses
// whatever of the first 3 words survives cleaning. Only when EVERY candidate
// word cleans to empty (or the prompt has none) does the name fall back to
// "wo-<ID>", since there is nothing honest left to slug.
func slugName(prompt string, id int) string {
	words := strings.Fields(prompt)
	if len(words) > 3 {
		words = words[:3]
	}

	var cleaned []string
	for _, w := range words {
		var b strings.Builder
		for _, r := range w {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				b.WriteRune(unicode.ToLower(r))
			}
		}
		if b.Len() > 0 {
			cleaned = append(cleaned, b.String())
		}
	}

	if len(cleaned) == 0 {
		return fmt.Sprintf("wo-%d", id)
	}
	return strings.Join(cleaned, "-")
}

// Fold projects ledger DispatchViews into Packets — the read-model aggregate
// for the design concept model. It is PURE data→data: no I/O, no ledger
// writes, no import of internal/app. openQuestions is the caller's injected
// lookup (the app's findings cache) for how many open review questions a
// given order left; this package stays ignorant of where that number comes
// from. Order identity is preserved 1:1 with views (same length, same order,
// matched by ID).
//
// The dispatch status maps onto lifecycle/hold BINDING per this table (fail
// toward attention, never silently Verified, on anything not explicitly
// listed here):
//
//	queued                                  → Composing, no hold
//	running                                 → InFlight, no hold
//	done, Caught, 0 open questions           → Verified, no hold
//	done, !Caught or open questions > 0      → Held, advisory
//	  (open-questions reason wins over the gap reason when both apply)
//	failed                                  → Held, blocking, "run failed"
//	anything else (unknown/future status)   → Held, blocking,
//	                                           "unknown state · <status>"
func Fold(views []ledger.DispatchView, addr Addr, openQuestions func(orderID int) int) []Packet {
	packets := make([]Packet, len(views))
	for i, v := range views {
		questions := openQuestions(v.ID)
		state, hold, reason := lifecycleFor(v.Status, v.Caught, questions)
		packets[i] = Packet{
			ID:            v.ID,
			Name:          slugName(v.Target.Prompt, v.ID),
			Addr:          addr,
			Intent:        v.Target.Prompt,
			BaseRev:       v.Target.BaseRev,
			FixRev:        v.Target.FixRev,
			Caught:        v.Caught,
			Verdict:       v.Verdict,
			OpenQuestions: questions,
			State:         state,
			Hold:          hold,
			HoldReason:    reason,
		}
	}
	return packets
}

// lifecycleFor encodes the BINDING status→lifecycle mapping documented on
// Fold. It is the single place that decides State/Hold/HoldReason, so the
// table lives in exactly one spot.
func lifecycleFor(status string, caught bool, openQuestions int) (Lifecycle, HoldKind, string) {
	switch status {
	case "queued":
		return Composing, HoldNone, ""
	case "running":
		return InFlight, HoldNone, ""
	case "done":
		if caught && openQuestions == 0 {
			return Verified, HoldNone, ""
		}
		if openQuestions > 0 {
			return Held, HoldAdvisory, fmt.Sprintf("open questions · %d", openQuestions)
		}
		return Held, HoldAdvisory, "gap found · handshake not tightened"
	case "failed":
		return Held, HoldBlocking, "run failed"
	default:
		return Held, HoldBlocking, fmt.Sprintf("unknown state · %s", status)
	}
}
