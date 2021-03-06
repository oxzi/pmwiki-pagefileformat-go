// SPDX-FileCopyrightText: 2020 Alvar Penning
//
// SPDX-License-Identifier: GPL-3.0-or-later

package pmwiki

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

// pageFileLexType are the different tokens of a pageFileLexItem.
type pageFileLexType int

const (
	_ pageFileLexType = iota

	// pageFileEOF is an internal io.EOF.
	pageFileEOF
	// pageFileError in case of an invalid state transition.
	pageFileError

	// pageFileKey name.
	pageFileKey
	// pageFileKeyOpt are additional options for the previous pageFileKey.
	pageFileKeyOpt
	// pageFileValue for the previous pageFileKey.
	pageFileValue
)

// pageFileLexItem is a tuple of a pageFileLexType with its value.
type pageFileLexItem struct {
	t pageFileLexType
	v string
}

// pageFileLexer is a lexer to tokenize PmWiki's PageFileFormat.
//
// Its logic and code structure is heavily inspired by Rob Pike's "Lexical Scanning in Go" talk,
// <https://talks.golang.org/2011/lex.slide>. However, it operates on an io.Reader instead of a string.
type pageFileLexer struct {
	reader *bufio.Reader
	items  chan pageFileLexItem
}

// pageFileLexStateFunc is lexing a pageFileLexer and returns its successive pageFileLexStateFunc.
type pageFileLexStateFunc func(*pageFileLexer) pageFileLexStateFunc

// pageFileLexBegin inspects a line's start and selects between an pageFileEOF or a pageFileKey.
func pageFileLexBegin(lexer *pageFileLexer) pageFileLexStateFunc {
	if _, err := lexer.next(); err == nil {
		lexer.backup()
		return pageFileLexKey
	} else if err == io.EOF {
		return lexer.eof()
	} else {
		return lexer.errorf("%v", err)
	}
}

var (
	// pageFileLexKey extracts a pageFileKey.
	pageFileLexKey pageFileLexStateFunc

	// pageFileLexKeyOpt extracts a pageFileKey's pageFileKeyOpt.
	pageFileLexKeyOpt pageFileLexStateFunc
)

func init() {
	// The two lexers for pageFileKey and pageFileKeyOpt are almost identical. Thus, the pageFileLexKeyOrKeyOptGenerator
	// creates them. However, within the generator, the pageFileLexKeyOpt is referenced. This level of indirection is
	// too much for the Go compiler. That's why this hacky hack is here.

	pageFileLexKey = pageFileLexKeyOrKeyOptGenerator(pageFileKey)
	pageFileLexKeyOpt = pageFileLexKeyOrKeyOptGenerator(pageFileKeyOpt)
}

// pageFileLexKeyOrKeyOptGenerator generates pageFileLexKey and pageFileLexKeyOpt.
func pageFileLexKeyOrKeyOptGenerator(t pageFileLexType) pageFileLexStateFunc {
	return func(lexer *pageFileLexer) pageFileLexStateFunc {
		var field string
		for {
			r, err := lexer.next()
			if err != nil {
				return lexer.errorf("%v", err)
			}

			switch r {
			case ':':
				return lexer.emit(t, field, pageFileLexKeyOpt)

			case '=':
				return lexer.emit(t, field, pageFileLexVal)

			default:
				if unicode.IsSpace(r) {
					return lexer.errorf("unexpected white space")
				}
				field += string(r)
			}
		}
	}
}

// pageFileLexVal extracts a pageFileKey's pageFileValue.
func pageFileLexVal(lexer *pageFileLexer) pageFileLexStateFunc {
	var field string
	for {
		r, err := lexer.next()
		if err != nil {
			return lexer.errorf("%v", err)
		}

		if r == '\n' {
			return lexer.emit(pageFileValue, field, pageFileLexBegin)
		}
		field += string(r)
	}
}

// lexPageFile starts a lexical analysis for PmWiki's PageFileFormat. The tokens are sent to the channel.
func lexPageFile(reader io.Reader) <-chan pageFileLexItem {
	lexer := &pageFileLexer{
		reader: bufio.NewReader(reader),
		items:  make(chan pageFileLexItem),
	}
	go lexer.run()

	return lexer.items
}

// run the pageFileLexer's states.
func (lexer *pageFileLexer) run() {
	for state := pageFileLexBegin; state != nil; state = state(lexer) {
	}
	close(lexer.items)
}

// next rune from the underlying buffer.
func (lexer *pageFileLexer) next() (r rune, err error) {
	r, _, err = lexer.reader.ReadRune()
	return
}

// backup the last rune.
func (lexer *pageFileLexer) backup() {
	if err := lexer.reader.UnreadRune(); err != nil {
		panic(err)
	}
}

// emit a token back and return the successive state.
func (lexer *pageFileLexer) emit(t pageFileLexType, v string, succ pageFileLexStateFunc) pageFileLexStateFunc {
	lexer.items <- pageFileLexItem{t, v}
	return succ
}

// errorf emits an error back.
func (lexer *pageFileLexer) errorf(format string, args ...interface{}) pageFileLexStateFunc {
	return lexer.emit(pageFileError, fmt.Sprintf(format, args...), nil)
}

// eof emits an pageFileEOF back.
func (lexer *pageFileLexer) eof() pageFileLexStateFunc {
	return lexer.emit(pageFileEOF, "", nil)
}
