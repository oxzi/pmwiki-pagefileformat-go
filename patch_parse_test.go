package pmwiki

import (
	"testing"
)

func TestParsePatchValid(t *testing.T) {
	input1 := "\n"
	check1 := func(patch Patch) bool { return len(patch) == 0 }

	input2 := "0a1\n> addition\n"
	check2 := func(patch Patch) bool {
		return len(patch) == 1 && patch[0].mode == addition && patch[0].startLine == 0 &&
			len(patch[0].additionLines) == 1 && patch[0].additionLines[0] == "addition"
	}

	input3 := "0a1,3\n> multiline\n> addition\n> yay\n"
	check3 := func(patch Patch) bool {
		return len(patch) == 1 && patch[0].mode == addition && patch[0].startLine == 0 && len(patch[0].additionLines) == 3
	}

	input4 := "23d23\n< gone\n"
	check4 := func(patch Patch) bool {
		return len(patch) == 1 && patch[0].mode == deletion && patch[0].startLine == 23 &&
			len(patch[0].deletionLines) == 1 && patch[0].deletionLines[0] == "gone"
	}

	input5 := "5c5\n< foo\n---\n> bar\n"
	check5 := func(patch Patch) bool {
		return len(patch) == 1 && patch[0].mode == change && patch[0].startLine == 5 &&
			len(patch[0].deletionLines) == 1 && patch[0].deletionLines[0] == "foo" &&
			len(patch[0].additionLines) == 1 && patch[0].additionLines[0] == "bar"
	}

	// Example from <https://en.wikipedia.org/wiki/Diff#Usage>
	input6 := "0a1,6\n> This is an important\n> notice! It should\n> therefore be located at\n" +
		"> the beginning of this\n> document!\n>\n11,15d16\n< This paragraph contains\n" +
		"< text that is outdated.\n< It will be deleted in the\n< near future.\n<\n17c18\n" +
		"< check this dokument. On\n---\n> check this document. On\n24a26,29\n>\n" +
		"> This paragraph contains\n> important new additions\n> to this document.\n"
	check6 := func(patch Patch) bool {
		return len(patch) == 4 &&
			patch[0].mode == addition && patch[0].startLine == 0 && len(patch[0].additionLines) == 6 &&
			patch[1].mode == deletion && patch[1].startLine == 11 && len(patch[1].deletionLines) == 5 &&
			patch[2].mode == change && patch[2].startLine == 17 && len(patch[2].additionLines) == 1 && len(patch[2].deletionLines) == 1 &&
			patch[3].mode == addition && patch[3].startLine == 24 && len(patch[3].additionLines) == 4
	}

	tests := []struct {
		name  string
		input string
		check func(Patch) bool
	}{
		{"empty patch", input1, check1},
		{"single addition", input2, check2},
		{"multiline addition", input3, check3},
		{"single deletion", input4, check4},
		{"single change", input5, check5},
		{"Wikipedia example", input6, check6},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if patch, err := parsePatch(test.input); err != nil {
				t.Fatal(err)
			} else if !test.check(patch) {
				t.Fatal("check failed")
			}
		})
	}
}

func TestParsePatchInvalid(t *testing.T) {
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
		{"addition only", "> addition\n"},
		{"deletion only", "< deletion\n"},
		{"double dash only", "--\n"},
		{"range-mode-range-mode", "0a1a1\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if patch, err := parsePatch(test.input); err == nil {
				t.Fatalf("did not fail, produced %v", patch)
			}
		})
	}
}
