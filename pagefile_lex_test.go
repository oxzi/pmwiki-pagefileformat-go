// SPDX-FileCopyrightText: 2020 Alvar Penning
//
// SPDX-License-Identifier: GPL-3.0-or-later

package pmwiki

import (
	"reflect"
	"strings"
	"testing"
)

func TestLexPageFileValid(t *testing.T) {
	input1 := ""
	items1 := []pageFileLexItem{
		{pageFileEOF, ""},
	}

	input2 := "version=pmwiki-2.1.0 urlencoded=1\ntext=Markup text\n"
	items2 := []pageFileLexItem{
		{pageFileKey, "version"},
		{pageFileValue, "pmwiki-2.1.0 urlencoded=1"},
		{pageFileKey, "text"},
		{pageFileValue, "Markup text"},
		{pageFileEOF, ""},
	}

	input3 := "version=pmwiki-2.1.0 ordered=1 urlencoded=1\ntext=This is a line.%0aThis is another.\nctime=1142030000\n"
	items3 := []pageFileLexItem{
		{pageFileKey, "version"},
		{pageFileValue, "pmwiki-2.1.0 ordered=1 urlencoded=1"},
		{pageFileKey, "text"},
		{pageFileValue, "This is a line.%0aThis is another."},
		{pageFileKey, "ctime"},
		{pageFileValue, "1142030000"},
		{pageFileEOF, ""},
	}

	input4 := "version=pmwiki-2.1.0 urlencoded=1\ntext=text\ntext:foo=foo\n"
	items4 := []pageFileLexItem{
		{pageFileKey, "version"},
		{pageFileValue, "pmwiki-2.1.0 urlencoded=1"},
		{pageFileKey, "text"},
		{pageFileValue, "text"},
		{pageFileKey, "text"},
		{pageFileKeyOpt, "foo"},
		{pageFileValue, "foo"},
		{pageFileEOF, ""},
	}

	input5 := "version=pmwiki-2.1.0 urlencoded=1\ntext=text\ntext:foo:bar=foo\n"
	items5 := []pageFileLexItem{
		{pageFileKey, "version"},
		{pageFileValue, "pmwiki-2.1.0 urlencoded=1"},
		{pageFileKey, "text"},
		{pageFileValue, "text"},
		{pageFileKey, "text"},
		{pageFileKeyOpt, "foo"},
		{pageFileKeyOpt, "bar"},
		{pageFileValue, "foo"},
		{pageFileEOF, ""},
	}

	input6 := "author:1527448031=user\ndiff:1527448031:1527446923:=\nhost:1527448031=fc80::1\n"
	items6 := []pageFileLexItem{
		{pageFileKey, "author"},
		{pageFileKeyOpt, "1527448031"},
		{pageFileValue, "user"},
		{pageFileKey, "diff"},
		{pageFileKeyOpt, "1527448031"},
		{pageFileKeyOpt, "1527446923"},
		{pageFileKeyOpt, ""},
		{pageFileValue, ""},
		{pageFileKey, "host"},
		{pageFileKeyOpt, "1527448031"},
		{pageFileValue, "fc80::1"},
		{pageFileEOF, ""},
	}

	tests := []struct {
		name  string
		input string
		items []pageFileLexItem
	}{
		{"eof", input1, items1},
		{"short", input2, items2},
		{"multiline", input3, items3},
		{"key options", input4, items4},
		{"multiple key options", input5, items5},
		{"empty diff", input6, items6},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var items []pageFileLexItem
			for item := range lexPageFile(strings.NewReader(test.input)) {
				items = append(items, item)
			}

			if items == nil {
				t.Fatal("there are no items")
			}
			if len(test.items) != len(items) {
				t.Fatalf("length mismatches")
			}

			for i := 0; i < len(test.items); i++ {
				if !reflect.DeepEqual(test.items[i], items[i]) {
					t.Fatalf("%d, %v != %v", i, test.items[i], items[i])
				}
			}
		})
	}
}

func TestLexPageFileInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"newline", "\n"},
		{"key-whitespace", " "},
		{"key-abort", "name"},
		{"keyopt-abort", "name:foo"},
		{"keyopt-whitespace", "name: "},
		{"val-abort", "name=foo"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for item := range lexPageFile(strings.NewReader(test.input)) {
				if item.t == pageFileError {
					return
				}
			}
			t.Fatal("received no error")
		})
	}
}
