//go:build js && wasm
// +build js,wasm

package js

import (
	"context"
)

type JsPromise struct {
	v    JsValue
	done <-chan struct{}
	res  []JsValue
	err  error
}

type JsValueWrapper struct {
	v JsValue
}

func JsNewPromise(fn func() ([]interface{}, error)) JsValue {
	var initFunc JsFunc
	initFunc = AsyncCallbackOf(func(args []JsValue) {
		initFunc.Release()
		resolve, reject := args[0], args[1]
		res, err := fn()
		if err != nil {
			if w, ok := err.(Wrapper); ok {
				reject.Invoke(w)
			} else {
				reject.Invoke(err.Error())
			}
		} else {
			resolve.Invoke(res...)
		}
	})
	return JsNew("Promise", initFunc)
}

func (p *JsPromise) JsValue() JsValue {
	return p.v
}

func (p *JsPromise) JsAwait() ([]JsValue, error) {
	<-p.done
	return p.res, p.err
}

func (p *JsPromise) JsAwaitContext(ctx context.Context) ([]JsValue, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.done:
		return p.res, p.err
	}
}

func (v JsValueWrapper) Promised() *JsPromise {
	done := make(chan struct{})
	p := &JsPromise{
		v:    v.v,
		done: done,
	}

	var then, catch JsFunc
	then = JsCallbackOf(func(v []JsValue) {
		then.Release()
		catch.Release()
		p.res = v
		close(done)
	})
	catch = JsCallbackOf(func(v []JsValue) {
		then.Release()
		catch.Release()
		var e JsValue
		if len(v) != 0 {
			e = v[0]
		}
		if e.IsUndefined() {
			e = object
		}
		p.err = NewError(e)
		close(done)
	})
	v.v.Call("then", then).Call("catch", catch)
	return p
}

func JsPromiseInstance(v JsValue) *JsPromise {
	wrapper := JsValueWrapper{
		v: v,
	}
	return wrapper.Promised()
}
func JsValueAwait(v JsValue) ([]JsValue, error) {
	wrapper := JsValueWrapper{
		v: v,
	}
	return wrapper.Promised().JsAwait()
}
