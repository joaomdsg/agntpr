package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/joaomdsg/packets/internal/ledger"
)

// swapFileReader points the source reader at canned per-(rev,path) content so a
// test can prove WHICH file+revision the island embedded. Not parallel (pkg var).
func swapFileReader(t *testing.T) {
	t.Helper()
	orig := reviewFileReader
	reviewFileReader = func(_ context.Context, _, rev, path string) (string, error) {
		return "SRC[" + path + "@" + rev + "]", nil
	}
	t.Cleanup(func() { reviewFileReader = orig })
}

// Clicking a tree leaf must re-point the diff editor at THAT file, not stay on
// the order's anchored path — otherwise the tree selector is inert.
func TestOrderDiffIsland_embedsTheSelectedFileNotTheAnchor(t *testing.T) {
	swapFileReader(t)
	tgt := ledger.Target{BaseRev: "base", FixRev: "fix", Path: "anchor.go"}

	body := renderHTML(t, orderDiffIsland(LiveConfig{RepoDir: "."}, tgt, "pkg/picked.go"))

	assert.Contains(t, body, `"path":"pkg/picked.go"`, "the payload carries the selected path")
	assert.Contains(t, body, "SRC[pkg/picked.go@base]", "base source is read for the selected file")
	assert.Contains(t, body, "SRC[pkg/picked.go@fix]", "fix source is read for the selected file")
	assert.NotContains(t, body, "anchor.go@", "the anchor's source is NOT read when another file is selected")
}

// Selecting a different file changes which base/fix the island embeds — the
// per-file re-point the tree relies on.
func TestOrderDiffIsland_reEmbedsWhenSelectionChanges(t *testing.T) {
	swapFileReader(t)
	tgt := ledger.Target{BaseRev: "base", FixRev: "fix", Path: "anchor.go"}

	a := renderHTML(t, orderDiffIsland(LiveConfig{RepoDir: "."}, tgt, "a.go"))
	b := renderHTML(t, orderDiffIsland(LiveConfig{RepoDir: "."}, tgt, "b.go"))

	assert.Contains(t, a, "SRC[a.go@fix]")
	assert.NotContains(t, a, "SRC[b.go@fix]")
	assert.Contains(t, b, "SRC[b.go@fix]")
	assert.NotContains(t, b, "SRC[a.go@fix]")
}

// An empty selection falls back to the order's anchored path, so the surface is
// never blank before the Lead has clicked a leaf.
func TestOrderDiffIsland_emptySelectionFallsBackToTheAnchor(t *testing.T) {
	swapFileReader(t)
	tgt := ledger.Target{BaseRev: "base", FixRev: "fix", Path: "anchor.go"}

	body := renderHTML(t, orderDiffIsland(LiveConfig{RepoDir: "."}, tgt, ""))

	assert.Contains(t, body, `"path":"anchor.go"`)
	assert.Contains(t, body, "SRC[anchor.go@base]")
}

// The island stays a calm control-room surface — no alarm hues, no gauges.
func TestOrderDiffIsland_carriesNoAlarmColorsOrGauges(t *testing.T) {
	swapFileReader(t)
	tgt := ledger.Target{BaseRev: "base", FixRev: "fix", Path: "anchor.go"}
	body := renderHTML(t, orderDiffIsland(LiveConfig{RepoDir: "."}, tgt, "anchor.go"))
	for _, banned := range []string{"#ff0000", "#00ff00", "progress-bar", "<progress", "<meter"} {
		assert.NotContains(t, body, banned)
	}
}
