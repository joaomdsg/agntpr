package app

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-via/via/vt"
)

// DRY: .review-answer__submit reuses .pk-btn (which already owns the hairline
// border). The submit rule must NOT re-declare a full `border:` shorthand — it only
// reinforces the accent via `border-color`, letting .pk-btn own the border width/style.
func TestReviewSubmit_doesNotRedeclareTheFullBorderShorthand(t *testing.T) {
	require.Contains(t, packetsStyle, "border-color: var(--signal)",
		".review-answer__submit reinforces the accent via border-color, not a full border shorthand")
	require.NotContains(t, packetsStyle, "border: 1px solid var(--signal)",
		".review-answer__submit must not re-declare the hairline border that .pk-btn already owns")
}

// Without the base stylesheet attached to the page <head>, the calm visual
// language never reaches the browser — the whole UX/UI direction is dead on
// arrival. Every rendered page must carry our stylesheet. NOT parallel (shared
// liveReg/liveFabric).
func TestNewServer_attachesTheBaseStylesheetToEveryPage(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	for _, path := range []string{"/", "/board"} {
		body := vt.NewClient(t, server, path).HTML()
		require.Containsf(t, body, "--signal:", "%s must carry OUR design tokens (the state grammar), proving it is the packets stylesheet", path)
		require.Containsf(t, body, ".board-row", "%s's stylesheet must target the real class hooks", path)
		// The <style> must live in the <head> (not stray into the body).
		headEnd := strings.Index(body, "</head>")
		stylePos := strings.Index(body, "<style")
		require.Greaterf(t, stylePos, -1, "%s must carry an inline <style>", path)
		require.Greaterf(t, headEnd, stylePos, "%s's <style> must be inside the <head>", path)
	}

	// Attaching the head must not disturb the body render — the board markup is
	// unchanged.
	board := vt.NewClient(t, server, "/board").HTML()
	require.Contains(t, board, "board-row__stock", "the board body still renders its rows after the head is attached")
	require.NotContains(t, strings.ToLower(board), "progress-bar", "no gauges/progress bars (calm guardrail)")
}

// Every honest verdict + land STATE the card can render must have a per-state
// style rule, so the Lead reads catch-vs-miss-vs-lost at a glance in the calm
// language — not as undifferentiated text. We pin the SELECTOR coverage (every
// real data-state value is targeted), never the colors (taste). If a renderer
// gains a new state, this test fails until the stylesheet styles it too.
func TestBaseStylesheet_stylesEveryVerdictAndLandState(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	// The full page (the <style> lives in the head) — we WANT to find the
	// selectors in the stylesheet here, so check the whole HTML.
	page := vt.NewClient(t, server, "/").HTML()
	// Every data-state value the surface renderers emit (verdict + land).
	for _, state := range []string{
		"catch", "no-catch", "partial-catch", "no-oracle-signal",
		"lost-via-rename", "anchor-edited", "tested", "in-flight",
		"land-clean", "land-conflict", "land-checks-red", "land-pending",
	} {
		require.Containsf(t, page, `[data-state="`+state+`"]`,
			"the stylesheet must give state %q its own calm per-state rule (legible at a glance)", state)
	}
	// Calm guardrail: per-state color must use the design tokens, never a raw
	// alarm red/green hardcode.
	require.NotContains(t, strings.ToLower(page), "#ff0000", "no alarm red")
	require.NotContains(t, strings.ToLower(page), "#00ff00", "no alarm green")
}

// The system layer promotes the brand-pack scale into named tokens (the
// MVP.md port) and one shared component layer every surface hooks. We pin the
// token names + the canonical :focus-visible rule so the contract later
// surfaces depend on cannot silently disappear. NOT parallel (shared
// liveReg/liveFabric).
func TestBaseStylesheet_definesTheSystemLayerTokens(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	page := vt.NewClient(t, server, "/").HTML()

	for _, token := range []string{
		"--r-btn:", "--r-glyph:", "--hairline:",
		"--fs-small:", "--fs-micro:",
	} {
		require.Containsf(t, page, token,
			"the system layer must define the %q scale token on :root", token)
	}
	// The WCAG 2.4.7 fix: a real focus ring on the shared components.
	require.Contains(t, page, ":focus-visible",
		"the system layer must define the shared :focus-visible focus ring")
	require.Contains(t, page, "outline: 2px solid var(--signal)",
		"the focus ring is the documented signal-cyan accent outline")
}

