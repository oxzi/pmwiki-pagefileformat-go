// SPDX-FileCopyrightText: 2020 Alvar Penning
//
// SPDX-License-Identifier: GPL-3.0-or-later

package pmwiki

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// patchParser parses a Patch based on a patchLexer's channel.
type patchParser struct {
	patches []patchAction
	err     error

	lexBuff  []patchLexItem
	lexItems <-chan patchLexItem

	nextPatch *patchAction
}

type patchParseStateFunc func(*patchParser) patchParseStateFunc

// next item from the parser.
func (parser *patchParser) next() (item patchLexItem) {
	if len(parser.lexBuff) > 0 {
		item = parser.lexBuff[len(parser.lexBuff)-1]
		parser.lexBuff = parser.lexBuff[:len(parser.lexBuff)-1]
	} else {
		item = <-parser.lexItems
	}

	return
}

// backup a previously read item.
func (parser *patchParser) backup(item patchLexItem) {
	parser.lexBuff = append(parser.lexBuff, item)
}

// nextType returns the next matching item's value. A positive max value restricts the amount skipped items.
func (parser *patchParser) nextType(lexType patchLexType, max int) (v string, err error) {
	for i := 0; max <= 0 || i < max; i++ {
		if item := parser.next(); item.t == lexType {
			return item.v, nil
		} else if item.t == patchEOF {
			return "", io.EOF
		} else if item.t == patchError {
			return "", fmt.Errorf(item.v)
		}
	}
	return "", fmt.Errorf("no item with type %v found in %d messages", lexType, max)
}

// errorf stores an error and aborts.
func (parser *patchParser) errorf(format string, args ...interface{}) patchParseStateFunc {
	parser.patches = nil
	parser.err = fmt.Errorf(format, args...)

	return nil
}

// emit the current patch action.
func (parser *patchParser) emit(succ patchParseStateFunc) patchParseStateFunc {
	if parser.nextPatch != nil {
		parser.patches = append(parser.patches, *parser.nextPatch)
	}

	return succ
}

// patchParseStart parses the begin of a new item, an EOF, or an error.
func patchParseStart(parser *patchParser) patchParseStateFunc {
	switch next := parser.next(); next.t {
	case patchEOF:
		return parser.emit(nil)

	case patchError:
		return parser.errorf("%v", next.v)

	case patchRange:
		parser.backup(next)
		return parser.emit(patchParseHeader)

	default:
		return parser.errorf("invalid item %v", next)
	}
}

// patchParseHeader parses the patch header with the modified lines and mode.
func patchParseHeader(parser *patchParser) (succ patchParseStateFunc) {
	parser.nextPatch = new(patchAction)

	if startRange, err := parser.nextType(patchRange, 1); err != nil {
		return parser.errorf("%w", err)
	} else if start, err := strconv.Atoi(strings.Split(startRange, ",")[0]); err != nil {
		return parser.errorf("cannot parse start range, %w", err)
	} else {
		parser.nextPatch.startLine = start
	}

	if mode, err := parser.nextType(patchMode, 1); err != nil {
		return parser.errorf("%w", err)
	} else {
		switch mode {
		case "a":
			parser.nextPatch.mode = addition
			succ = patchParseAddition
		case "d":
			parser.nextPatch.mode = deletion
			succ = patchParseDeletion
		case "c":
			parser.nextPatch.mode = change
			succ = patchParseDeletion
		default:
			return parser.errorf("invalid mode value %s", mode)
		}
	}

	if _, err := parser.nextType(patchRange, 1); err != nil {
		return parser.errorf("%w", err)
	}

	return
}

// patchParseAddition parses addition lines.
func patchParseAddition(parser *patchParser) patchParseStateFunc {
	next := parser.next()
	for ; next.t == patchAddition; next = parser.next() {
		parser.nextPatch.additionLines = append(parser.nextPatch.additionLines, next.v)
	}

	parser.backup(next)
	return patchParseStart
}

// patchParseDeletion parses deletion lines and might switch to patchParseAddition in case of a change patchAction.
func patchParseDeletion(parser *patchParser) patchParseStateFunc {
	next := parser.next()
	for ; next.t == patchDeletion; next = parser.next() {
		parser.nextPatch.deletionLines = append(parser.nextPatch.deletionLines, next.v)
	}

	parser.backup(next)
	if parser.nextPatch.mode == change && next.t == patchAddition {
		return patchParseAddition
	} else {
		return patchParseStart
	}
}

// parsePatch parses a Patch from an input string.
func parsePatch(data string) (Patch, error) {
	parser := &patchParser{lexItems: lexPatch(data)}
	for state := patchParseStart; state != nil; state = state(parser) {
	}

	return parser.patches, parser.err
}
