package pmwiki

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestParsePageFileValid(t *testing.T) {
	input1 := "version=pmwiki-2.1.0 urlencoded=1\n"
	check1 := func(pf PageFile) bool { return pf.Version == "pmwiki-2.1.0 urlencoded=1" }

	input2 := "version=pmwiki-2.1.0 urlencoded=1\ntext=Markup text\n"
	check2 := func(pf PageFile) bool { return pf.Text == "Markup text" }

	input3 := "version=pmwiki-2.1.0 urlencoded=1\nauthor=user\nhost=2001:db8::1\nname=Main.Test\nrev=42\ntime=1603541891\ntext=%0ahello%0aworld\n"
	check3 := func(pf PageFile) bool {
		return pf.Version == "pmwiki-2.1.0 urlencoded=1" &&
			pf.Author == "user" &&
			pf.Host.Equal(net.ParseIP("2001:db8::1")) &&
			pf.Name == "Main.Test" &&
			pf.Rev == 42 &&
			pf.Time.Equal(time.Unix(1603541891, 0).UTC()) &&
			pf.Text == "\nhello\nworld"
	}

	input4 := "version=pmwiki-2.1.0 urlencoded=1\nrandom=data\n"
	check4 := func(PageFile) bool { return true }

	input5 := "version=pmwiki-2.1.0 urlencoded=1\ntext=text\nhost:42=::1\ndiff:42:23:=0a1%0a> add%0a\n"
	check5 := func(pf PageFile) bool {
		pfrUnix := time.Unix(42, 0).UTC()
		if pfr, ok := pf.Revs[pfrUnix]; !ok {
			return false
		} else {
			return pfr.Host.Equal(net.ParseIP("::1")) &&
				pfr.DiffAgainst.Equal(time.Unix(23, 0).UTC())
		}
	}

	input6 := "version=pmwiki-2.1.0 urlencoded=1\nauthor=user\nhost=2001:db8::1\nname=Main.Test\nrev=42\ntime=1603541891\ntext=%0ahello%0aworld\n" +
		"author:10=foo\nhost:10=::1\ndiff:10:5:=0a1%0a> A%0a\nauthor:5=bar\nhost:5=::2\ndiff:5:3:=1d2%0a%3c B%0a\n"
	check6 := func(pf PageFile) bool {
		mainCheck := pf.Version == "pmwiki-2.1.0 urlencoded=1" &&
			pf.Author == "user" &&
			pf.Host.Equal(net.ParseIP("2001:db8::1")) &&
			pf.Name == "Main.Test" &&
			pf.Rev == 42 &&
			pf.Time.Equal(time.Unix(1603541891, 0).UTC()) &&
			pf.Text == "\nhello\nworld"
		if !mainCheck {
			return false
		}

		if len(pf.Revs) != 2 {
			return false
		}

		pfrB := time.Unix(10, 0).UTC()
		pfrA := time.Unix(5, 0).UTC()

		if pfr, ok := pf.Revs[pfrB]; !ok {
			return false
		} else {
			chk := pfr.Author == "foo" &&
				pfr.Host.Equal(net.ParseIP("::1")) &&
				pfr.DiffAgainst.Equal(pfrA)
			if !chk {
				return false
			}
		}

		if pfr, ok := pf.Revs[pfrA]; !ok {
			return false
		} else {
			return pfr.Author == "bar" &&
				pfr.Host.Equal(net.ParseIP("::2")) &&
				pfr.DiffAgainst.Equal(time.Unix(3, 0).UTC())
		}
	}

	input7 := "version=pmwiki-2.1.0 urlencoded=1\ntext=text\nfoo:bar=buz\n"
	check7 := func(PageFile) bool { return true }

	input8 := "version=pmwiki-2.1.0 urlencoded=1\ntext=text\nfoo:23=buz\n"
	check8 := func(pf PageFile) bool { return len(pf.Revs) == 0 }

	input9 := "version=pmwiki-2.1.0 urlencoded=1\ntext=foo%0abar\n"
	check9 := func(pf PageFile) bool { return pf.Text == "foo\nbar" }

	input10 := "version=pmwiki-2.1.0\ntext=foo%0abar\n"
	check10 := func(pf PageFile) bool { return pf.Text == "foo%0abar" }

	input11 := "version=pmwiki-2.1.0\ntext=foo\nauthor:1527448031=user\ndiff:1527448031:1527446923:=\nhost:1527448031=fc80::1\n"
	check11 := func(pf PageFile) bool {
		return len(pf.Revs) == 1 && pf.Revs[time.Unix(1527448031, 0).UTC()].DiffAgainst != (time.Time{})
	}

	tests := []struct {
		name  string
		input string
		check func(PageFile) bool
	}{
		{"version only", input1, check1},
		{"text field", input2, check2},
		{"all main fields", input3, check3},
		{"ignore unsupported items", input4, check4},
		{"simple revision", input5, check5},
		{"two revisions", input6, check6},
		{"ignore non-timestamp keyopts", input7, check7},
		{"ignore non supported revision key", input8, check8},
		{"check URL encoding", input9, check9},
		{"check disabled URL encoding", input10, check10},
		{"empty diff", input11, check11},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if pf, err := ParsePageFile(strings.NewReader(test.input)); err != nil {
				t.Fatal(err)
			} else if !test.check(pf) {
				t.Fatal("check failed")
			}
		})
	}
}

func TestParsePageFileInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"not starting with version", "foo=bar\nversion=pmwiki-2.1.0 urlencoded=1\n"},
		{"version with keyopts", "version:nope=pmwiki-23.42\n"},
		{"early eof", "version=pmwiki-2.1.0 urlencoded=1\ntext=Markup text"},
		{"double name", "version=pmwiki-2.1.0 urlencoded=1\nname=foo\nname=bar\n"},
		{"double time", "version=pmwiki-2.1.0 urlencoded=1\ntime=123456\ntime=1234567\n"},
		{"invalid time", "version=pmwiki-2.1.0 urlencoded=1\ntime=0xacab\n"},
		{"double text", "version=pmwiki-2.1.0 urlencoded=1\ntext=foo\ntext=bar\n"},
		{"double author", "version=pmwiki-2.1.0 urlencoded=1\nauthor=foo\nauthor=bar\n"},
		{"double host", "version=pmwiki-2.1.0 urlencoded=1\nhost=2001:db8::1\nhost=172.23.42.128\n"},
		{"invalid host", "version=pmwiki-2.1.0 urlencoded=1\nhost=dtn://host/\n"},
		{"double rev", "version=pmwiki-2.1.0 urlencoded=1\nrev=1\nrev=2\n"},
		{"invalid rev", "version=pmwiki-2.1.0 urlencoded=1\nrev=latest and greatest\n"},
		{"rev, double author", "version=pmwiki-2.1.0 urlencoded=1\nauthor:23=foo\nauthor:23=bar\n"},
		{"rev, double host", "version=pmwiki-2.1.0 urlencoded=1\nhost:23=2001:db8::1\nhost:23=172.23.42.128\n"},
		{"rev, invalid host", "version=pmwiki-2.1.0 urlencoded=1\nhost:23=dtn://host/\n"},
		{"rev, double diff", "version=pmwiki-2.1.0 urlencoded=1\ndiff:42:23:=foo\ndiff:42:23:=bar\n"},
		{"rev, diff less keyopts", "version=pmwiki-2.1.0 urlencoded=1\ndiff:42=foo\n"},
		{"rev, diff invalid against", "version=pmwiki-2.1.0 urlencoded=1\ndiff:42:old:=foo\n"},
		{"rev, diff invalid", "version=pmwiki-2.1.0 urlencoded=1\ndiff:42:23:=AAAAAAAAAA\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if pf, err := ParsePageFile(strings.NewReader(test.input)); err == nil {
				t.Fatalf("did not fail, produced %v", pf)
			}
		})
	}
}
