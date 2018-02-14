package gitutil

import (
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Commit is the structured git commit object
type Commit struct {
	Hash       string    `json:"hash"`
	Time       time.Time `json:"time"`
	Author     string    `json:"author"`
	OtherRefs  []string  `json:"otherRefs"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	NameStatus string    `json:"nameStatus"`
}

// RecentCommits runs `git log` and parse the result
func RecentCommits(dir string) ([]Commit, error) {
	delim1 := "1PPLOY1YOLPP1"
	delim2 := "2PPLOY2YOLPP2"
	format := delim1 + strings.Join([]string{"%H", "%ai", "%an", "%d", "%s", "%b", ""}, delim2) // hash, isoLikeDate, author, refs, subject, body, nameStatus
	cmd := exec.Command(
		"git",
		"log",
		"-n",
		"20",              //TODO: make it configurable
		"--decorate=full", // prefix refs with refs/heads/, refs/remotes/origin/ and so on
		"--name-status",   // show list of file diffs
		"-m",              // show file diffs for merge commit
		"--first-parent",  // -m shows file diffs from each parent. --first-parent make it from the first parent
		"--pretty=format:"+format,
	)
	cmd.Dir = dir
	// err := cmd.Run()
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec git command")
	}

	chunks := strings.Split(string(out), delim1)
	commits := []Commit{}
	for _, chunk := range chunks[1:] {
		parts := strings.Split(chunk, delim2)

		t, err := time.Parse("2006-01-02 15:04:05 -0700", parts[1])
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse time")
		}

		commits = append(commits, Commit{
			Hash:       parts[0],
			Time:       t,
			Author:     parts[2],
			OtherRefs:  parseRefs(parts[3]),
			Subject:    parts[4],
			Body:       strings.TrimSpace(parts[5]),
			NameStatus: strings.TrimSpace(parts[6]),
		})
	}
	return commits, nil
}

// refString looks like
// " (HEAD -> refs/heads/master, refs/remotes/origin/master, refs/remotes/origin/HEAD)"
// and parseRefs returns []string{"HEAD","refs/heads/master","refs/remotes/origin/master","refs/remotes/origin/HEAD"}
func parseRefs(refString string) []string {
	refs := []string{}
	for _, ref := range strings.Split(strings.TrimSuffix(strings.TrimPrefix(refString, " ("), ")"), ", ") {
		if ref == "" {
			continue
		}
		for _, r := range strings.Split(ref, " -> ") {
			refs = append(refs, r)
		}
	}
	return refs
}