// Each surface keeps its semantic class and ADDS one shared component class via
// multi-class. We pin both the component selectors in the stylesheet AND the
// multi-class hooks on the rendered surfaces, so the collapse cannot regress a
// surface back to its hand-rolled box. NOT parallel (shared liveReg/liveFabric).
func TestBaseStylesheet_extractsTheSharedComponentLayer(t *testing.T) {
	defLogPath := filepath.Join(t.TempDir(), "default.jsonl")
	var server *httptest.Server
	viaApp, log, err := NewServer(LiveConfig{
		RepoDir: ".", BaseRev: "b", FixRev: "f", TipRev: "f", Anchor: anchorForCap(),
		TestCmd: []string{"true"}, LedgerPath: defLogPath,
	})
	require.NoError(t, err)
	server = httptest.NewServer(viaApp)
	t.Cleanup(func() { _ = log.Close() })

	page := vt.NewClient(t, server, "/").HTML()
	board := vt.NewClient(t, server, "/board").HTML()

	for _, sel := range []string{
		".pk-btn", ".pk-btn--quiet", ".pk-input", ".pk-chip", ".pk-section-label",
		".pk-card",
	} {
		require.Containsf(t, page, sel,
			"the system layer must define the shared %q component selector", sel)
	}

	// The padded box rows compose the shared .pk-card; the semantic class keeps
	// only hue/state/layout. We pin the multi-class hooks so the collapse cannot
	// regress a row back to its hand-rolled box.
	require.Contains(t, board, `class="pk-card board-row"`,
		"each fleet row composes .pk-card")
	for _, hook := range []string{
		"pk-card stock-row", "pk-card balance-row", "pk-card bandwidth-row",
	} {
		require.Containsf(t, page, hook,
			"the session card's %q row composes .pk-card", hook)
	}

	// The board's create input + button compose the shared classes.
	require.Contains(t, board, `class="pk-input board-create__key"`,
		"the create-key input composes .pk-input")
	require.Contains(t, board, "pk-btn",
		"the board surfaces a .pk-btn control")
	// The retire control is a quiet variant.
	require.Contains(t, board, "pk-btn--quiet",
		"the board's retire control composes .pk-btn--quiet")
	// The uppercase labels are the shared section-label component.
	require.Contains(t, board, "pk-section-label",
		"the board's uppercase labels compose .pk-section-label")
}

// ROADMAP slice 1 (token port): the state grammar IS the product (MVP.md
// "Brand pack") — a wrong hex silently breaks the honest-state legibility the
// whole design depends on, so the exact values are pinned, not just presence.
func TestBaseStylesheet_definesStateGrammarTokensWithExactHexValues(t *testing.T) {
	for _, tt := range []struct {
		name  string
		token string
		hex   string
	}{
		{"in-flight/you/accent reads signal-cyan", "--signal", "#4cc4d4"},
		{"a real catch reads verified-green", "--verified", "#46c08a"},
		{"an advisory hold reads held-amber", "--held", "#e6b23e"},
		{"a blocking hold reads risk-red", "--risk", "#f0666b"},
		{"agent authorship reads agent-purple", "--agent", "#a78bfa"},
		{"a landed/settled thing reads delivered-teal", "--delivered", "#2a7683"},
		{"the page ground is near-black", "--ground", "#0a0d13"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Containsf(t, packetsStyle, tt.token+": "+tt.hex,
				"%s must be exactly %s", tt.token, tt.hex)
		})
	}
}

// The type system is IBM Plex end to end (mono = the machine's voice, sans =
// prose) — without the import the base font stacks fall back to the system
// font and the whole voice distinction is lost.
func TestBaseStylesheet_loadsIBMPlexFonts(t *testing.T) {
	require.Contains(t, packetsStyle,
		`fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@400;500;600;700&family=IBM+Plex+Mono:wght@400;500;600;700`,
		"the stylesheet must @import IBM Plex Sans + Mono at weights 400/500/600/700")
	require.Contains(t, packetsStyle, "--font-ui:", "the sans (prose) stack must be a named token")
	require.Contains(t, packetsStyle, "--font-mono:", "the mono (machine voice) stack must be a named token")
}

// The three motion primitives (live pulse, held pulse, settle flow) are the
// ONLY animation kinds the design allows (guardrail: no bounces/slides) — a
// missing keyframe silently strands whatever rule references it.
func TestBaseStylesheet_definesTheThreeMotionKeyframes(t *testing.T) {
	for _, name := range []string{"pk-pulse", "pk-held-pulse", "pk-flow"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Containsf(t, packetsStyle, "@keyframes "+name,
				"the stylesheet must define @keyframes %s", name)
		})
	}
}

// A surviving `--pk-` token would mean some rule still points at the retired
// skin instead of the ported brand pack — the two palettes must never
// coexist, or a surface could silently re-skin back to the old system.
func TestBaseStylesheet_carriesNoLegacyPkTokens(t *testing.T) {
	require.NotContains(t, packetsStyle, "--pk-",
		"no `--pk-` custom property may remain once the token port is complete")
}

// MVP.md marks --heading-cream, --glow-cta, --fs-h2, and --fs-h1 as
// marketing-only ("never in-app"). A verbatim copy-paste of the design/tokens
// source files would wrongly drag these into the product surface, so their
// absence is pinned as its own contract, not just a review note.
func TestBaseStylesheet_excludesMarketingOnlyTokens(t *testing.T) {
	for _, token := range []string{"--heading-cream", "--glow-cta", "--fs-h2", "--fs-h1"} {
		t.Run(token, func(t *testing.T) {
			t.Parallel()
			require.NotContainsf(t, packetsStyle, token,
				"%s is marketing-only (MVP.md) and must never reach the in-app stylesheet", token)
		})
	}
}
