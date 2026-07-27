//go:build windows

package main

import "github.com/wberry-godaddy/switchboard-poc-natives/windows"

func nativeAdd(a, b int) int { return windows.Add(a, b) }
