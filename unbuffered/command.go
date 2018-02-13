package unbuffered

import "os/exec"

// libc does not line buffer if output is not a terminal, instead use full buffering.
// `script` runs a given command in a pseudo terminal.
// it also redirects stderr to stdout which fits our usage

// see: http://unix.stackexchange.com/questions/25372/turn-off-buffering-in-pipe

var hasStdbuf bool

func init() {
	err := exec.Command("which", "stdbuf").Run()
	if err != nil {
		hasStdbuf = false
	}
	hasStdbuf = true
}
