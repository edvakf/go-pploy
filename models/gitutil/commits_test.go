package gitutil

import (
	"flag"
	"testing"
)

var gitDir string

// go test . -dir=xxx
func init() {
	flag.StringVar(&gitDir, "dir", ".", "a git directory")
	flag.Parse()
}

func TestRecentCommits(t *testing.T) {
	commits, err := RecentCommits(gitDir)
	if err != nil {
		t.Error(err) // ok as long as no error are retuned
	}
	// fmt.Println(commits[0])
	t.Logf("%v", commits[0])
}
