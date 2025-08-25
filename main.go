//go:build js && wasm
// +build js,wasm

package main

import (
	"github.com/wrtx-dev/gowasmssh/ssh"
)

func main() {
	ssh.RegisterSSHNewConnection()
	<-make(chan struct{})
}
