.PHONY: client clean serve all
client: export GOOS=js
client: export GOARCH=wasm

client:
	@go build  -ldflags "-w -s" -o webpage/public/ssh.wasm
	@cp -v $(shell go env GOROOT)/misc/wasm/wasm_exec.js webpage/public/

clean:
	@rm -f webpage/public/ssh.wasm
