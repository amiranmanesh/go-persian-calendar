# Security Policy

## Supported versions

The latest minor release receives security fixes.

| Version | Supported |
|---------|-----------|
| 1.4.x   | ✅        |
| < 1.4   | ❌        |

## Reporting a vulnerability

Please report vulnerabilities privately through
[GitHub Security Advisories](https://github.com/amiranmanesh/go-persian-calendar/security/advisories/new)
rather than in a public issue.

Include the affected version, a description of the impact, and a reproducer if
you have one. You can expect an initial response within seven days.

## Scope

This package has no dependencies, performs no I/O beyond loading time zone data
from the standard library, and does not execute untrusted input. The realistic
attack surface is the parser: `Parse`, `ParseInLocation`, `ParseTimeFormat` and
`ParseTimeFormatInLocation` accept arbitrary text, and a panic or unbounded
resource use in any of them is treated as a security issue. They are covered by
fuzz targets in `fuzz_test.go`.
