# gorep

A small, fast grep-like tool written in Go. `gorep` searches text for a regular expression and prints matching lines with colored highlights. It supports reading from stdin, searching an inline string, a file, or recursively searching files in a directory.

**Key Features**
- Pattern search using Go regular expressions (`regexp`).
- Colored output: matches highlighted in green, file/line labels in red/white (via `fortio.org/terminal/ansipixels/tcolor`).
- Supports reading from `stdin`, a single file, or a directory (recurses into subdirectories).
- `-no-trim` option to preserve leading whitespace in printed lines.

Requirements
- Go (1.18+ recommended). Verify with `go version`.

Build

```powershell
go mod tidy
go build -o gorep .
```

Run (examples)

- Search piped input:

```powershell
echo "hello world" | gorep "world"
```

- Search an inline string (pattern then text):

```powershell
gorep "foo" "this is foo and that is bar"
```

- Search a file:

```powershell
gorep "TODO" -f README.md
```

- Search a directory (recursively):

```powershell
gorep "TODO" -f ./src
```

Flags
- `-f <path>` : read input from a file or directory. If a directory is provided, `gorep` will walk the directory and search files it can read.
- `-no-trim` : disable trimming leading indentation in each printed line. By default `gorep` trims leading tabs/spaces around matches.

Behavior details
- The first non-flag argument is treated as the regular expression pattern.
- If additional non-flag arguments are provided after the pattern they are joined into a single input string to search (convenient for one-off searches from the CLI).
- If no `-f` is provided and no inline text is given, `gorep` reads from `stdin` until EOF.
- Matches in a line are highlighted in green; printed lines are numbered and prefixed with color-coded labels. When searching directories, each file's results are prefixed by the filename.

Limitations & Notes
- The directory traversal implementation changes the working directory while recursing; if you rely on a stable working directory in other parts of a larger script, be cautious.
- Files that cannot be read are skipped silently when walking directories.
- Errors (invalid regexp, unreadable file, etc.) will log a message and exit with a non-zero status.

Testing

```powershell
go test ./...
```

Contributing
- Open issues or pull requests. Small, focused changes are easiest to review.

License
- No license file is included. Add a license (for example `MIT` or `Apache-2.0`) if you want to allow reuse.

Source
- Main implementation: `main.go`
