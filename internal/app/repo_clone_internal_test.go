package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Session create must tell a clonable URL from a local path, so a URL triggers a clone
// while a path is used in place (DESIGN §15.2). The common remote forms — https, scp-
// style git@, ssh://, and a bare .git — are URLs; an absolute or relative filesystem
// path is not.
func TestIsRepoURL_distinguishesRemotesFromLocalPaths(t *testing.T) {
	urls := []string{
		"https://github.com/owner/repo.git",
		"https://github.com/owner/repo",
		"git@github.com:owner/repo.git",
		"ssh://git@host/owner/repo.git",
		"http://example.com/r.git",
	}
	for _, u := range urls {
		assert.True(t, isRepoURL(u), "%q is a clonable URL", u)
	}
	paths := []string{"/home/jgonc/repos/packets", "./relative", "myrepo", "", "../up"}
	for _, p := range paths {
		assert.False(t, isRepoURL(p), "%q is a local path, not a URL", p)
	}
}

// The clone lands under a directory named for the repo, so two clones don't collide and
// the local dir is recognizable — derived from the URL's last segment with any .git
// suffix stripped, regardless of remote form.
func TestCloneDirName_derivesTheRepoNameFromAnyRemoteForm(t *testing.T) {
	assert.Equal(t, "repo", cloneDirName("https://github.com/owner/repo.git"))
	assert.Equal(t, "repo", cloneDirName("https://github.com/owner/repo"))
	assert.Equal(t, "repo", cloneDirName("git@github.com:owner/repo.git"))
	assert.Equal(t, "repo", cloneDirName("ssh://git@host/owner/repo.git"))
	assert.Equal(t, "repo", cloneDirName("https://github.com/owner/repo/")) // trailing slash tolerated
}

// A URL with no recognizable repo segment must not yield an empty (or traversal) dir
// name — it falls back to a safe default so the clone target is always a real folder.
func TestCloneDirName_fallsBackOnAnUnparseableURL(t *testing.T) {
	assert.NotEqual(t, "", cloneDirName("https://"))
	assert.NotContains(t, cloneDirName("https://host/.git"), "..")
}
