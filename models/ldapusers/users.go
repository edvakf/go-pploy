package ldapusers

import (
	"fmt"
	"log"
	"time"

	ldap "gopkg.in/ldap.v2"
)

// Config is LDAP config
type Config struct {
	Host     string
	Port     int
	BaseDN   string
	CacheTTL time.Duration
}

var config Config

// SetConfig updates LDAP config
func SetConfig(c Config) {
	config = c
}

var users []string

var nextUpdate time.Time

// All reloads user list from ldap if config is set and returns and caches them
func All() []string {
	if config.Host == "" || config.BaseDN == "" {
		return users
	}
	if time.Now().After(nextUpdate) {
		u, err := fetch(config.Host, config.Port, config.BaseDN)
		if err != nil {
			// deliberately miss error
			log.Println(err)
		} else {
			nextUpdate = time.Now().Add(config.CacheTTL)
			users = u
		}
	}
	return users
}

func fetch(host string, port int, baseDN string) ([]string, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		baseDN, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(cn=*)",       // The filter to apply
		[]string{"cn"}, // A list attributes to retrieve
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	u := []string{}
	for _, entry := range sr.Entries {
		u = append(u, entry.GetAttributeValue("cn"))
	}

	return u, nil
}
