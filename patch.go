// SPDX-FileCopyrightText: 2020 Alvar Penning
//
// SPDX-License-Identifier: GPL-3.0-or-later

package pmwiki

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// patchType describes the change of a patchAction.
type patchType int

const (
	_ patchType = iota

	// addition of lines from a Patch.
	addition
	// deletion of lines from a Patch.
	deletion
	// change of lines from a Patch, combined addition and deletion.
	change
)

// patchAction is one change within a Patch.
type patchAction struct {
	mode          patchType
	startLine     int
	additionLines []string
	deletionLines []string
}

// apply this patchAction. Will be called from Patch.Apply.
func (patchAction patchAction) apply(scanner *bufio.Scanner, out io.Writer) (consumed int, err error) {
	if patchAction.mode == deletion || patchAction.mode == change {
		for consumed = 0; consumed < len(patchAction.deletionLines) && scanner.Scan(); consumed++ {
			// Sometimes PmWiki truncates whitespaces from patches, because oh̨ m͞y gǫd ͜p̀mwi͝ki s̡̧͝t͏a̢̧̛h̀p̕ ͝w͏̸̵h͢a͢͞t̴̨̀ a̧̧re ̢͠y̨͠o̧͏ư̴̴ d͡o̴i͜ng̕?̸̕͞!̡͠!
			input := scanner.Text()
			expected := patchAction.deletionLines[consumed]

			if input != expected && input != strings.TrimSpace(expected) {
				err = fmt.Errorf("patch:%d expected \"%s\", got \"%s\"", consumed, patchAction.deletionLines[consumed], input)
				return
			}
		}
		if scannerErr := scanner.Err(); scannerErr != nil {
			err = scannerErr
			return
		}
	}

	if patchAction.mode == addition || patchAction.mode == change {
		for _, line := range patchAction.additionLines {
			if _, err = fmt.Fprint(out, line, "\n"); err != nil {
				return
			}
		}
	}

	return
}

// Patch is the difference between two revisions stored as a `diff`.
type Patch []patchAction

// Apply this Patch to an input stream and write the patched result back to an output stream.
func (patch Patch) Apply(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)

	for patchNo, line := 0, 0; ; {
		// Deletion / Change first
		if patchNo < len(patch) && patch[patchNo].startLine == line && (patch[patchNo].mode == deletion || patch[patchNo].mode == change) {
			if consumed, err := patch[patchNo].apply(scanner, out); err != nil {
				return err
			} else {
				patchNo++
				line += consumed
			}
		}

		// Consume a line, unless we have not started reading, e.g., for an addition patch starting at line zero
		if line > 0 {
			if !scanner.Scan() {
				return scanner.Err()
			} else if _, err := fmt.Fprint(out, scanner.Text(), "\n"); err != nil {
				return err
			}
		}

		// Addition second
		if patchNo < len(patch) && patch[patchNo].startLine == line && patch[patchNo].mode == addition {
			if _, err := patch[patchNo].apply(scanner, out); err != nil {
				return err
			} else {
				patchNo++
			}
		}

		line++
	}
}
