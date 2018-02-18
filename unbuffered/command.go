package unbuffered

import "os/exec"

// libc does not line buffer if output is not a terminal, instead use full buffering.
// `script` runs a given command in a pseudo terminal.
// it also redirects stderr to stdout which fits our usage

// see: http://unix.stackexchange.com/questions/25372/turn-off-buffering-in-pipe

var stdbuf string

func init() {
	err := exec.Command("which", "stdbuf").Run()
	if err == nil {
		stdbuf = "stdbuf"
		return
	}
	err = exec.Command("which", "gstdbuf").Run()
	if err == nil {
		stdbuf = "gstdbuf"
		return
	}
	panic("stdbuf not installed (for macOS, run `brew install coreutils`)")
}

// Command takes a shell command and wraps it with either
func Command(c string) *exec.Cmd {
	return exec.Command(stdbuf, "-oL", "-eL", c)
}
