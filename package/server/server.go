//go:build !wasm && !js
// +build !wasm,!js

package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type WsToTcpServer struct {
	// Addr is the address to listen on.
	Addr   string
	Port   int
	ctx    context.Context
	server *http.Server
}

func NewWsToTcpServer(ctx context.Context, addr string, port int) *WsToTcpServer {
	return &WsToTcpServer{
		Addr:   addr,
		Port:   port,
		ctx:    ctx,
		server: nil,
	}
}

func ping(ctx context.Context, conn *websocket.Conn) {
	w, err := conn.NewFrameWriter(websocket.PingFrame)
	if err != nil {
		return
	}
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := w.Write(nil)
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func parseOrigin(origin string) (string, error) {
	u, err := url.Parse(origin)
	if err != nil {
		return "", err
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		var addrErr *net.AddrError
		if errors.As(err, &addrErr) {
			if strings.Contains(addrErr.Err, "missing port in address") {
				return u.Host, nil
			}
		}
		return "", err
	}
	return host, nil
}

func genWsHandler(ctx context.Context, server string, port int) websocket.Handler {
	return websocket.Handler(func(conn *websocket.Conn) {
		tcp, err := net.DialTimeout("tcp", net.JoinHostPort(server, fmt.Sprintf("%d", port)), 30*time.Second)
		if err != nil {
			conn.Close()
			return
		}

		// 创建子上下文用于管理这个连接的goroutine
		connCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		defer tcp.Close()
		defer conn.Close()

		conn.PayloadType = websocket.BinaryFrame

		var wg sync.WaitGroup
		done := make(chan struct{})
		countChan := make(chan int, 10)
		stopChan := make(chan struct{}, 2)

		// Ping goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			ping(connCtx, conn)
		}()

		// Copy goroutines
		copier := func(from IReaderWithTimeout, to io.Writer) {
			defer wg.Done()
			err := copyData(connCtx, from, to, countChan, stopChan)
			if err != nil {
				fmt.Printf("Copy error: %v\n", err)
			}
		}

		wg.Add(2)
		go copier(conn, tcp)
		go copier(tcp, conn)

		total := 0

		// Monitor total transfer size
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-connCtx.Done():
					return
				case <-done:
					return
				case count := <-countChan:
					total += count

				}
			}
		}()

		// 等待所有goroutine完成或上下文取消
		wg.Wait()
		close(countChan)
		close(stopChan)
	})
}

func (s *WsToTcpServer) wsUpgradeHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if len(origin) == 0 {
		http.Error(w, "not allow", http.StatusForbidden)
		return
	}
	host, err := parseOrigin(origin)
	if err != nil {
		http.Error(w, "not allow", http.StatusForbidden)
		return
	}
	if host != "localhost" && host != "wrtx.dev" && host != "www.wrtx.dev" {
		http.Error(w, "not allow", http.StatusForbidden)
		return
	}
	connection := r.Header.Get("Connection")
	if strings.ToLower(connection) != "upgrade" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	remoteAddr := r.PathValue("server")
	portAddr := r.PathValue("port")
	if len(remoteAddr) == 0 || len(portAddr) == 0 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if isLocalIP(remoteAddr) {
		http.Error(w, "bad request", http.StatusForbidden)
		return
	}
	port, err := strconv.Atoi(portAddr)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	websocket.Handler(genWsHandler(s.ctx, remoteAddr, port)).ServeHTTP(w, r)
}

func (s *WsToTcpServer) Serve(staticFS embed.FS) {
	var staticFiles = fs.FS(staticFS)
	sub, err := fs.Sub(staticFiles, "webpage/dist")
	if err != nil {
		fmt.Println("can't load embed files")
		return
	}
	hfs := http.FileServer(http.FS(sub))
	mux := http.NewServeMux()
	mux.Handle("/", hfs)
	mux.Handle("/ws/{server}/{port}", http.HandlerFunc(s.wsUpgradeHandler))
	server := http.Server{
		Addr:    net.JoinHostPort(s.Addr, fmt.Sprintf("%d", s.Port)),
		Handler: mux,
	}
	s.server = &server
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("serve err:", err)
	}
}

func (s *WsToTcpServer) Shutdown() {
	if s.server != nil {
		s.server.Shutdown(s.ctx)
	}
}

func isLocalIP(ip string) bool {
	localIPs := []string{"127.0.0.1", "localhost", "::1", "0.0.0.0"}

	for _, localIP := range localIPs {
		if ip == localIP {
			return true
		}
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP != nil {
		return parsedIP.IsLoopback() || parsedIP.IsPrivate() || parsedIP.IsUnspecified()
	}

	return false
}
