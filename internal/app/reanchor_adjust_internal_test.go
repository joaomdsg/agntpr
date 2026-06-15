package app

import "testing"

// reanchorAdjustment is the content-relocation core of the review-thread loop's
// "outdated / moved" signal (DESIGN §28, thin slice): after the agent settles a new
// revision, the Lead's adjustment must say whether the line they commented on is still
// there, moved, or was edited away — instead of silently pointing at a stale number.
func TestReanchorAdjustment_tellsTheLeadWhetherTheirLineSurvived(t *testing.T) {
	const file = "package main\n\nfunc add(a, b int) int {\n\treturn a + b\n}\n"
	cases := []struct {
		name      string
		origLine  string
		origNum   int
		content   string
		wantState adjAnchorState
		wantLine  int
	}{
		{"unchanged line at the same number is still anchored",
			"\treturn a + b", 4, file, adjSame, 4},
		{"the same content shifted down by inserted lines reads as moved",
			"\treturn a + b", 4, "// a new header comment\n// and another\n" + file, adjMoved, 6},
		{"the anchored line edited away is outdated — the agent addressed it",
			"\treturn a + b", 4, "package main\n\nfunc add(a, b int) int {\n\treturn a - b\n}\n", adjOutdated, 0},
		{"the anchored line deleted (file shorter) is outdated",
			"\treturn a + b", 4, "package main\n", adjOutdated, 0},
		{"content found elsewhere even when the old number is out of range reads as moved",
			"\treturn a + b", 999, file, adjMoved, 4},
		{"a \\r\\n file still matches a \\n-anchored line (trailing-CR tolerance)",
			"\treturn a + b", 4, "package main\r\n\r\nfunc add(a, b int) int {\r\n\treturn a + b\r\n}\r\n", adjSame, 4},
		{"an empty anchor has nothing to relocate and is outdated",
			"", 4, file, adjOutdated, 0},
		{"first match wins when a moved anchor's content appears more than once",
			"\tx := 1", 2, "\tx := 1\n\ty := 2\n\tx := 1\n", adjMoved, 1},
		{"an unchanged line at its own number stays Same, not reported moved to an earlier duplicate",
			"\tx := 1", 2, "\tx := 1\n\tx := 1\n", adjSame, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := reanchorAdjustment(tc.origLine, tc.origNum, tc.content)
			if got.State != tc.wantState {
				t.Errorf("reanchorAdjustment(%q,%d,...) State = %v, want %v", tc.origLine, tc.origNum, got.State, tc.wantState)
			}
			if got.Line != tc.wantLine {
				t.Errorf("reanchorAdjustment(%q,%d,...) Line = %d, want %d", tc.origLine, tc.origNum, got.Line, tc.wantLine)
			}
		})
	}
}
