package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/joaomdsg/packets/internal/ledger"
)

// validClaimTarget is the canonical in-flight claim the lifecycle tests submit
// through the in-process ingress (the host-side equivalent of an authenticated
// producer publishing the same encoded ClaimRecord over the NATS socket). Used by
// claim_consumer_internal_test.go, claims_internal_test.go, gc_hook_internal_
// test.go, producer_e2e_internal_test.go, and lazy_consumer_internal_test.go — all
// of which stay internal (they swap/read unexported seams), so this helper stays
// here rather than moving with the one test that no longer needs it
// (TestPostClaim_isRetiredFromTheUnauthenticatedHTTPSurface, now in
// claim_ingress_test.go).
var validClaimTarget = ledger.Target{BaseRev: "basesha", FixRev: "fixsha", TipRev: "fixsha", Path: "adult.go", Line: 4}

// publishClaim submits a claim to key's grant-confined claim subtree via the
// authenticated NATS ingress's in-process equivalent (ledger.PublishClaim on the
// shared fabric). Claim submission is NATS-only since R82 — there is no HTTP
// claim edge to post to — so tests publish here exactly as the host does.
func publishClaim(t *testing.T, key string, tgt ledger.Target) {
	t.Helper()
	_, err := ledger.PublishClaim(context.Background(), liveFabric, key, LedgerInstance, ledger.ClaimRecord{Target: tgt})
	require.NoError(t, err)
}
