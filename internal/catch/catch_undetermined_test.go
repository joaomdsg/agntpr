package catch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joaomdsg/packets/internal/catch"
)

// The phantom-catch regression: if the after-fix oracle run did NOT finish (its mutants
// timed out → Undetermined), the survivor-set is empty only because the run was
// incomplete, not because the line was constrained. Minting a Catch here would credit a
// fix for merely making the suite hang. Detect must fail closed to NoOracleSignal.
func TestDetect_refusesPhantomCatchWhenAfterRunIsIncomplete(t *testing.T) {
	t.Parallel()
	before := catch.LineState{Inventory: []string{"=="}, Survivors: []string{"=="}}
	after := catch.LineState{Inventory: []string{"=="}, Survivors: nil, Undetermined: []string{"=="}}
	assert.Equal(t, catch.NoOracleSignal, catch.Detect(before, after),
		"an incomplete after-run must never mint a Catch from an empty survivor-set")
}

// A genuinely complete after-run that killed every survivor still mints a Catch — the
// fail-closed guard must not over-suppress real catches.
func TestDetect_stillMintsCatchWhenAfterRunCompletedAndKilledAll(t *testing.T) {
	t.Parallel()
	before := catch.LineState{Inventory: []string{"=="}, Survivors: []string{"=="}}
	after := catch.LineState{Inventory: []string{"=="}, Survivors: nil, Undetermined: nil}
	assert.Equal(t, catch.Catch, catch.Detect(before, after),
		"a complete run that cleared the survivors is a real catch")
}

// Even when the after survivor-set merely shrank (a partial-looking transition), an
// incomplete run taints the comparison — we cannot trust a shrink we didn't fully
// observe. Fail closed rather than mint a PartialCatch.
func TestDetect_refusesPartialCatchWhenAfterRunIsIncomplete(t *testing.T) {
	t.Parallel()
	before := catch.LineState{Inventory: []string{"==", "&&"}, Survivors: []string{"==", "&&"}}
	after := catch.LineState{Inventory: []string{"==", "&&"}, Survivors: []string{"=="}, Undetermined: []string{"&&"}}
	assert.Equal(t, catch.NoOracleSignal, catch.Detect(before, after),
		"a shrink observed through an incomplete run is not a trustworthy partial catch")
}

// Only the AFTER run's completeness gates the verdict: an incomplete BEFORE run does not
// suppress a catch when the after run completed and cleared the survivors. (The before
// survivor-set is what it is; the guard is about not trusting an incomplete AFTER.)
func TestDetect_incompleteBeforeRunDoesNotSuppressAGenuineAfterCatch(t *testing.T) {
	t.Parallel()
	before := catch.LineState{Inventory: []string{"=="}, Survivors: []string{"=="}, Undetermined: []string{"=="}}
	after := catch.LineState{Inventory: []string{"=="}, Survivors: nil, Undetermined: nil}
	assert.Equal(t, catch.Catch, catch.Detect(before, after),
		"a complete after-run that cleared survivors is a catch regardless of the before-run's completeness")
}

// A complete partial transition (survivors shrank, nothing undetermined) still reports
// PartialCatch — the guard only fires on incompleteness.
func TestDetect_stillReportsPartialCatchWhenAfterRunIsComplete(t *testing.T) {
	t.Parallel()
	before := catch.LineState{Inventory: []string{"==", "&&"}, Survivors: []string{"==", "&&"}}
	after := catch.LineState{Inventory: []string{"==", "&&"}, Survivors: []string{"=="}, Undetermined: nil}
	assert.Equal(t, catch.PartialCatch, catch.Detect(before, after),
		"a complete shrink is still a partial catch")
}
