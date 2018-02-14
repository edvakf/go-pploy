package unbuffered

import (
	"testing"
)

var gitDir string

func TestCommand(t *testing.T) {
	cmd := Command("ls")
	out, err := cmd.Output()
	if err != nil {
		t.Error(err)
	}
	t.Log(string(out))
	t.Logf("%x\n", out)
}
