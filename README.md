# urlcheck

URLCheck is a project which provides a secure, compatible solution for external e-mail security rewriting, including warning text and URL rewriting.

# Project scope

## Milter

- Support common e-mail formats
- Global language encoding support
- Prefix warnings regarding external e-mails

## Proxy

- Show interstitial warning page during normal use
- Prevent URL access for known-malicious URLs
- Match against mirrored databases of malicious URLs
  - Currently supported: Google Safebrowsing, Cisco Phishtank, Abuse.ch URLHaus

# Out of scope

- Other e-mail filtering (combine multiple milters for this)
- Alerting (currently client-only, accepting contributions)
- Detection of malicious content (currently relying on curated databases)

## Installation

Currently, no pre-compiled release binaries are available. To install `urlmilter` or `urlproxy`, install the [Go toolchain](https://golang.org/) and run

```
go get github.com/fabian-z/urlcheck/cmd/urlproxy
go get github.com/fabian-z/urlcheck/cmd/urlmilter
```

The resulting binaries will be located in `$GOPATH/bin` or in `$HOME/go/bin` and can be automatically started with a service manager of your choice. Example service files are provided in `res/`.

# TODO

- Make KT branding optional, provide default blank branding

# Contributions

Contributions welcome - feel free to fork, experiment and open an issue and / or pull request.

# License

Apache License 2.0, see LICENSE
