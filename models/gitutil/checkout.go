package gitutil

import (
	"io"
	"os/exec"
	"strings"
)

// Checkout runs a better version of `git checkout` or `git pull`
func Checkout(dir string, commit string) io.Reader {
	script := strings.Join([]string{
		"git fetch --prune",
		"git checkout -f $DEPLOY_COMMIT",
		"git reset --hard $DEPLOY_COMMIT",
		"git clean -fdx",
		"git submodule sync",
		"git submodule init",
		"git submodule update --recursive",
	}, " && ")
	cmd := exec.Command("bash", "-x", "-c", "("+script+") 2>&1")

	cmd.Dir = dir
	cmd.Env = append(cmd.Env, "DEPLOY_COMMIT="+commit)
	r, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w
	go func() {
		defer w.Close()
		_ = cmd.Run()
	}()
	return r
}
