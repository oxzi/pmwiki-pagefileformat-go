package pmwiki

import (
	"reflect"
	"strings"
	"testing"
)

func TestLexPageFileValid(t *testing.T) {
	input1 := ""
	items1 := []PageFileLexItem{
		{EOF, ""},
	}

	input2 := "version=pmwiki-2.1.0 urlencoded=1\ntext=Markup text\n"
	items2 := []PageFileLexItem{
		{Key, "version"},
		{Value, "pmwiki-2.1.0 urlencoded=1"},
		{Key, "text"},
		{Value, "Markup text"},
		{EOF, ""},
	}

	input3 := "version=pmwiki-2.1.0 ordered=1 urlencoded=1\ntext=This is a line.%0aThis is another.\nctime=1142030000\n"
	items3 := []PageFileLexItem{
		{Key, "version"},
		{Value, "pmwiki-2.1.0 ordered=1 urlencoded=1"},
		{Key, "text"},
		{Value, "This is a line.%0aThis is another."},
		{Key, "ctime"},
		{Value, "1142030000"},
		{EOF, ""},
	}

	input4 := "version=pmwiki-2.1.0 urlencoded=1\ntext=text\ntext:foo=foo\n"
	items4 := []PageFileLexItem{
		{Key, "version"},
		{Value, "pmwiki-2.1.0 urlencoded=1"},
		{Key, "text"},
		{Value, "text"},
		{Key, "text"},
		{KeyOpt, "foo"},
		{Value, "foo"},
		{EOF, ""},
	}

	tests := []struct {
		name  string
		input string
		items []PageFileLexItem
	}{
		{"eof", input1, items1},
		{"short", input2, items2},
		{"multiline", input3, items3},
		{"key options", input4, items4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var items []PageFileLexItem
			for item := range LexPageFile(strings.NewReader(test.input)) {
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
			for item := range LexPageFile(strings.NewReader(test.input)) {
				if item.T == Error {
					return
				}
			}
			t.Fatal("received no error")
		})
	}
}
