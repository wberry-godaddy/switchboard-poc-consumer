//go:build linux && !musl

package main

import "github.com/wberry-godaddy/switchboard-poc-natives/linux"

func nativeAdd(a, b int) int { return linux.Add(a, b) }
