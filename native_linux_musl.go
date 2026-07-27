//go:build linux && musl

package main

import linuxmusl "github.com/wberry-godaddy/switchboard-poc-natives/linux-musl"

func nativeAdd(a, b int) int { return linuxmusl.Add(a, b) }
