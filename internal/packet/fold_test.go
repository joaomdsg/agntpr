package packet_test

import (
	"testing"

	"github.com/joaomdsg/packets/internal/ledger"
	"github.com/joaomdsg/packets/internal/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFold_mapsEachDispatchStatusToItsLifecycleState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		status         string
		caught         bool
		openQuestions  int
		wantState      packet.Lifecycle
		wantHold       packet.HoldKind
		wantHoldReason string
	}{
		{
			name:           "queued is the ghost-outline promised state",
			status:         "queued",
			wantState:      packet.Composing,
			wantHold:       packet.HoldNone,
			wantHoldReason: "",
		},
		{
			name:           "running is in-flight",
			status:         "running",
			wantState:      packet.InFlight,
			wantHold:       packet.HoldNone,
			wantHoldReason: "",
		},
		{
			name:           "done, caught, no open questions is verified",
			status:         "done",
			caught:         true,
			openQuestions:  0,
			wantState:      packet.Verified,
			wantHold:       packet.HoldNone,
			wantHoldReason: "",
		},
		{
			name:           "done but not caught is an advisory hold for a gap",
			status:         "done",
			caught:         false,
			openQuestions:  0,
			wantState:      packet.Held,
			wantHold:       packet.HoldAdvisory,
			wantHoldReason: "gap found · handshake not tightened",
		},
		{
			name:           "done, caught, but a single open question is still an advisory hold",
			status:         "done",
			caught:         true,
			openQuestions:  1,
			wantState:      packet.Held,
			wantHold:       packet.HoldAdvisory,
			wantHoldReason: "open questions · 1",
		},
		{
			name:           "done, caught, but open questions is an advisory hold",
			status:         "done",
			caught:         true,
			openQuestions:  2,
			wantState:      packet.Held,
			wantHold:       packet.HoldAdvisory,
			wantHoldReason: "open questions · 2",
		},
		{
			name:           "done, not caught, AND open questions: the questions reason wins",
			status:         "done",
			caught:         false,
			openQuestions:  4,
			wantState:      packet.Held,
			wantHold:       packet.HoldAdvisory,
			wantHoldReason: "open questions · 4",
		},
		{
			name:           "failed is a blocking hold",
			status:         "failed",
			wantState:      packet.Held,
			wantHold:       packet.HoldBlocking,
			wantHoldReason: "run failed",
		},
		{
			name:           "an unknown status fails toward attention, never silently verified",
			status:         "cancelled",
			caught:         true,
			openQuestions:  0,
			wantState:      packet.Held,
			wantHold:       packet.HoldBlocking,
			wantHoldReason: "unknown state · cancelled",
		},
		{
			name:           "an empty status is also unknown — it never defaults to a benign state",
			status:         "",
			caught:         true,
			openQuestions:  0,
			wantState:      packet.Held,
			wantHold:       packet.HoldBlocking,
			wantHoldReason: "unknown state · ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			views := []ledger.DispatchView{{
				ID:     1,
				Target: ledger.Target{Prompt: "some prompt"},
				Status: tt.status,
				Caught: tt.caught,
			}}

			got := packet.Fold(views, packet.Addr{}, constQuestions(tt.openQuestions))

			require.Len(t, got, 1)
			assert.Equal(t, tt.wantState, got[0].State, "state")
			assert.Equal(t, tt.wantHold, got[0].Hold, "hold kind")
			assert.Equal(t, tt.wantHoldReason, got[0].HoldReason, "hold reason")
		})
	}
}

func TestFold_preservesOrderIdentityAndOrderingAcrossMultipleViews(t *testing.T) {
	t.Parallel()

	views := []ledger.DispatchView{
		{ID: 3, Target: ledger.Target{Prompt: "third order"}, Status: "queued"},
		{ID: 1, Target: ledger.Target{Prompt: "first order"}, Status: "running"},
		{ID: 2, Target: ledger.Target{Prompt: "second order"}, Status: "failed"},
	}

	got := packet.Fold(views, packet.Addr{}, zeroQuestions)

	require.Len(t, got, len(views))
	for i, v := range views {
		assert.Equal(t, v.ID, got[i].ID, "packet at position %d must match the view fed at that position", i)
	}
}

func TestFold_looksUpOpenQuestionsByEachViewsOwnID(t *testing.T) {
	t.Parallel()

	views := []ledger.DispatchView{
		{ID: 10, Target: ledger.Target{Prompt: "no questions here"}, Status: "done", Caught: true},
		{ID: 20, Target: ledger.Target{Prompt: "has questions here"}, Status: "done", Caught: true},
	}
	byID := map[int]int{10: 0, 20: 3}

	got := packet.Fold(views, packet.Addr{}, func(id int) int { return byID[id] })

	require.Len(t, got, 2)
	assert.Equal(t, packet.Verified, got[0].State, "order 10 has no open questions of its own — verified")
	assert.Equal(t, packet.Held, got[1].State, "order 20 has its OWN open questions — held, not order 10's count")
	assert.Equal(t, "open questions · 3", got[1].HoldReason)
}

func TestFold_returnsNoPacketsForNoViews(t *testing.T) {
	t.Parallel()

	got := packet.Fold(nil, packet.Addr{}, zeroQuestions)

	assert.Empty(t, got)
}
