# go-pploy

[pploy](https://github.com/edvakf/pploy) is a deploy manager written in Scala.
[go-pploy](https://github.com/edvakf/go-pploy) is it's port in golang.

# Build

```
make prepare
```

installs dependencies. Some commands are installed globally.

```
make
```

builds into a single binary `go-pploy`. Frontend files are compiled with `go-assets-builder`.

# CI

When tagged, travis uploads binaries to GitHub Release.

# Usage

```
Usage of ./go-pploy:
  -ddapikey string
    	Datadog API key
  -ddappkey string
    	Datadog APP key
  -ddlockgained string
    	Message template for Datadog when lock is gained
  -ddlockreleased string
    	Message template for Datadog when lock is released
  -ddlockextended string
    	Message template for Datadog when lock is extended
  -ddlockextended string
    	Message template for Datadog when deploy is ended
  -deployed string
    	Message template for when deploy is ended
  -ldapdn string
    	LDAP base DN of user list
  -ldaphost string
    	LDAP host (leave empty if ldap is not needed)
  -ldapport int
    	LDAP port (default 389)
  -ldapttl duration
    	LDAP cache TTL (default 10m0s)
  -lock duration
    	Duration (ex. 10m) for lock gain (default 10m0s)
  -lockextended string
    	Message template for when lock is extended
  -lockgained string
    	Message template for when lock is gained
  -lockreleased string
    	Message template for when lock is released
  -pidfile string
    	pid file path
  -port int
    	HTTP port (default 9000)
  -prefix string
    	Path prefix of the app (eg. /pploy/), useful for proxied apps (default "/")
  -webhook string
    	Incoming web hook URL for slack notification
  -workdir string
    	Working directory
```

# Example

```
/home/deploy/go-pploy \
  -pidfile=/home/deploy/pploy.pid \
  -prefix=/deploy/ \
  -port=9000 \
  -lock=10m \
  -workdir=/home/deploy/pploy-working-dir \
  -ldaphost="ldap.example.com" \
  -ldapdn="cn=dev,dc=example,dc=private" \
  -webhook="https://hooks.slack.com/services/xxxxxxxxxxxxxxxxxx" \
  -lockgained='[{{.Project}}] {{.User}}さんがデプロイ中になりました' \
  -lockreleased='[{{.Project}}] {{.User}}さんがデプロイを終了しました' \
  -lockextended='[{{.Project}}] {{.User}}さんがデプロイを終了しました' \
  -deployed='[{{.Project}}] {{.User}}さんが{{.Env}}環境にデプロイしました'
```
