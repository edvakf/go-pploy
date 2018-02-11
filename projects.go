package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"time"

	"github.com/pkg/errors"
)

// lockされているかどうかだけを管理し、一覧などはworkdirから毎回作る
var locks map[string]Lock

// Lock デプロイ中状態を管理
type Lock struct {
	User    string   `json:"user"`
	EndTime JSONTime `json:"endTime"`
}

// JSONTime シリアライズ可能なTime型
type JSONTime time.Time

// MarshalJSON JSONTimeをシリアライズするためのMarshalerインターフェイスの実装
func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(t).Format(time.RFC3339)), nil
}

// Project プロジェクト。DeployEnvsとReadmeはAllProjectではセットされずCurrentProjectではセットされる
type Project struct {
	Lock       *Lock    `json:"lock"`
	Name       string   `json:"name"`
	DeployEnvs []string `json:"deployEnvs"`
	Readme     *string  `json:"readme"`
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
	names, err := getAllProjectNames()
	if err != nil {
		return nil, err
	}
	projects := []Project{}
	now := time.Now()
	// TODO: map操作をmutexで排他制御、あるいはlocksをsync.Mapにする
	for _, name := range names {
		var lock *Lock
		if l, ok := locks[name]; ok && now.Before(time.Time(l.EndTime)) {
			lock = &l
		}
		projects = append(projects, Project{
			Lock: lock,
			Name: name,
		})
	}
	return projects, nil
}

func getAllProjectNames() ([]string, error) {
	files, err := ioutil.ReadDir(WorkDir + "/projects")
	if err != nil {
		return nil, err
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

func CreateProject(url string) (string, error) {
	os.Chdir(WorkDir + "/projects")
	err := exec.Command("git", "clone", url).Run()
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
