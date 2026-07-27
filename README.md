# switchboard-poc-consumer

Proof-of-concept for a proposed restructuring of
[switchboard-client](https://github.com/gdcorp-uxp/switchboard-client)'s Go
client: splitting per-platform native FFI binaries into separate Go modules
(see [switchboard-poc-natives](https://github.com/wberry-godaddy/switchboard-poc-natives)),
selected via build-tag-gated blank imports, instead of shipping every
platform's binaries in one module.

## What this proves

The real client's problem: today, `go/switchboard/lib/` ships all 7
platforms' `.so`/`.a`/`.dylib`/`.dll` artifacts in **one** Go module. Build
tags only affect what's *compiled*, not what's *downloaded* -- a Go module's
zip is fetched all-or-nothing regardless of `//go:build` constraints on
individual files within it (this is explicitly documented: `go mod vendor`
and `go mod tidy` "act as if all build tags are enabled"). So every consumer
downloads every platform's binaries no matter what they're building for.

This repo restructures that pattern: `main.go` has no direct import of any
platform package. Instead, 4 build-tag-gated files
(`native_linux.go` `//go:build linux && !musl`,
`native_linux_musl.go` `//go:build linux && musl`,
`native_darwin.go` `//go:build darwin`,
`native_windows.go` `//go:build windows`) each blank-import exactly one
platform module from `switchboard-poc-natives`. Because normal build commands
(unlike `tidy`/`vendor`) resolve the package graph using the *actual* target
build constraints, only one of those four files is ever active for a given
build -- so only its module needs to be fetched.

## Evidence

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) builds for each of the
4 targets (`linux` default, `linux` `-tags musl`, `darwin`, `windows`) on a
fresh GitHub Actions runner (clean module cache by construction) and asserts,
by inspecting `$(go env GOMODCACHE)`, that **only** the matching platform's
module (15/20/25/30MB respectively, deliberately distinct sizes) was
downloaded -- and explicitly fails the job if any other platform's module
shows up.

A fifth job (`tidy-control`) runs plain `go mod tidy` with no target-specific
tags as a deliberate contrast: it fetches **all four** modules, documenting
that the win applies to `go build`/`go get`/`go mod download`, not to
`tidy`/`vendor`.

Reproduced locally (clean `GOPATH` per run) before pushing CI, e.g.:

```
$ GOOS=linux GOARCH=amd64 go build .
go: downloading .../switchboard-poc-natives/linux v0.1.0
$ du -sh $(go env GOMODCACHE)/github.com/wberry-godaddy/switchboard-poc-natives/*
 15M .../linux@v0.1.0
# linux-musl (20M), darwin (25M), windows (30M) -- absent
```
