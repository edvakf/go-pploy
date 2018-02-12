package models

import "time"

// Lock is project's deployment lock
type Lock struct {
	User    string    `json:"user"`
	EndTime time.Time `json:"endTime"`
}
