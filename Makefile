.PHONY: client clean serve all
client: export GOOS=js
client: export GOARCH=wasm

client:
	@go build  -ldflags "-w -s" -o ssh-page/public/ssh.wasm
	@cp -v $(shell go env GOROOT)/misc/wasm/wasm_exec.js ssh-page/public/

clean:
	@rm -f ssh-page/public/ssh.wasm