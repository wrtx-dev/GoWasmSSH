.PHONY: client clean page server serve all help
client: export GOOS=js
client: export GOARCH=wasm

deps:
	@go mod tidy
	@cd webpage&&pnpm i
	@cd cf-workers&&pnpm i

page: client
	@cd webpage&&pnpm run build
	
client:
	@go build  -ldflags "-w -s" -o webpage/public/ssh.wasm
	@cp -v $(shell go env GOROOT)/misc/wasm/wasm_exec.js webpage/public/

server: client page
	@go build -ldflags "-w -s"

all: client page server

honodev: client page
	@rm -rf cf-workers/public
	@mkdir -v cf-workers/public
	@cp -rvf webpage/dist/* cf-workers/public/
	@cd cf-workers&&pnpm run dev

deploycf: client page
	@rm -rf cf-workers/public
	@mkdir -v cf-workers/public
	@cd cf-workers&&cp -rvf ../webpage/dist/* public/&&pnpm run deploy

serve: all
	./gowasmssh

clean:
	@rm -f webpage/public/ssh.wasm
	@rm -f webpage/public/wasm_exec.js
	@rm -f gowasmssh

help:
	@echo "可用命令:"
	@echo "  make help      - 显示此帮助信息"
	@echo "  make deps      - 安装项目依赖"
	@echo "  make client    - 构建WASM客户端"
	@echo "  make page      - 构建网页前端"
	@echo "  make server    - 构建服务器"
	@echo "  make all       - 构建所有组件"
	@echo "  make honodev   - 设置开发环境"
	@echo "  make deploycf  - 部署到Cloudflare Workers"
	@echo "  make serve     - 构建并启动服务器"
	@echo "  make clean     - 清理生成的文件"

