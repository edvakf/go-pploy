package models

import "time"

// Lock is project's deployment lock
type Lock struct {
	User    string    `json:"user"`
	EndTime time.Time `json:"endTime"`
}

// Commit is the structured git commit object
type Commit struct {
	Hash       string    `json:"hash"`
	Time       time.Time `json:"time"`
	Author     string    `json:"author"`
	OtherRefs  []string  `json:"otherRefs"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	NameStatus string    `json:"nameStatus"`
}
