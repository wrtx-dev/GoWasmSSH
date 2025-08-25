//go:build js && wasm
// +build js,wasm

package ssh

import (
	"io"
	gjs "syscall/js"

	"github.com/pkg/sftp"
	"github.com/wrtx-dev/gowasmssh/js"
)

// SFTPClient implements a simple SFTP client
type SFTPClient struct {
	client *sftp.Client
}

func NewSFTPClient(client *sftp.Client) interface{} {

	jsClient := js.Global().Get("Object").New()
	sftClient := &SFTPClient{client: client}
	jsClient.Set("list", js.JsFuncOf(sftClient.list))
	jsClient.Set("close", js.JsFuncOf(sftClient.close))
	jsClient.Set("cwd", js.JsFuncOf(sftClient.sftpCWD))
	jsClient.Set("download", js.JsFuncOf(sftClient.jsDownloadFile))
	jsClient.Set("upload", js.JsFuncOf(sftClient.jsUploadFile))
	return jsClient
}

func (c *SFTPClient) list(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 1 {
		return js.Global().Get("error").New("need path")
	}
	go func() {
		infos, err := c.client.ReadDir(args[0].String())
		if err != nil {
			return
		}
		if len(args) > 1 && args[1].Type().String() == "function" {
			func() {
				files := js.Global().Get("Array").New(len(infos))
				for _, info := range infos {
					finfo := js.Global().Get("Object").New()
					finfo.Set("name", info.Name())
					finfo.Set("isDir", info.IsDir())
					finfo.Set("size", info.Size())
					finfo.Set("modTime", info.ModTime().Local().String())
					finfo.Set("mode", info.Mode().String())
					finfo.Set("path", args[0].String()+"/"+info.Name())
					files.SetIndex(files.Length(), finfo)
				}
				args[1].Invoke(files)
			}()
		}
	}()
	return nil
}

func (c *SFTPClient) close(_ js.JsValue, _ []js.JsValue) interface{} {
	go func() {
		if c.client != nil {
			c.client.Close()
			c.client = nil
		}
	}()
	return nil
}

func (c *SFTPClient) sftpCWD(_ js.JsValue, args []js.JsValue) interface{} {
	go func() {
		if c.client != nil {
			wd, err := c.client.Getwd()
			if err != nil {
				return
			}
			if len(args) > 0 && args[0].Type().String() == "function" {
				args[0].Invoke(wd)
			}
		}
	}()
	return nil
}

func (c *SFTPClient) jsDownloadFile(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 2 {
		return js.Global().Get("error").New("need remote path and callback")
	}
	remotePath := args[0].String()
	callback := args[1]
	go func() {

		remoteFile, err := c.client.Open(remotePath)
		if err != nil {

			return
		}
		defer remoteFile.Close()
		data, err := io.ReadAll(remoteFile)
		if err != nil {

			return
		}
		uint8Array := js.JsNew("Uint8Array", len(data))
		gjs.CopyBytesToJS(uint8Array, data)
		if !callback.IsUndefined() {
			callback.Invoke(uint8Array)
		}
	}()
	return nil
}

// upload(path, data, callback)
func (c *SFTPClient) jsUploadFile(_ js.JsValue, args []js.JsValue) interface{} {
	if len(args) < 3 {
		return js.Global().Get("error").New("need remote path, data and callback")
	}
	remotePath := args[0].String()
	data := args[1]
	callback := args[2]
	go func() {
		remoteFile, err := c.client.Create(remotePath)
		if err != nil {
			return
		}
		defer remoteFile.Close()
		uint8Array := gjs.Global().Get("Uint8Array").New(data)
		bytes := make([]byte, uint8Array.Length())
		gjs.CopyBytesToGo(bytes, uint8Array)
		_, err = remoteFile.Write(bytes)
		if err != nil {

			return
		}
		if !callback.IsUndefined() {
			callback.Invoke()
		}
	}()
	return nil
}
