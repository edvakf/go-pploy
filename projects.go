package main

import "time"

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

// Project プロジェクト。ReadmeはAllProjectではセットされずCurrentProjectではセットされる
type Project struct {
	Lock       []Lock   `json:"lock"`
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
