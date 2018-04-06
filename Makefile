ROOT_DIR     := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
IMAGE_NAME   := crawlera-headless-proxy
APP_NAME     := $(IMAGE_NAME)
GOMETALINTER := gometalinter.v2

VENDOR_DIRS  := $(shell find ./vendor -type d 2>/dev/null || echo "vendor")
VENDOR_FILES := $(shell find ./vendor -type f 2>/dev/null || echo "vendor")

# -----------------------------------------------------------------------------

.PHONY: all docker clean lint test install_cli install_dep install_lint

# -----------------------------------------------------------------------------

define crosscompile
  GOOS=$(2) GOARCH=$(3) go build -o "$(1)/$(APP_NAME)-$(2)-$(3)" -ldflags="-s -w"
endef

# -----------------------------------------------------------------------------


all: $(APP_NAME)

$(APP_NAME): version.go proxy/certs.go $(VENDOR_DIRS) $(VENDOR_FILES)
	@go build -o "$(APP_NAME)" -ldflags="-s -w"

version.go:
	@go generate main.go

proxy/certs.go:
	@go generate proxy/proxy.go

vendor: Gopkg.lock Gopkg.toml install_cli
	@dep ensure

test: install_cli
	@go test -v ./...

lint: vendor install_cli
	@$(GOMETALINTER) --deadline=2m ./...

clean:
	@git clean -xfd && \
		git reset --hard && \
		git submodule foreach --recursive sh -c 'git clean -xfd && git reset --hard' && \
		rm -rf ./vendor

docker:
	@docker build --pull -t "$(IMAGE_NAME)" "$(ROOT_DIR)"

install_cli: install_dep install_lint

install_dep:
	@go get github.com/golang/dep/cmd/dep

install_lint:
	@go get gopkg.in/alecthomas/gometalinter.v2 && \
		$(GOMETALINTER) --install
