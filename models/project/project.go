package project

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/edvakf/go-pploy/models/cache"
	"github.com/edvakf/go-pploy/models/datadog"
	"github.com/edvakf/go-pploy/models/headreader"
	"github.com/edvakf/go-pploy/models/hook"
	"github.com/edvakf/go-pploy/models/locks"
	"github.com/edvakf/go-pploy/models/workdir"
	"github.com/edvakf/go-pploy/unbuffered"
	"github.com/pkg/errors"
)

// Project is a git-controlled deployable project directory
type Project struct {
	Lock          *locks.Lock `json:"lock"`
	Name          string      `json:"name"`
	DeployEnvs    []string    `json:"deployEnvs"`
	Readme        string      `json:"readme"`
	DefaultBranch string      `json:"defaultBranch"`
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

// Full creates a Project from its name and populates properties
func Full(name string) (*Project, error) {
	p, err := FromName(name)
	if err != nil {
		return nil, err
	}
	err = p.readReadme()
	if err != nil {
		return nil, err
	}
	err = p.readDeployEnvs()
	if err != nil {
		return nil, err
	}
	p.Lock = locks.Check(p.Name, time.Now())

	defaultBranch, err := p.GetCachedDefaultBranch()
	if err != nil {
		return nil, err
	}
	p.DefaultBranch = defaultBranch

	return p, nil
}

// Clone runs `git clone` for project repo
func Clone(url string) (*Project, error) {
	cmd := exec.Command("git", "clone", url, "--depth", "20", "--no-single-branch") // TODO: make it configurable
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
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DEPLOY_COMMIT="+commit)

	return stdoutStderrReader(cmd, nil)
}

// Deploy runs project's deploy script
func (p *Project) Deploy(env string, user string) (io.Reader, error) {
	script := workdir.ProjectDir(p.Name) + "/.deploy/bin/deploy"
	cmd := unbuffered.Command(script)
	cmd.Dir = workdir.ProjectDir(p.Name)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DEPLOY_ENV="+env)
	cmd.Env = append(cmd.Env, "DEPLOY_USER="+user)

	err := workdir.RotateLogs(p.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to rotate log files")
	}

	// write to log file
	f, err := os.OpenFile(workdir.LogFile(p.Name, 0), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open log file")
	}
	callback := func() {
		f.Close()
		datadog.Deployed(p.Name, user, env)
		hook.Deployed(p.Name, user, env)
	}
	r, err := stdoutStderrReader(cmd, callback)
	if err != nil {
		f.Close()
		return nil, err
	}

	r2 := io.TeeReader(r, f)
	return r2, nil
}

func stdoutStderrReader(cmd *exec.Cmd, callback func()) (io.Reader, error) {
	// StdoutPipe returns a ReadCloser, but it's not meant to be Close()'ed by users
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stdout pipe")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stderr pipe")
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
	return io.MultiReader(stdout, stderr), nil
}

func (p *Project) readReadme() error {
	readmeFile := workdir.ProjectDir(p.Name) + "/.deploy/config/readme.html"
	if fileExists(readmeFile) {
		b, err := ioutil.ReadFile(readmeFile)
		if err != nil {
			return errors.Wrap(err, "failed reading file") // TODO: should panic?
		}
		readme := string(b)
		p.Readme = readme
	}

	return nil
}

func (p *Project) readDeployEnvs() error {
	envsFile := workdir.ProjectDir(p.Name) + "/.deploy/config/deploy_envs"
	envs := []string{"staging", "production"} // default
	if fileExists(envsFile) {
		b, err := ioutil.ReadFile(envsFile)
		if err != nil {
			return errors.Wrap(err, "failed reading file") // TODO: should panic?
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
func (p *Project) LogReader(full bool, generation int) (io.ReadCloser, error) {
	logFile := workdir.LogFile(p.Name, generation)
	f, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &EmptyReadCloser{}, nil
		}
		return nil, err
	}
	if full {
		return f, nil
	}
	return headreader.New(f, 10000), nil // first 10000 bytes
}

// GetDefaultBranch returns default branch of repository
func (p *Project) GetDefaultBranch() (string, error) {
	cmd := exec.Command("git", "remote", "show", "origin")
	cmd.Dir = workdir.ProjectDir(p.Name)
	reader, err := stdoutStderrReader(cmd, nil)

	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	stdout := buf.String()
	re := regexp.MustCompile("HEAD branch: (\\S+)")
	group := re.FindStringSubmatch(stdout)

	return group[1], nil
}

// GetCachedDefaultBranch returns cached default branch if exists.
// GetCachedDefaultBranch returns the default branch from memory if cached, otherwise, compute and cache it.
func (p *Project) GetCachedDefaultBranch() (string, error) {
	cachedDefaultBranch := cache.DefaultBranch.Load(p.Name)

	if cachedDefaultBranch != "" {
		// returns cached default branch
		return cachedDefaultBranch, nil
	}

	defaultBranch, err := p.GetDefaultBranch()

	if err != nil {
		return "", err
	}

	cache.DefaultBranch.Store(p.Name, defaultBranch)

	return defaultBranch, nil
}

type EmptyReadCloser struct{}

func (b *EmptyReadCloser) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (b *EmptyReadCloser) Close() error {
	return nil
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
