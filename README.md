# switchboard-poc-consumer

Proof-of-concept for a proposed restructuring of
[switchboard-client](https://github.com/gdcorp-uxp/switchboard-client)'s Go
client: splitting per-platform native FFI binaries into separate Go modules
(see [switchboard-poc-natives](https://github.com/wberry-godaddy/switchboard-poc-natives)),
selected via build-tag-gated imports, instead of shipping every platform's
binaries in one module.

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
`native_windows.go` `//go:build windows`) each import exactly one platform
module from `switchboard-poc-natives` and define `nativeAdd`. Because normal
build commands (unlike `tidy`/`vendor`) resolve the package graph using the
*actual* target build constraints, only one of those four files is ever
active for a given build -- so only its module needs to be fetched.
`main.go` calls `nativeAdd(2, 3)` and prints the result, proving the native
library genuinely linked and executed, not just compiled.

## Evidence

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs six jobs on
fresh GitHub Actions runners (clean module cache by construction):

| Job | Runner | Proves |
|---|---|---|
| `linux-glibc-dynamic` | ubuntu-latest | only `linux` module fetched; real dynamic cgo call returns `5`; `ldd` confirms dynamic link |
| `linux-musl-dynamic` | ubuntu-latest (Alpine) | same, musl variant |
| `linux-musl-static` | ubuntu-latest (Alpine) | `--ldflags '-linkmode external -extldflags "-static"'` succeeds; `file` confirms a fully static binary -- reproducing, through a dependency module instead of the main module's own `lib/`, the exact scenario a previous attempt at trimming static archives broke and had to be reverted for |
| `darwin-dynamic` | macos-latest | only `darwin` module fetched; `otool -L` confirms dynamic link |
| `windows-dynamic` | windows-latest | only `windows` module fetched; DLL resolved via `PATH` at runtime |
| `tidy-control` | ubuntu-latest | plain `go mod tidy` (no target-specific tags) fetches **all four** modules -- documents that the selective-download win applies to `go build`/`go get`/`go mod download`, not to `tidy`/`vendor` |

Each job that builds also asserts, by inspecting `$(go env GOMODCACHE)`, that
only the matching platform's native module was downloaded -- and explicitly
fails if any other platform's module shows up.

Example (`linux-musl-static`):

```
$ go build -tags musl --ldflags '-linkmode external -extldflags "-static"' -o /tmp/out .
$ /tmp/out
5
$ file /tmp/out
/tmp/out: ELF 64-bit LSB executable, ..., statically linked, ...
$ ldd /tmp/out
/lib/ld-musl-x86_64.so.1: /tmp/out: Not a valid dynamic program
```
