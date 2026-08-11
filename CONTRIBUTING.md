# Contributing

Thanks for taking the time to contribute. This is a small, focused library, so
the bar is mostly about correctness and keeping the API close to the standard
`time` package.

## Getting started

```bash
git clone https://github.com/amiranmanesh/go-persian-calendar.git
cd go-persian-calendar
make test
```

You need Go 1.21 or newer. The linter is pinned in `tools/go.mod` and installed
automatically by `make lint`, so there is nothing else to set up.

## Before opening a pull request

```bash
make fmt      # gofmt + gofumpt
make lint     # golangci-lint, must be clean
make test     # tests with -race and coverage
make fuzz     # a short fuzzing pass, worth running for parser changes
```

## What a good change looks like

- **Tests come with the change.** Calendar bugs are subtle; a table entry that
  fails before the fix and passes after is the most convincing review comment
  you can write.
- **Behavior changes are documented.** If output changes, say so in
  `CHANGELOG.md` under `Unreleased`, with a before and after.
- **The API stays close to `time`.** If the standard library has a method for
  the same idea, match its name, signature and semantics.
- **Exported identifiers have doc comments** that start with the identifier
  name, as `go doc` expects.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
feat: add ParseInLocation
fix: correct fractional seconds in TimeFormat
docs: document the two layout languages
test: add fuzz target for the parser
refactor: split formatting into its own file
build: bump the minimum Go version
ci: run the test matrix on Windows
```

Breaking changes carry a `!` (`feat!:`) and a `BREAKING CHANGE:` footer.

## Releases

Releases are cut from `main` by pushing a tag:

```bash
git tag -a v1.5.0 -m "v1.5.0"
git push origin v1.5.0
```

The release workflow verifies the tag, runs the full test suite and publishes a
GitHub release from the matching `CHANGELOG.md` section.

## Reporting bugs

Open an issue with the Persian and Gregorian dates involved, the layout string
if formatting or parsing is affected, and what you expected instead. A failing
test case is ideal.

## Code of conduct

Participation is governed by the [Code of Conduct](CODE_OF_CONDUCT.md).
