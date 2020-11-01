package pmwiki

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// author:${DATE_NOW}
// host:${DATE_NOW}
// diff:${DATE_NOW}:${DATE_PREV}:${DIFF_CLASS}
type PageFileRevision struct {
	Time   time.Time
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

// Revisions calls a function with a "view" copy of each revision of this PageFile.
func (pageFile PageFile) Revisions(callback func(view PageFile)) error {
	if len(pageFile.Revs) == 0 {
		callback(pageFile)
		return nil
	}

	text := pageFile.Text
	revNo := pageFile.Rev

	for rev, ok := pageFile.Revs[pageFile.Time]; text != ""; rev, ok = pageFile.Revs[rev.DiffAgainst] {
		if !ok {
			return fmt.Errorf("revision %d is missing", revNo)
		}

		curr := PageFile{
			Version: pageFile.Version,
			Name:    pageFile.Name,
			Time:    rev.Time,
			Text:    text,
			Author:  rev.Author,
			Host:    rev.Host,
			Rev:     revNo,
		}
		callback(curr)

		revNo--

		var textOut strings.Builder
		if err := rev.Diff.Apply(strings.NewReader(text), &textOut); err != nil {
			return err
		}
		text = textOut.String()
	}

	return nil
}
