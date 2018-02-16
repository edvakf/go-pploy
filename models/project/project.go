package project

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/edvakf/go-pploy/models"
	"github.com/edvakf/go-pploy/models/headreader"
	"github.com/edvakf/go-pploy/models/hook"
	"github.com/edvakf/go-pploy/models/locks"
	"github.com/edvakf/go-pploy/models/workdir"
	"github.com/edvakf/go-pploy/unbuffered"
	"github.com/pkg/errors"
)

// Project is a git-controlled deployable project directory
type Project struct {
	Lock       *models.Lock `json:"lock"`
	Name       string       `json:"name"`
	DeployEnvs []string     `json:"deployEnvs"`
	Readme     *string      `json:"readme"`
}

// All returns all projects
func All() ([]Project, error) {
	names, err := workdir.ProjectNames()
	if err != nil {
		return nil, err
	}
	projects := []Project{}
	now := time.Now()
	for _, name := range names {
		p, err := FromName(name)
		if err != nil {
			continue // should not happen
		}
		p.Lock = locks.Check(name, now)
		projects = append(projects, *p)
	}
	return projects, nil
}

// FromName creates a Project from its name
func FromName(name string) (*Project, error) {
	if name == "" {
		return nil, errors.New("name is empty")
	}
	dir := workdir.ProjectDir(name)
	if !fileExists(dir) {
		return nil, errors.New("project directory does not exist")
	}
	return &Project{Name: name}, nil
}

func Full(name string) (*Project, error) {
	p, err := FromName(name)
	if err != nil {
		return nil, err
	}
	err = p.ReadReadme()
	if err != nil {
		return nil, err
	}
	err = p.ReadDeployEnvs()
	if err != nil {
		return nil, err
	}
	p.Lock = locks.Check(p.Name, time.Now())

	return p, nil
}

// Clone runs `git clone` for project repo
func Clone(url string) (*Project, error) {
	cmd := exec.Command("git", "clone", url)
	cmd.Dir = workdir.ProjectsDir()
	err := cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "failed to clone repo")
	}

	// extract repo name
	submatch := regexp.MustCompile(`([^/]+?)(?:\.git)?(?:/)?$`).FindStringSubmatch(url)
	if submatch == nil {
		return nil, errors.New("failed to determine repo name")
	}
	name := submatch[1]

	return FromName(name)
}

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

	return stdoutReader(cmd, nil)
}

// Deploy runs project's deploy script
func (p *Project) Deploy(env string, user string) (io.Reader, error) {
	script := workdir.ProjectDir(p.Name) + "/.deploy/bin/deploy"
	cmd := unbuffered.Command(script)
	cmd.Dir = workdir.ProjectDir(p.Name)
	cmd.Env = append(cmd.Env, "DEPLOY_ENV="+env)
	cmd.Env = append(cmd.Env, "DEPLOY_USER="+user)

	// write to log file
	f, err := os.OpenFile(workdir.LogFile(p.Name), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open log file")
	}
	callback := func() {
		f.Close()
		hook.Deployed(p.Name, user, env)
	}
	r, err := stdoutReader(cmd, callback)
	if err != nil {
		f.Close()
		return nil, err
	}

	r2 := io.TeeReader(r, f)
	return r2, nil
	// return stdoutReader(cmd, callback)
}

func stdoutReader(cmd *exec.Cmd, callback func()) (io.Reader, error) {
	// StdoutPipe returns a ReadCloser, but it's not meant to be Close()'ed by users
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run command")
	}
	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run command")
	}
	go func() {
		cmd.Wait()
		if callback != nil {
			callback()
		}
	}()
	return out, nil
}

// ReadReadme reads readme.html file from the project directory
func (p *Project) ReadReadme() error {
	readmeFile := workdir.ProjectDir(p.Name) + "/.deploy/config/readme.html"
	if fileExists(readmeFile) {
		b, err := ioutil.ReadFile(readmeFile)
		if err != nil {
			return errors.Wrap(err, "wailed reading file") // TODO: should panic?
		}
		readme := string(b)
		p.Readme = &readme
	}

	return nil
}

// ReadDeployEnvs reads deploy_envs file from the project directory
func (p *Project) ReadDeployEnvs() error {
	envsFile := workdir.ProjectDir(p.Name) + "/.deploy/config/deploy_envs"
	envs := []string{"production"} // default
	if fileExists(envsFile) {
		b, err := ioutil.ReadFile(envsFile)
		if err != nil {
			return errors.Wrap(err, "wailed reading file") // TODO: should panic?
		}
		envs2 := removeEmpty(strings.Split(string(b), "\n"))
		if len(envs2) != 0 {
			envs = envs2
		}
	}
	p.DeployEnvs = envs

	return nil
}

// LogReader returns a ReadCloser which reads either an entire file
// or first 10000 bytes of it depending on the `full` parameter
func (p *Project) LogReader(full bool) (io.ReadCloser, error) {
	logFile := workdir.LogFile(p.Name)
	f, err := os.Open(logFile)
	if err != nil {
		return nil, err
	}
	if full {
		return f, nil
	}
	return headreader.New(f, 10000), nil // first 10000 bytes
}

func removeEmpty(a []string) (r []string) {
	for _, s := range a {
		if s != "" {
			r = append(r, s)
		}
	}
	return
}

// checkoutCommand is a better version of `git checkout` or `git pull`
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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
