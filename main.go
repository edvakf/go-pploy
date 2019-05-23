package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/edvakf/go-pploy/models/hook"
	"github.com/edvakf/go-pploy/models/ldapusers"
	"github.com/edvakf/go-pploy/models/locks"
	"github.com/edvakf/go-pploy/models/workdir"
	"github.com/edvakf/go-pploy/web"
	"github.com/facebookarchive/pidfile"
)

var GitCommit string

func main() {
	web.Server()
}

func init() {
	// commit hash is passed at build time with -ldflags
	fmt.Printf("commit: %s\n", GitCommit)

	var lockDuration time.Duration
	var workDir string
	var sc hook.SlackConfig
	var lc ldapusers.Config

	flag.DurationVar(&lockDuration, "lock", 10*time.Minute, "Duration (ex. 10m) for lock gain")
	flag.StringVar(&workDir, "workdir", "", "Working directory")
	flag.IntVar(&workdir.LogMax, "logmax", 20, "Max number of log files to keep")

	flag.StringVar(&web.PathPrefix, "prefix", "/", "Path prefix of the app (eg. /pploy/), useful for proxied apps")
	flag.IntVar(&web.Port, "port", 9000, "HTTP port")

	flag.StringVar(&sc.WebHookURL, "webhook", "", "Incoming web hook URL for slack notification")
	flag.StringVar(&sc.LockGainedMessage, "lockgained", "", "Message template for when lock is gained")
	flag.StringVar(&sc.LockReleasedMessage, "lockreleased", "", "Message template for when lock is released")
	flag.StringVar(&sc.LockExtendedMessage, "lockextended", "", "Message template for when lock is extended")
	flag.StringVar(&sc.DeployedMessage, "deployed", "", "Message template for when deploy is ended")

	flag.StringVar(&lc.Host, "ldaphost", "", "LDAP host (leave empty if ldap is not needed)")
	flag.IntVar(&lc.Port, "ldapport", 389, "LDAP port")
	flag.StringVar(&lc.BaseDN, "ldapdn", "", "LDAP base DN of user list")
	flag.DurationVar(&lc.CacheTTL, "ldapttl", 10*time.Minute, "LDAP cache TTL")

	flag.Parse()

	if workDir == "" {
		log.Fatalf("Please set workdir flag")
	}

	if pidfile.GetPidfilePath() != "" {
		err := pidfile.Write()
		if err != nil {
			log.Fatalf("failed to create pid file:%s", err.Error())
		}

		pidValue, err := pidfile.Read()
		fmt.Printf("pid:%d\n", pidValue)

		if err != nil {
			log.Fatalf("failed to read pid file:%s", err.Error())
		}
	}

	locks.SetDuration(lockDuration)
	workdir.Init(workDir)
	hook.SetSlackConfig(sc)
	ldapusers.SetConfig(lc)
}
