// Command consumer is a stand-in for the switchboard-client Go module's core
// package: it never directly imports any native platform package by name.
// Instead, build-tag-gated files (native_linux.go, native_linux_musl.go,
// native_darwin.go, native_windows.go) each blank-import exactly one
// platform-specific native module, gated so that only one is ever active for
// a given GOOS/GOARCH/musl-tag combination.
package main

func main() {}
