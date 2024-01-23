GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

VENDOR ?= "Iceber"
CONTRACT_ID = "ollama:llm"
REMOTE_PROVIDER_NAME = "Ollama LLM Remote Provider"
STANDALONE_PROVIDER_NAME = "Ollama LLM Standalone Provider"

WASI_SNAPSHOT_PREVIEW1_ADAPTOR ?= "./hack/wasi_snapshot_preview1.proxy.wasm"

all: build-adaptor-component openai-server remote-provider-par

.PHONY: build-adaptor-component
build-adaptor-component:
	cd ./actor-adaptor && cargo build --release --target wasm32-wasi
	wasm-tools component new ./actor-adaptor/target/wasm32-wasi/release/ollama_actor_adaptor.wasm -o ./ollama_wasmcloud_adaptor.wasm --adapt $(WASI_SNAPSHOT_PREVIEW1_ADAPTOR)

.PHONY: build-openai-server-component
build-openai-server-component:
	cd ./examples/openai-server && cargo build --release --target wasm32-wasi
	mkdir -p ./build
	wasm-tools component new ./examples/openai-server/target/wasm32-wasi/release/ollama_openai_server.wasm \
		-o ./build/openai_server.wasm --adapt $(WASI_SNAPSHOT_PREVIEW1_ADAPTOR)

.PHONY: openai-server
openai-server: build-adaptor-component build-openai-server-component
	wasm-tools compose -o ollama_openai_server.wasm -d ./ollama_wasmcloud_adaptor.wasm ./build/openai_server.wasm
	wash claims sign -q -l -c $(CONTRACT_ID) -n "ollama-openai-server" -r 1 -v 0.1.0 ./ollama_openai_server.wasm

.PHONY: build-remote-provider
build-remote-provider:
	go build -C ./provider -o ../build/remote-provider ./cmd/remote-provider

.PHONY: build-standalone-provider
build-standalone-provider:
	git submodule update --init --recursive
	go generate -C ./provider/ollama ./...
	go build -C ./provider -o ../build/standalone-provider ./cmd/standalone-provider

.PHONY: build-remote-provider-to-par
build-remote-provider-to-par:
	go build -C ./provider -o ../build/remote-provider-$(GOOS)-$(GOARCH) ./cmd/remote-provider
ifeq (,$(wildcard remote-provider.par))
		wash par create --name $(REMOTE_PROVIDER_NAME) --arch $(RUSTARCH)-$(RUSTOS) --capid $(CONTRACT_ID) --binary ./build/remote-provider-$(GOOS)-$(GOARCH) --vendor $(VENDOR) --destination=remote-provider.par
else
		wash par insert --arch $(RUSTARCH)-$(RUSTOS) --binary ./build/remote-provider-$(GOOS)-$(GOARCH) remote-provider.par
endif

.PHONY: remote-provider-par
remote-provider-par:
	rm remote-provider.par 2>/dev/null | true
	GOOS="linux" GOARCH="arm64" RUSTOS="linux" RUSTARCH="aarch64" $(MAKE) build-remote-provider-to-par

.PHONY: remote-provider-multi-platform-par
remote-provider-multi-platform-par:
	rm remote-provider.par 2>/dev/null | true
	GOOS="darwin" GOARCH="arm64" RUSTOS="macos" RUSTARCH="aarch64" $(MAKE) build-remote-provider-to-par
	GOOS="darwin" GOARCH="amd64" RUSTOS="macos" RUSTARCH="x86_64" $(MAKE) build-remote-provider-to-par
	GOOS="linux" GOARCH="arm64" RUSTOS="linux" RUSTARCH="aarch64" $(MAKE) build-remote-provider-to-par
	GOOS="linux" GOARCH="amd64" RUSTOS="linux" RUSTARCH="x86_64" $(MAKE) build-remote-provider-to-par
	GOOS="windows" GOARCH="amd64" RUSTOS="windows" RUSTARCH="x86_64" $(MAKE) build-remote-provider-to-par

.PHONY: standalone-provider-par
standalone-provider-par: build-standalone-provider
	wash par create --name $(STANDALONE_PROVIDER_NAME) --arch aarch64-macos --capid $(CONTRACT_ID) --binary ./provider/standalone-provider --vendor $(VENDOR) --destination=standalone-provider.par

.PHONY: clean
clean:
	rm -rf *.wasm *.par ./build
