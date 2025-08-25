//go:build js && wasm
// +build js,wasm

package js

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall/js"
	"time"
)

var script = `return new Promise(function(resolve, reject){
	var resolved = false;
	var s = new WebSocket(addr);
	s.binaryType = 'arraybuffer';
	s.onerror = (e) => {
		console.log("websocket error");
		if (!resolved) {
			resolved = true;
			reject(e);
			return;
		}
		event(0, e);
	}
	s.onopen = (e) => {
		if (!resolved) {
			resolved = true;
			resolve(s);
			return;
		}
		event(1, e);
	}
	s.onclose = (e) => {
		if (!state.running) return;
		event(2, e);
	}
	s.onmessage = (m) => {
		event(3, new Uint8Array(m.data));
	}
})`

type WsConn struct {
	ws     JsValue
	cb     JsFunc
	state  JsValue
	events chan jsevent
	done   chan struct{}
	read   chan struct{}

	mu  sync.Mutex
	buf bytes.Buffer
	err error
}

type jsevent struct {
	event EventType
	Data  JsValue
}

func (ws *WsConn) openSocket(addr string) *JsPromise {
	ws.cb = JsCallbackOf(func(v []JsValue) {
		ev := jsevent{
			event: EventType(v[0].Int()),
			Data:  v[1],
		}
		select {
		case ws.events <- ev:
		case <-ws.done:
		}
	})

	ws.state = JsValueOf(JsObj{
		"running": true,
	})

	setup := JsNativeFuncOf("addr", "event", "state", script)
	return JsPromiseInstance(setup.Invoke(addr, ws.cb, ws.state))
}

func (ws *WsConn) Close() error {
	select {
	case <-ws.done:
	default:
		close(ws.done)
	}
	ws.state.Set("running", false)
	ws.ws.Call("close")
	ws.cb.Release()
	return ws.err
}

func (ws *WsConn) wakeRead() {
	select {
	case ws.read <- struct{}{}:
	default:
	}
}

const (
	eventError  = EventType(0)
	eventOpened = EventType(1)
	eventClosed = EventType(2)
	eventData   = EventType(3)
)

func (ws *WsConn) loop() {
	defer ws.Close()
	for {
		select {
		case <-ws.done:
			return
		case ev := <-ws.events:
			switch ev.event {
			case eventClosed:
				ws.mu.Lock()
				ws.err = errors.New("ws: connection closed")
				ws.mu.Unlock()
				ws.wakeRead()
				return
			case eventError:
				ws.mu.Lock()
				ws.err = NewError(ev.Data)
				ws.mu.Unlock()
				ws.wakeRead()
				return
			case eventData:
				arr := ev.Data
				size := arr.Get("length").Int()
				data := make([]byte, size)
				js.CopyBytesToGo(data, arr)
				ws.buf.Write(data)
				ws.wakeRead()
			}
		}
	}
}

func cloneToJS(data []byte) JsValue {
	v := JsNew("Uint8Array", len(data))
	js.CopyBytesToJS(v, data)
	return v
}

func (ws *WsConn) send(data []byte) {
	arr := cloneToJS(data)
	ws.ws.Call("send", arr)
}

func (ws *WsConn) Read(b []byte) (int, error) {
	for {
		var (
			n   int
			err error
		)
		ws.mu.Lock()
		if ws.buf.Len() != 0 {
			n, err = ws.buf.Read(b)
		} else {
			err = ws.err
		}
		ws.mu.Unlock()
		if err != nil || n != 0 {
			return n, err
		}
		select {
		case <-ws.done:
			return 0, io.EOF
		case <-ws.read:
		}
	}
}

func (ws *WsConn) Write(b []byte) (int, error) {
	ws.mu.Lock()
	err := ws.err
	ws.mu.Unlock()
	if err != nil {
		return 0, err
	}
	ws.send(b)
	return len(b), nil
}

type wsAddr struct{}

func (wsAddr) Network() string {
	return "ws"
}

func (wsAddr) String() string {
	// TODO: proper address, if possible
	return "ws://localhost"
}

func (ws *WsConn) LocalAddr() net.Addr {
	return wsAddr{}
}

func (ws *WsConn) RemoteAddr() net.Addr {
	return wsAddr{}
}

func (ws *WsConn) SetDeadline(t time.Time) error {
	return nil
}

func (ws *WsConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (ws *WsConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func DialWithContext(ctx context.Context, addr string) (_ net.Conn, gerr error) {
	ws := &WsConn{
		events: make(chan jsevent, 2),
		done:   make(chan struct{}),
		read:   make(chan struct{}),
	}
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				gerr = e
			} else {
				gerr = fmt.Errorf("%v", r)
			}
		}
	}()

	p := ws.openSocket(addr)
	out, err := p.JsAwaitContext(ctx)
	if err != nil {
		defer func() {
			ws.state.Set("running", false)
			ws.cb.Release()
		}()
		e := err.(js.Error).Value
		if !e.Get("message").IsUndefined() {
			return nil, fmt.Errorf("ws.dial: %v", err)
		}

		ev := <-ws.events
		if ev.event == eventClosed {
			code := ev.Data.Get("code").Int()
			return nil, fmt.Errorf("ws.dial: connection closed with code %d", code)
		}
		return nil, fmt.Errorf("ws.dial: connection failed, see console")
	}
	ws.ws = out[0]
	go ws.loop()
	return ws, nil
}

func Dial(addr string) (net.Conn, error) {
	return DialWithContext(context.Background(), addr)
}
