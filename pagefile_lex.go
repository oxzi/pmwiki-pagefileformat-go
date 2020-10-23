package pmwiki

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

// PageFileLexType are the different tokens which might be extracted.
type PageFileLexType int

const (
	_ PageFileLexType = iota

	// EOF is an internal io.EOF.
	EOF
	// Error in case of an invalid state transition.
	Error

	// Key name.
	Key
	// KeyOpt are additional options for the previous Key.
	KeyOpt
	// Value for the previous Key.
	Value
)

func (t PageFileLexType) String() string {
	switch t {
	case EOF:
		return "eof"
	case Error:
		return "error"
	case Key:
		return "key"
	case KeyOpt:
		return "key_opt"
	case Value:
		return "value"
	default:
		panic("unknown type")
	}
}

// PageFileLexItem is a tuple of a PageFileLexType with its value.
type PageFileLexItem struct {
	T PageFileLexType
	V string
}

func (item PageFileLexItem) String() string {
	return fmt.Sprintf("%v\t%s", item.T, item.V)
}

// pageFileLexer is a lexer to tokenize PmWiki's PageFileFormat.
//
// Its logic and code structure is heavily inspired by Rob Pike's "Lexical Scanning in Go" talk,
// <https://talks.golang.org/2011/lex.slide>. However, it operates on an io.Reader instead of a string.
type pageFileLexer struct {
	reader *bufio.Reader
	items  chan PageFileLexItem
}

// pageFileLexStateFunc is lexing until a pageFileLexer and returns its successive pageFileLexStateFunc.
type pageFileLexStateFunc func(*pageFileLexer) pageFileLexStateFunc

// pageFileLexBegin inspects a line's start and selects between an EOF or a Key.
func pageFileLexBegin(lexer *pageFileLexer) pageFileLexStateFunc {
	if _, err := lexer.next(); err == nil {
		lexer.backup()
		return pageFileLexKey
	} else if err == io.EOF {
		return lexer.eof()
	} else {
		return lexer.errorf("pageFileLexBegin: %v", err)
	}
}

// pageFileLexKey extracts a Key.
func pageFileLexKey(lexer *pageFileLexer) pageFileLexStateFunc {
	var field string
	for {
		r, err := lexer.next()
		if err != nil {
			return lexer.errorf("pageFileLexKey: %v", err)
		}

		switch r {
		case ':':
			return lexer.emit(Key, field, pageFileLexKeyOpt)

		case '=':
			return lexer.emit(Key, field, pageFileLexVal)

		default:
			if unicode.IsSpace(r) {
				return lexer.errorf("pageFileLexKey: unexpected white space")
			}
			field += string(r)
		}
	}
}

// pageFileLexKeyOpt extracts a Key's KeyOpt.
func pageFileLexKeyOpt(lexer *pageFileLexer) pageFileLexStateFunc {
	var field string
	for {
		r, err := lexer.next()
		if err != nil {
			return lexer.errorf("pageFileLexKeyOpt: %v", err)
		}

		switch r {
		case '=':
			return lexer.emit(KeyOpt, field, pageFileLexVal)

		default:
			if unicode.IsSpace(r) {
				return lexer.errorf("pageFileLexKeyOpt: unexpected white space")
			}
			field += string(r)
		}
	}
}

// pageFileLexVal extracts a Key's Value.
func pageFileLexVal(lexer *pageFileLexer) pageFileLexStateFunc {
	var field string
	for {
		r, err := lexer.next()
		if err != nil {
			return lexer.errorf("pageFileLexVal: %v", err)
		}

		if r == '\n' {
			return lexer.emit(Value, field, pageFileLexBegin)
		}
		field += string(r)
	}
}

// LexPageFile starts a lexical analysis for PmWiki's PageFileFormat. The tokens are sent to the channel.
func LexPageFile(reader io.Reader) chan PageFileLexItem {
	lexer := &pageFileLexer{
		reader: bufio.NewReader(reader),
		items:  make(chan PageFileLexItem),
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
func (lexer *pageFileLexer) emit(t PageFileLexType, v string, succ pageFileLexStateFunc) pageFileLexStateFunc {
	lexer.items <- PageFileLexItem{t, v}
	return succ
}

// errorf emits an error back.
func (lexer *pageFileLexer) errorf(format string, args ...interface{}) pageFileLexStateFunc {
	return lexer.emit(Error, fmt.Sprintf(format, args...), nil)
}

// eof emits an EOF back.
func (lexer *pageFileLexer) eof() pageFileLexStateFunc {
	return lexer.emit(EOF, "", nil)
}
