package pmwiki

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// pageFileParser parses a PageFile based on the pageFileLexer's channel.
type pageFileParser struct {
	pf  PageFile
	err error

	urlencoded bool

	lexItems <-chan pageFileLexItem
}

// pageFileParseStateFunc parses a pageFileLexer and returns its successive pageFileParseStateFunc.
type pageFileParseStateFunc func(*pageFileParser) pageFileParseStateFunc

// next item from the lexer.
func (parser *pageFileParser) next() pageFileLexItem {
	return <-parser.lexItems
}

// nextType returns the next matching item's value. A positive max value restricts the amount skipped items.
func (parser *pageFileParser) nextType(lexType pageFileLexType, max int) (v string, err error) {
	for i := 0; max <= 0 || i < max; i++ {
		if item := parser.next(); item.t == lexType {
			return item.v, nil
		} else if item.t == EOF {
			return "", io.EOF
		} else if item.t == Error {
			return "", fmt.Errorf(item.v)
		}
	}
	return "", fmt.Errorf("no item with type %v found in %d messages", lexType, max)
}

// acceptItem expects the given items and consumes them.
func (parser *pageFileParser) acceptItem(lexItems ...pageFileLexItem) error {
	for _, lexItem := range lexItems {
		if recv := parser.next(); recv.t != lexItem.t {
			return fmt.Errorf("type %v was expected, received %v", lexItem.t, recv.t)
		} else if recv.v != lexItem.v {
			return fmt.Errorf("value %v was expected, received %v", lexItem.v, recv.v)
		}
	}
	return nil
}

// errorf stores an error and aborts.
func (parser *pageFileParser) errorf(format string, args ...interface{}) pageFileParseStateFunc {
	parser.pf = PageFile{}
	parser.err = fmt.Errorf(format, args...)

	return nil
}

// pageFileParseVersion parses the initial version item.
func pageFileParseVersion(parser *pageFileParser) pageFileParseStateFunc {
	if err := parser.acceptItem(pageFileLexItem{Key, "version"}); err != nil {
		return parser.errorf("initial version expected, %w", err)
	}

	if version, err := parser.nextType(Value, 1); err != nil {
		return parser.errorf("initial version expected, %w", err)
	} else {
		parser.urlencoded = strings.Contains(version, "urlencoded=1")
		parser.pf.Version = version
	}

	return pageFileParseFields
}

// pageFileParseFields parses the other items.
func pageFileParseFields(parser *pageFileParser) pageFileParseStateFunc {
	for {
		key, err := parser.nextType(Key, 0)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return parser.errorf("parsing key errored, %w", err)
		}

		var opts []string
		var value string

	itemTokenLoop:
		for item := parser.next(); ; item = parser.next() {
			switch item.t {
			case KeyOpt:
				opts = append(opts, item.v)

			case Value:
				value = item.v
				break itemTokenLoop

			default:
				return parser.errorf("received unexpected item type %v", item)
			}
		}

		if len(opts) == 0 {
			err = pageFileParseMainItem(parser, key, value)
		} else {
			// TODO
		}
		if err != nil {
			return parser.errorf("parsing item errored, %w", err)
		}
	}
}

// pageFileParseMainItem parses the main items without KeyOpts.
func pageFileParseMainItem(parser *pageFileParser, key, value string) error {
	switch key {
	case "name":
		if parser.pf.Name != "" {
			return fmt.Errorf("name field was already set")
		}
		parser.pf.Name = value

	case "time":
		if parser.pf.Time != (time.Time{}) {
			return fmt.Errorf("time field was already set")
		}
		if unix, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("time parsing errored, %w", err)
		} else {
			parser.pf.Time = time.Unix(unix, 0).UTC()
		}

	case "text":
		if parser.pf.Text != "" {
			return fmt.Errorf("text field was already set")
		}
		parser.pf.Text = value

	case "author":
		if parser.pf.Author != "" {
			return fmt.Errorf("author field was already set")
		}
		parser.pf.Author = value

	case "host":
		if len(parser.pf.Host) != 0 {
			return fmt.Errorf("host field was already set")
		}
		if host := net.ParseIP(value); host == nil {
			return fmt.Errorf("parsing host %s errored", value)
		} else {
			parser.pf.Host = host
		}

	case "rev":
		if parser.pf.Rev != 0 {
			return fmt.Errorf("rev field was already set")
		}
		if rev, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("rev parsing errored, %w", err)
		} else {
			parser.pf.Rev = int(rev)
		}

	default:
		// unknown / unsupported item
	}

	return nil
}

// ParsePageFile parses PmWiki's PageFileFormat into a PageFile.
func ParsePageFile(r io.Reader) (PageFile, error) {
	parser := &pageFileParser{
		pf:       PageFile{Revs: make(map[time.Time]PageFileRevision)},
		lexItems: lexPageFile(r),
	}

	for state := pageFileParseVersion; state != nil; state = state(parser) {
	}

	return parser.pf, parser.err
}
