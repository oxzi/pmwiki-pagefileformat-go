package pmwiki

import (
	"net"
	"time"
)

// author:${DATE_NOW}
// host:${DATE_NOW}
// diff:${DATE_NOW}:${DATE_PREV}:${DIFF_CLASS}
type PageFileRevision struct {
	Author string
	Host   net.IP

	Diff        Patch
	DiffAgainst time.Time
}

type PageFile struct {
	Version string
	Name    string

	Time   time.Time
	Text   string
	Author string
	Host   net.IP
	Rev    int

	Revs map[time.Time]PageFileRevision
}
