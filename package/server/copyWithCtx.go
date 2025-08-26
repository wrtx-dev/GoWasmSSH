//go:build !wasm && !js
// +build !wasm,!js

package server

import (
	"context"
	"io"
	"net"
	"time"
)

type IReaderWithTimeout interface {
	io.Reader
	SetReadDeadline(t time.Time) error
}

func copyData(ctx context.Context, src IReaderWithTimeout, dst io.Writer, counter chan<- int, stop <-chan struct{}) error {
	buf := make([]byte, 4096)
	var remaining []byte

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stop:
			return nil
		default:
		}

		// 先写入剩余的数据
		if len(remaining) > 0 {
			nw, err := dst.Write(remaining)
			if err != nil && err != io.ErrShortWrite {
				return err
			}
			select {
			case counter <- nw:
			case <-ctx.Done():
				return ctx.Err()
			}
			if nw < len(remaining) {
				remaining = remaining[nw:]
				continue
			}
			remaining = nil
		}

		src.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := src.Read(buf)
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				continue
			}
			return err
		}
		if n == 0 {
			return nil
		}

		// 尝试写入所有读取的数据
		nw, err := dst.Write(buf[:n])
		if err != nil && err != io.ErrShortWrite {
			return err
		}

		select {
		case counter <- nw:
		case <-ctx.Done():
			return ctx.Err()
		}

		// 如果有未写入的数据，保存到remaining
		if nw < n {
			remaining = make([]byte, n-nw)
			copy(remaining, buf[nw:n])
		}
	}
}
