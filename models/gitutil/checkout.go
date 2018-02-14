package gitutil

import (
	"io"
	"strings"

	"github.com/edvakf/go-pploy/unbuffered"
	"github.com/pkg/errors"
)

// Checkout runs a better version of `git checkout` or `git pull`
func Checkout(dir string, commit string) (io.Reader, error) {
	script := strings.Join([]string{
		"git fetch --prune",
		"git checkout -f $DEPLOY_COMMIT",
		"git reset --hard $DEPLOY_COMMIT",
		"git clean -fdx",
		"git submodule sync",
		"git submodule init",
		"git submodule update --recursive",
	}, " && ")
	cmd := unbuffered.Command("bash -x -c '" + script + "'")

	cmd.Dir = dir
	cmd.Env = append(cmd.Env, "DEPLOY_COMMIT="+commit)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run command")
	}
	go cmd.Wait()

	return out, nil
}
