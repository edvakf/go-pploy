package workdir

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/pkg/errors"
)

var workDir string

var LogMax int

// Init sets an internal workDir variable and prepares the working directory
func Init(dir string) {
	workDir = dir

	os.MkdirAll(WorkDir(), os.ModePerm)
	os.MkdirAll(ProjectsDir(), os.ModePerm)
	os.MkdirAll(LogsDir(), os.ModePerm)
}

// WorkDir returns the working directory
func WorkDir() string {
	assetInitialized()
	return workDir
}

// ProjectsDir returns the directory for git repos
func ProjectsDir() string {
	assetInitialized()
	return workDir + "/projects"
}

// LogsDir returns the directory for deploy logs
func LogsDir() string {
	assetInitialized()
	return workDir + "/logs"
}

// ProjectDir returns the git repo directory for of a project
func ProjectDir(name string) string {
	return ProjectsDir() + "/" + name
}

// LogFile returns the log file for of a project
func LogFile(name string, generation int) string {
	suffix := ""
	if generation != 0 {
		suffix = fmt.Sprintf(".%d", generation)
	}
	return LogsDir() + "/" + name + ".log" + suffix
}

func RotateLogs(name string) error {
	for i := LogMax; i > 0; i-- {
		err := os.Rename(LogFile(name, i-1), LogFile(name, i))
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to move log file")
		}
	}
	return nil
}

func assetInitialized() {
	if workDir == "" {
		panic("please initialize workdir")
	}
}

// ProjectNames returns directory names under the working directory
func ProjectNames() ([]string, error) {
	files, err := ioutil.ReadDir(ProjectsDir())
	if err != nil {
		return nil, errors.Wrap(err, "failed to list directory")
	}

	dirs := []string{}
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

// RemoveProjectFiles deletes project's git directory and log directory
func RemoveProjectFiles(name string) error {
	err := os.RemoveAll(ProjectDir(name))
	if err != nil {
		return errors.Wrap(err, "failed to delete project files")
	}

	for i := LogMax; i >= 0; i-- {
		err = os.Remove(LogFile(name, i))
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to delete log file")
		}
	}
	return nil
}
