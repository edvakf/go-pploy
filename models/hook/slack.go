package hook

import (
	"bytes"
	"html/template"
	"log"

	slack "github.com/hnakamur/slack-incoming-webhook"
)

// SlackConfig is a config for slack
type SlackConfig struct {
	WebHookURL          string
	LockGainedMessage   string
	LockReleasedMessage string
	LockExtendedMessage string
	DeployedMessage     string
}

var config SlackConfig

// SetSlackConfig sets global slack config
func SetSlackConfig(c SlackConfig) {
	config = c
}

// LockGained sends hook when lock is gained
func LockGained(project, user string) {
	process(config.LockGainedMessage, project, user, "")
}

// LockReleased sends hook when lock is released
func LockReleased(project, user string) {
	process(config.LockReleasedMessage, project, user, "")
}

// LockExtended sends hook when lock is extended
func LockExtended(project, user string) {
	process(config.LockExtendedMessage, project, user, "")
}

// Deployed sends hook when a user deployed
func Deployed(project, user, env string) {
	process(config.DeployedMessage, project, user, env)
}

func process(message, project, user, env string) {
	if config.WebHookURL == "" || message == "" {
		return
	}
	go slack.Send(
		config.WebHookURL,
		slack.Payload{
			Text: makeText(message, params{project, user, env}),
		},
	)
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
