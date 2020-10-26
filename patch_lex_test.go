package pmwiki

import (
	"reflect"
	"testing"
)

func TestLexPatchValid(t *testing.T) {
	input1 := ""
	items1 := []patchLexItem{
		{patchEOF, ""},
	}

	input2 := "\n\n\n\n"
	items2 := []patchLexItem{
		{patchEOF, ""},
	}

	input3 := "0a1\n"
	items3 := []patchLexItem{
		{patchRange, "0"},
		{patchMode, "a"},
		{patchRange, "1"},
		{patchEOF, ""},
	}

	input4 := "0a1,5\n"
	items4 := []patchLexItem{
		{patchRange, "0"},
		{patchMode, "a"},
		{patchRange, "1,5"},
		{patchEOF, ""},
	}

	input5 := "0a1\n> addition\n"
	items5 := []patchLexItem{
		{patchRange, "0"},
		{patchMode, "a"},
		{patchRange, "1"},
		{patchAddition, "addition"},
		{patchEOF, ""},
	}

	input6 := "0a1,3\n> multiline\n> addition\n> yay\n"
	items6 := []patchLexItem{
		{patchRange, "0"},
		{patchMode, "a"},
		{patchRange, "1,3"},
		{patchAddition, "multiline"},
		{patchAddition, "addition"},
		{patchAddition, "yay"},
		{patchEOF, ""},
	}

	input7 := "23d23\n< gone\n"
	items7 := []patchLexItem{
		{patchRange, "23"},
		{patchMode, "d"},
		{patchRange, "23"},
		{patchDeletion, "gone"},
		{patchEOF, ""},
	}

	input8 := "5c5\n< foo\n---\n> bar\n"
	items8 := []patchLexItem{
		{patchRange, "5"},
		{patchMode, "c"},
		{patchRange, "5"},
		{patchDeletion, "foo"},
		{patchAddition, "bar"},
		{patchEOF, ""},
	}

	// Example from <https://en.wikipedia.org/wiki/Diff#Usage>
	input9 := "0a1,6\n> This is an important\n> notice! It should\n> therefore be located at\n" +
		"> the beginning of this\n> document!\n>\n11,15d16\n< This paragraph contains\n" +
		"< text that is outdated.\n< It will be deleted in the\n< near future.\n<\n17c18\n" +
		"< check this dokument. On\n---\n> check this document. On\n24a26,29\n>\n" +
		"> This paragraph contains\n> important new additions\n> to this document.\n"
	items9 := []patchLexItem{
		{patchRange, "0"},
		{patchMode, "a"},
		{patchRange, "1,6"},
		{patchAddition, "This is an important"},
		{patchAddition, "notice! It should"},
		{patchAddition, "therefore be located at"},
		{patchAddition, "the beginning of this"},
		{patchAddition, "document!"},
		{patchAddition, ""},
		{patchRange, "11,15"},
		{patchMode, "d"},
		{patchRange, "16"},
		{patchDeletion, "This paragraph contains"},
		{patchDeletion, "text that is outdated."},
		{patchDeletion, "It will be deleted in the"},
		{patchDeletion, "near future."},
		{patchDeletion, ""},
		{patchRange, "17"},
		{patchMode, "c"},
		{patchRange, "18"},
		{patchDeletion, "check this dokument. On"},
		{patchAddition, "check this document. On"},
		{patchRange, "24"},
		{patchMode, "a"},
		{patchRange, "26,29"},
		{patchAddition, ""},
		{patchAddition, "This paragraph contains"},
		{patchAddition, "important new additions"},
		{patchAddition, "to this document."},
		{patchEOF, ""},
	}

	tests := []struct {
		name  string
		input string
		items []patchLexItem
	}{
		{"empty input", input1, items1},
		{"newlines", input2, items2},
		{"simple header", input3, items3},
		{"simple range header", input4, items4},
		{"simple addition", input5, items5},
		{"multiline addition", input6, items6},
		{"simple deletion", input7, items7},
		{"simple change", input8, items8},
		{"Wikipedia example", input9, items9},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var items []patchLexItem
			for item := range lexPatch(test.input) {
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

func TestLexPatchInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"early eof", "1"},
		{"invalid range", "1-2"},
		{"double comma range", "1,2,3"},
		{"invalid mode", "1x"},
		{"invalid beginning", "| foo"},
		{"invalid addition", ">A"},
		{"invalid deletion", "<A"},
		{"addition early end", "> addition"},
		{"deletion early end", "< addition"},
		{"double dash", "--"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for item := range lexPatch(test.input) {
				if item.t == patchError {
					return
				}
			}
			t.Fatal("received no error")
		})
	}
}
