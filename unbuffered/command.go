package unbuffered

import "os/exec"

// libc does not line buffer if output is not a terminal, instead use full buffering.
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
	// TODO: quote the command
	return exec.Command("bash", "-c", stdbuf+" -oL -eL "+c+" 2>&1") // Linux
}
