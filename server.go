//go:build !wasm && !js
// +build !wasm,!js

package main

import (
	"context"
	"embed"
	"flag"
	"os"
	"os/signal"
	"syscall"

	server "github.com/wrtx-dev/gowasmssh/package/server"
)

//go:embed webpage/dist/*
var staticFS embed.FS

var addr string
var port int

func init() {
	flag.StringVar(&addr, "listen", "0.0.0.0", "listen address")
	flag.IntVar(&port, "port", 9090, "listen port")
}
func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	server := server.NewWsToTcpServer(ctx, addr, port)
	go func() {
		defer cancel()
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		server.Shutdown()
	}()

	server.Serve(staticFS)
}
