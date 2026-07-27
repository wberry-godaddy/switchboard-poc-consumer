//go:build darwin

package main

import "github.com/wberry-godaddy/switchboard-poc-natives/darwin"

func nativeAdd(a, b int) int { return darwin.Add(a, b) }
