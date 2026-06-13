package app

import "strings"

// isRepoURL reports whether a repo pick is a clonable REMOTE URL (vs a local path):
// session create clones a URL but uses a path in place (DESIGN §15.2). It recognizes
// the common remote forms — https/http, ssh, git protocol, and the scp-style git@host:
// — by prefix; anything else (an absolute or relative filesystem path) is a local path.
func isRepoURL(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "ssh://") ||
		strings.HasPrefix(s, "git://") ||
		strings.HasPrefix(s, "git@")
}

// cloneDirName derives the local directory name a remote URL clones into: the URL's
// last path segment with any ".git" suffix (and trailing slash) stripped, so the clone
// lands in a recognizable, repo-named folder regardless of remote form. An unparseable
// URL with no usable segment falls back to "repo" so the target is always a real,
// non-traversal folder name.
func cloneDirName(url string) string {
	s := strings.TrimSpace(url)
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimSuffix(s, ".git")
	if i := strings.LastIndexAny(s, "/:"); i >= 0 {
		s = s[i+1:]
	}
	if s == "" || s == "." || s == ".." {
		return "repo"
	}
	return s
}
