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
			pf.Text == "%0ahello%0aworld"
	}

	input4 := "version=pmwiki-2.1.0 urlencoded=1\nrandom=data\n"
	check4 := func(PageFile) bool { return true }

	tests := []struct {
		name  string
		input string
		check func(PageFile) bool
	}{
		{"version only", input1, check1},
		{"text field", input2, check2},
		{"all main fields", input3, check3},
		{"ignore unsupported items", input4, check4},
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if pf, err := ParsePageFile(strings.NewReader(test.input)); err == nil {
				t.Fatalf("did not fail, produced %v", pf)
			}
		})
	}
}
