package models

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/edvakf/go-pploy/models/workdir"
	"github.com/edvakf/go-pploy/unbuffered"
	"github.com/pkg/errors"
)

// Checkout runs either default checkout command or checkout_overwrite script
func (p *Project) Checkout(commit string) (io.Reader, error) {
	var cmd *exec.Cmd

	script := workdir.ProjectDir(p.Name) + "/.deploy/bin/checkout_overwrite"
	if fileExists(script) {
		cmd = unbuffered.Command("bash -x -c '" + script + "'")
	} else {
		cmd = checkoutCommand()
	}
	cmd.Dir = workdir.ProjectDir(p.Name)
	cmd.Env = append(cmd.Env, "DEPLOY_COMMIT="+commit)

	return streamStdout(cmd)
}

// Deploy runs project's deploy script
func (p *Project) Deploy(env string, user string) (io.Reader, error) {
	script := workdir.ProjectDir(p.Name) + "/.deploy/bin/deploy"
	cmd := unbuffered.Command(script)
	cmd.Dir = workdir.ProjectDir(p.Name)
	cmd.Env = append(cmd.Env, "DEPLOY_ENV="+env)
	cmd.Env = append(cmd.Env, "DEPLOY_USER="+user)

	return streamStdout(cmd)
}

// CheckoutCommand is a better version of `git checkout` or `git pull`
func checkoutCommand() *exec.Cmd {
	script := strings.Join([]string{
		"git fetch --prune",
		"git checkout -f $DEPLOY_COMMIT",
		"git reset --hard $DEPLOY_COMMIT",
		"git clean -fdx",
		"git submodule sync",
		"git submodule init",
		"git submodule update --recursive",
	}, " && ")

	return unbuffered.Command("bash -x -c '" + script + "'")
}

func streamStdout(cmd *exec.Cmd) (io.Reader, error) {
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run command")
	}
	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run command")
	}
	go cmd.Wait()

	return out, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
