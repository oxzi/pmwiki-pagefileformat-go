package pmwiki

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// PageFileRevision describes a PageFile internal diff.
type PageFileRevision struct {
	Time   time.Time
	Author string
	Host   net.IP

	Diff        Patch
	DiffAgainst time.Time
}

// PageFile describes a PmWiki page including its history.
type PageFile struct {
	Version string
	Name    string

	Time   time.Time
	Text   string
	Author string
	Host   net.IP
	Rev    int

	Revs map[time.Time]PageFileRevision

	Deleted time.Time
}

// Revisions calls a function with a "view" copy of each revision of this PageFile.
//
// An error is returned when creating the successive revision fails. However, multiple previous revisions could be
// generated previously.
func (pageFile PageFile) Revisions(callback func(view PageFile)) error {
	if pageFile.Deleted != (time.Time{}) {
		defer callback(PageFile{
			Version: pageFile.Version,
			Name:    pageFile.Name,
			Time:    pageFile.Deleted,
			Text:    "",
			Author:  "", // the real author is written in the RecentChanges file
			Host:    net.ParseIP("::1"),
			Rev:     pageFile.Rev + 1,
		})
	}

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
