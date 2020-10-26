package pmwiki

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

// patchLexType are the different tokens of a patchLexItem.
type patchLexType int

const (
	_ patchLexType = iota

	// patchEOF is an internal io.EOF.
	patchEOF
	// patchError in case of an invalid state transition.
	patchError

	// patchRange within the header, e.g., 2 or 2,5.
	patchRange
	// patchMode within the header, one of a, d, or c.
	patchMode
	// patchAddition line for an addition.
	patchAddition
	// patchDeletion line for a deletion.
	patchDeletion
)

// patchLexItem is a tuple of a patchLexType with its value.
type patchLexItem struct {
	t patchLexType
	v string
}

// patchLexer is a lexer to tokenize a traditional Unix diff file.
//
// Its logic and code structure is heavily inspired by Rob Pike's "Lexical Scanning in Go" talk,
// <https://talks.golang.org/2011/lex.slide>.
type patchLexer struct {
	data string

	start int
	pos   int
	width int

	items chan patchLexItem
}

// patchLexStateFunc is lexing a patchLexer and returns its successive patchLexStateFunc.
type patchLexStateFunc func(*patchLexer) patchLexStateFunc

// patchLexBegin inspects the beginning of a line and dispatches the specific lexing state function.
func patchLexBegin(lexer *patchLexer) patchLexStateFunc {
	next, err := lexer.next()
	if err == io.EOF {
		return lexer.eof()
	} else if err != nil {
		return lexer.errorf("next element at new line errored, %v", err)
	}

	if next == '\n' {
		lexer.ignore()
		return patchLexBegin
	}

	defer lexer.backup()

	switch {
	case unicode.IsDigit(next):
		return patchLexRange
	case next == '>':
		return patchLexAdd
	case next == '<':
		return patchLexDel
	case next == '-':
		return patchLexDash
	default:
		return lexer.errorf("unsupported element %c", next)
	}
}

// patchLexRange reads a line range within a diff / patch header.
func patchLexRange(lexer *patchLexer) patchLexStateFunc {
	hadComma := false

	for {
		next, err := lexer.next()
		if err != nil {
			return lexer.errorf("%v", err)
		}

		if unicode.IsDigit(next) {
			continue
		} else if next == ',' {
			if hadComma {
				return lexer.errorf("range did already contain one comma so far")
			}
			hadComma = true
		} else if next == 'a' || next == 'd' || next == 'c' {
			lexer.backup()
			return lexer.emit(patchRange, patchLexMode)
		} else if next == '\n' {
			lexer.backup()
			return lexer.emit(patchRange, patchLexBegin)
		} else {
			return lexer.errorf("unsupported char %c within range", next)
		}
	}
}

// patchLexMode reads the mode within a header.
func patchLexMode(lexer *patchLexer) patchLexStateFunc {
	next, err := lexer.next()
	if err != nil {
		return lexer.errorf("%v", err)
	}

	if next == 'a' || next == 'd' || next == 'c' {
		return lexer.emit(patchMode, patchLexRange)
	} else {
		return lexer.errorf("unsupported mode char %c", next)
	}
}

// patchLexDash consumes the three dashes between a change's insertion and deletion.
func patchLexDash(lexer *patchLexer) patchLexStateFunc {
	if err := lexer.expect("---\n"); err != nil {
		return lexer.errorf("dash line errored, %v", err)
	}

	lexer.ignore()
	return patchLexBegin
}

var (
	// patchLexAdd reads patchAddition lines.
	patchLexAdd patchLexStateFunc

	// patchLexDel reads patchDeletion lines.
	patchLexDel patchLexStateFunc
)

func init() {
	// The patchLexAdd and patchLexAdd functions are almost identical. Thus, the patchLexAddOrDelGenerator function
	// generates them. However, this level of indirection is too much for the Go compiler.

	patchLexAdd = patchLexAddOrDelGenerator(patchAddition)
	patchLexDel = patchLexAddOrDelGenerator(patchDeletion)
}

// patchLexAddOrDelGenerator generates a patchLexStateFunc for patchLexAdd and patchLexDel.
func patchLexAddOrDelGenerator(t patchLexType) patchLexStateFunc {
	lineStart := ">"
	if t == patchDeletion {
		lineStart = "<"
	}

	return func(lexer *patchLexer) patchLexStateFunc {
		if err := lexer.expect(lineStart); err != nil {
			return lexer.errorf("start errored, %v", err)
		}
		lexer.ignore()

		if next, err := lexer.next(); err != nil {
			return lexer.errorf("%v", err)
		} else if next == '\n' {
			lexer.backup()
			return lexer.emit(t, patchLexBegin)
		} else if next != ' ' {
			return lexer.errorf("received unexpected char %c", next)
		}
		lexer.ignore()

		for {
			next, err := lexer.next()
			if err != nil {
				return lexer.errorf("%v", err)
			}

			if next == '\n' {
				lexer.backup()
				return lexer.emit(t, patchLexBegin)
			}
		}
	}
}

// lexPatch starts a lexical analysis for a diff / patch. The tokens are sent to the channel.
func lexPatch(data string) <-chan patchLexItem {
	lexer := &patchLexer{
		data:  data,
		items: make(chan patchLexItem),
	}
	go lexer.run()

	return lexer.items
}

// run the patchLexer's states.
func (lexer *patchLexer) run() {
	for stage := patchLexBegin; stage != nil; stage = stage(lexer) {
	}
	close(lexer.items)
}

// next rune from data.
func (lexer *patchLexer) next() (r rune, err error) {
	if lexer.pos >= len(lexer.data) {
		return 0, io.EOF
	}

	r, lexer.width = utf8.DecodeRuneInString(lexer.data[lexer.pos:])
	lexer.pos += lexer.width
	return
}

// expect the following string to occur in the data.
func (lexer *patchLexer) expect(expected string) error {
	for i, c := range expected {
		if next, err := lexer.next(); err != nil {
			return err
		} else if next != c {
			return fmt.Errorf("expected string received %c instead of %c at position %d", next, c, i)
		}
	}
	return nil
}

// backup the last rune.
func (lexer *patchLexer) backup() {
	lexer.pos -= lexer.width
}

// ignore the read data since the last emit / ignore.
func (lexer *patchLexer) ignore() {
	lexer.start = lexer.pos
}

// emit the read data to the channel.
func (lexer *patchLexer) emit(t patchLexType, succ patchLexStateFunc) patchLexStateFunc {
	lexer.items <- patchLexItem{t, lexer.data[lexer.start:lexer.pos]}
	lexer.start = lexer.pos
	return succ
}

// errorf emits an error back.
func (lexer *patchLexer) errorf(format string, args ...interface{}) patchLexStateFunc {
	lexer.items <- patchLexItem{patchError, fmt.Sprintf(format, args...)}
	return nil
}

// eof emits an patchEOF.
func (lexer *patchLexer) eof() patchLexStateFunc {
	lexer.items <- patchLexItem{patchEOF, ""}
	return nil
}
