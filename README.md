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

A sixth job (`tidy-control`) runs plain `go mod tidy` with no target-specific
tags as a deliberate contrast: it fetches **all four** modules, documenting
that the win applies to `go build`/`go get`/`go mod download`, not to
`tidy`/`vendor`.

## Phase 1: real cgo linking (dynamic *and* static)

Phase 0 (module-fetch selectivity) says nothing about whether actual cgo
linking still works once the C artifacts live inside a dependency module's
read-only `GOMODCACHE` directory rather than the main module's own `lib/`.
So the native packages now ship a real tiny C library (`int add(int, int)`)
instead of inert marker files, and each build-tag-gated file
(`native_linux.go`, etc.) calls through to it and prints the result -- a
passing run proves the native code actually *executed*, not just compiled.

This directly targets the exact regression from
[#585](https://github.com/gdcorp-uxp/switchboard-client/pull/585) /
[#590](https://github.com/gdcorp-uxp/switchboard-client/pull/590): #585
pruned `libnative.a`-equivalent static archives assuming the Go client only
links dynamically; #590 restored them because static builds
(`--tags musl --ldflags '-linkmode external -extldflags "-static"'`, the
Alpine/Lambda deploy shape the Domains org depends on) resolve `-lnative`
against the `.a`, not the `.so` -- and without it in the same `-L` directory,
linking fails with `ld: cannot find -lnative`. The `linux-musl` native module
ships *both* `libnative.so` and `libnative.a` in one `lib/` dir, and the
*same* single `#cgo LDFLAGS` line must work for both cases.

The `linux-musl-static` CI job reproduces this exact scenario, but through a
dependency module instead of the main module's own `lib/` directory:

```
$ go build -tags musl --ldflags '-linkmode external -extldflags "-static"' -o /tmp/out .
$ /tmp/out
5
$ file /tmp/out
/tmp/out: ELF 64-bit LSB executable, ARM aarch64, ..., statically linked, ...
$ ldd /tmp/out
/lib/ld-musl-aarch64.so.1: /tmp/out: Not a valid dynamic program
```

Static linking against a `.a` shipped inside a *dependency* module works
identically to the main module's own `lib/` directory -- the multi-module
restructuring does not reopen the #590 regression.

All five other jobs (`linux-glibc-dynamic`, `linux-musl-dynamic`,
`darwin-dynamic`, `windows-dynamic`, `tidy-control`) additionally assert the
program's output is `5` and that the binary is linked the expected way
(`ldd`/`otool -L` shows the dynamic library; Windows confirmed locally via
`objdump -p`, which shows `native.dll` as an import with the `add` symbol
bound).
