package main

import (
	"os/exec"
	"regexp"
	"time"

	"github.com/edvakf/go-pploy/models"
	"github.com/edvakf/go-pploy/models/locks"
	"github.com/edvakf/go-pploy/models/workdir"
	"github.com/pkg/errors"
)

// Project プロジェクト。DeployEnvsとReadmeはAllProjectではセットされずCurrentProjectではセットされる
type Project struct {
	Lock       *models.Lock `json:"lock"`
	Name       string       `json:"name"`
	DeployEnvs []string     `json:"deployEnvs"`
	Readme     *string      `json:"readme"`
}

// Status ステータスAPIのレスポンス形式
type Status struct {
	AllProjects    []Project `json:"allProjects"`
	CurrentProject *Project  `json:"currentProject"`
	AllUsers       []string  `json:"allUsers"`
	CurrentUser    *string   `json:"currentUser"`
}

func MakeCurrentProject(projects []Project, name string) *Project {
	for _, project := range projects {
		if project.Name == name {
			return &Project{
				Lock:       project.Lock,
				Name:       project.Name,
				DeployEnvs: []string{"production"}, // TODO: read from config
				Readme:     nil,                    // TODO:
			}
		}
	}
	return nil
}

func GetAllProjects() ([]Project, error) {
	names, err := workdir.ProjectNames()
	if err != nil {
		return nil, err
	}
	projects := []Project{}
	now := time.Now()
	for _, name := range names {
		projects = append(projects, Project{
			Lock: locks.Check(name, now),
			Name: name,
		})
	}
	return projects, nil
}

func CreateProject(url string) (string, error) {
	cmd := exec.Command("git", "clone", url)
	cmd.Dir = workdir.ProjectsDir()
	err := cmd.Run()
	if err != nil {
		return "", errors.Wrap(err, "failed to clone repo")
	}

	// extract repo name
	submatch := regexp.MustCompile(`([^/]+?)(?:\.git)?(?:/)?$`).FindStringSubmatch(url)
	if submatch == nil {
		return "", errors.New("failed to determine repo name")
	}
	name := submatch[1]

	return name, nil
}

func ProjectExists(name string) bool {
	names, err := workdir.ProjectNames()
	if err != nil {
		return false // TODO: ここはこれでいいんだっけ？
	}
	for _, n := range names {
		if name == n {
			return true
		}
	}
	return false
}
