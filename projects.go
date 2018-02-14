package main

import (
	"github.com/edvakf/go-pploy/models/project"
)

// Status ステータスAPIのレスポンス形式
type Status struct {
	AllProjects    []project.Project `json:"allProjects"`
	CurrentProject *project.Project  `json:"currentProject"`
	AllUsers       []string          `json:"allUsers"`
	CurrentUser    *string           `json:"currentUser"`
}
