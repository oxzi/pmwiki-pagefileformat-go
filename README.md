<!--
SPDX-FileCopyrightText: 2020 Alvar Penning

SPDX-License-Identifier: GPL-3.0-or-later
-->
# PmWiki's PageFileFormat in Go

[![GoDoc](https://godoc.org/github.com/oxzi/pmwiki-pagefileformat-go?status.svg)](https://godoc.org/github.com/oxzi/pmwiki-pagefileformat-go) ![CI](https://github.com/oxzi/pmwiki-pagefileformat-go/workflows/CI/badge.svg)

A somewhat overengineered programming library to handle PmWiki's [PageFileFormat][pagefileformat] in [Go][golang].
Revision history can also be evaluated to provide file progress information.
Lexer and parser included.


## pmwiki-to-git

Based on this library, the `pmwiki-to-git` program converts the PmWiki `wiki.d` directory into a git repository.

```
# Clone this repository and build the tool
git clone https://github.com/oxzi/pmwiki-pagefileformat-go
cd pmwiki-pagefileformat-go
go build ./cmd/pmwiki-to-git

# Create a new git repository
git init pmwiki-git

# Start the conversion
./pmwiki-to-git -pmwiki ~/pmwiki/wiki.d -git pmwiki-git
```

This may log some errors.
But as long as the program does not abort, these are negligible.
PmWiki is sometimes very quirky.


## License

GNU GPLv3 or later.


[golang]: https://golang.org/
[pagefileformat]: https://www.pmwiki.org/wiki/PmWiki/PageFileFormat
