package app

import (
	"context"
	"net/url"
	"strconv"

	"github.com/go-via/via/h"

	"github.com/joaomdsg/packets/internal/diff"
	"github.com/joaomdsg/packets/internal/ledger"
)

// diffCompute is the base→fix differ the tree overlay reads; a package var so
// tests inject a canned diff (the real one shells out to git).
var diffCompute = diff.Compute

// renderFileTree renders the order's full fix-tree as a nested, collapsible file
// tree with the base→fix changes highlighted in place — the review surface's
// left rail. Each leaf is a plain href click-through (/review?wo=<id>&file=<path>),
// NOT a datastar @post: selecting a file is pure navigation, and the diff island
// it drives (Slice 3) is data-ignore-morph + mount-guarded, so an SSE morph could
// never swap its content — only a fresh navigation re-mounts it (council R-converged).
func renderFileTree(cfg LiveConfig, tgt ledger.Target, woID int, selected string) h.H {
	ctx := context.Background()
	changed, _ := diffCompute(ctx, cfg.RepoDir, tgt.BaseRev, tgt.FixRev)
	allPaths, err := fileListAt(ctx, cfg.RepoDir, tgt.FixRev)
	if err != nil {
		// Without the fix tree, show the changed files AS changed — never mislabel a
		// real edit as a deletion just because the working-tree listing failed.
		allPaths = changedPaths(changed)
	}
	root := buildFileTree(allPaths, changed)
	kids := append([]h.H{h.Class("file-tree"), h.Attr("aria-label", "changed files")},
		renderTreeChildren(root.children, woID, selected)...)
	return h.Div(kids...)
}

func changedPaths(d diff.Diff) []string {
	out := make([]string, 0, len(d.Files))
	for _, f := range d.Files {
		out = append(out, f.Path)
	}
	return out
}

// renderTreeChildren renders a node's children: directories as expanded-by-default
// <details> groups, files as href leaves.
func renderTreeChildren(nodes []*treeNode, woID int, selected string) []h.H {
	out := make([]h.H, 0, len(nodes))
	for _, n := range nodes {
		if n.isDir {
			out = append(out, h.Details(
				h.Attr("open"),
				h.Summary(h.Class("file-tree__dir"), h.Text(n.name)),
				h.Div(append([]h.H{h.Class("file-tree__children")},
					renderTreeChildren(n.children, woID, selected)...)...),
			))
			continue
		}
		out = append(out, renderTreeLeaf(n, woID, selected))
	}
	return out
}

func renderTreeLeaf(n *treeNode, woID int, selected string) h.H {
	class := "file-tree__file"
	switch n.status {
	case statusChanged:
		class += " file-tree__file--changed"
	case statusDeleted:
		class += " file-tree__file--deleted"
	}
	attrs := []h.H{
		h.Class(class),
		h.Href("/review?wo=" + strconv.Itoa(woID) + "&file=" + url.QueryEscape(n.path)),
		h.Span(h.Class("file-tree__name"), h.Text(n.name)),
	}
	if n.path == selected {
		attrs[0] = h.Class(class + " file-tree__file--selected")
		attrs = append(attrs, h.Attr("aria-current", "true"))
	}
	if n.status != statusUnchanged {
		attrs = append(attrs, h.Span(h.Class("file-tree__counts"), h.Text(countLabel(n))))
	}
	return h.A(attrs...)
}

// countLabel formats a changed leaf's line delta as "+A −B" (U+2212 minus, the
// design language's glyph) — a quiet diff stat, never a gauge.
func countLabel(n *treeNode) string {
	return "+" + strconv.Itoa(n.added) + " −" + strconv.Itoa(n.deleted)
}
