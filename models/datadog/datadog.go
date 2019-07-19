package datadog

import (
	"bytes"
	"crypto/md5"
	"html/template"
	"log"

	"github.com/zorkian/go-datadog-api"
)

// DatadogConfig is a config for Datadog
type DatadogConfig struct {
	APIKey              string
	APPKey              string
	LockGainedMessage   string
	LockReleasedMessage string
	LockExtendedMessage string
	DeployedMessage     string
}

var config DatadogConfig

// SetDatadogConfig sets global Datadog config
func SetDatadogConfig(c DatadogConfig) {
	config = c
}

// LockGained sends Datadog when lock is gained
func LockGained(project, user string) {
	process(config.LockGainedMessage, project, user, "")
}

// LockReleased sends Datadog when lock is released
func LockReleased(project, user string) {
	process(config.LockReleasedMessage, project, user, "")
}

// LockExtended sends Datadog when lock is extended
func LockExtended(project, user string) {
	process(config.LockExtendedMessage, project, user, "")
}

// Deployed sends Datadog when a user deployed
func Deployed(project, user, env string) {
	process(config.DeployedMessage, project, user, env)
}

func process(message, project, user, env string) {
	if config.APIKey == "" || config.APPKey == "" || message == "" {
		return
	}

	eventTag := []string{}
	eventTag = append(eventTag, "project:"+project)
	if env != "" {
		eventTag = append(eventTag, "env:"+env)
	}

	aggregationKey := md5.Sum([]byte(project + user))

	e := datadog.Event{
		Title:       datadog.String(makeText(message, params{project, user, env})),
		Aggregation: datadog.String(string(aggregationKey[:])),
		SourceType:  datadog.String("pploy"),
		Tags:        eventTag,
		Url:         datadog.String("www.pixiv.net"),
		Resource:    datadog.String("pploy"),
	}

	client := datadog.NewClient(config.APIKey, config.APPKey)
	go client.PostEvent(&e)
}

type params struct {
	Project string
	User    string
	Env     string
}

func makeText(tmpl string, a interface{}) string {
	var buf bytes.Buffer
	tp := template.Must(template.New("calculator").Parse(tmpl))
	err := tp.Execute(&buf, a)
	if err != nil {
		log.Println("failed to process template: " + tmpl)
		return ""
	}
	return buf.String()
}
