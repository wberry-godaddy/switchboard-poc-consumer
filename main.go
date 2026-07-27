// Command consumer is a stand-in for the switchboard-client Go module's core
// package: it never directly imports any native platform package by name.
// Instead, build-tag-gated files (native_linux.go, native_linux_musl.go,
// native_darwin.go, native_windows.go) each import exactly one
// platform-specific native module and define nativeAdd, gated so that only
// one is ever active for a given GOOS/GOARCH/musl-tag combination. Printing
// the result of an actual cgo call (not just a successful compile) proves
// the native library genuinely linked and executed.
package main

import "fmt"

func main() {
	fmt.Println(nativeAdd(2, 3))
}
