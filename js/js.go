//go:build js && wasm
// +build js,wasm

package js

import (
	"strings"
	"sync"
	"syscall/js"
)

var (
	global    = js.Global()
	NULL      = js.Null()
	Undefined = js.Undefined()
	object    = global.Get("Object")
	array     = global.Get("Array")
	mu        sync.RWMutex
	jsClasses = make(map[string]JsValue)
)

type Wrapper interface {
	JSVaule() JsValue
}
type (
	JsValue   = js.Value
	JsFunc    = js.Func
	TJsFunc   = func(this JsValue, args []JsValue) any
	EventType = int
	JsObj     = map[string]interface{}
	Error     = js.Error
)

func Global() JsValue {
	return js.Global()
}

func Alert(msg string) {
	jalert := JsGet("alert")
	jalert.Invoke(msg)
}

func JsValueOf(o interface{}) JsValue {
	return js.ValueOf(o)
}
func JsGet(name string, path ...string) JsValue {
	v := global.Get(name)
	for _, p := range path {
		v = v.Get(p)
	}
	return v
}

func JsFuncOf(fn TJsFunc) JsFunc {
	return js.FuncOf(fn)
}

func jsCallbackOf(fn func(v []JsValue), asyncFlag bool) JsFunc {
	return JsFuncOf(func(this JsValue, args []JsValue) interface{} {
		vars := make([]JsValue, 0, len(args))
		vars = append(vars, args...)
		if asyncFlag {
			go fn(vars)
		} else {
			fn(vars)
		}
		return nil
	})
}

func JsCallbackOf(fn func(v []JsValue)) JsFunc {
	return jsCallbackOf(fn, false)
}

func AsyncCallbackOf(fn func(v []JsValue)) JsFunc {
	return jsCallbackOf(fn, true)
}

func JsNativeFuncOf(args ...string) JsValue {
	argList := make([]interface{}, len(args))
	for i, v := range args {
		argList[i] = v
	}
	return JsNew("Function", argList...)
}

func JsNew(class string, args ...interface{}) JsValue {
	v := JsClass(class)
	return v.New(args...)
}

func JsClass(class string, path ...string) JsValue {
	switch class {
	case "Object":
		return object
	case "Array":
		return array
	}
	key := class
	if len(path) != 0 {
		key += "." + strings.Join(path, ".")
	}
	mu.RLock()
	v, ok := jsClasses[key]
	mu.RUnlock()
	if !ok {
		v = JsGet(class, path...)
		mu.Lock()
		jsClasses[key] = v
		mu.Unlock()
	}
	return v
}

func NewError(e JsValue) error {
	return Error{Value: e}
}
