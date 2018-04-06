ROOT_DIR     := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
IMAGE_NAME   := crawlera-headless-proxy
APP_NAME     := $(IMAGE_NAME)
GOMETALINTER := gometalinter.v2
CC_DIR       := $(ROOT_DIR)/crosscompile

VENDOR_DIRS  := $(shell find "$(ROOT_DIR)/vendor" -type d 2>/dev/null || echo "$(ROOT_DIR)/vendor")
VENDOR_FILES := $(shell find "$(ROOT_DIR)/vendor" -type f 2>/dev/null || echo "$(ROOT_DIR)/vendor")
CC_BINARIES  := $(shell bash -c "echo $(APP_NAME)-{linux,windows,darwin}-{386,amd64} $(APP_NAME)-linux-{arm,arm64}")

# -----------------------------------------------------------------------------

.PHONY: all docker clean lint test install_cli install_dep install_lint \
	crosscompile crosscompile-dir

# -----------------------------------------------------------------------------

define crosscompile
  GOOS=$(2) GOARCH=$(3) go build -o "$(1)/$(APP_NAME)-$(2)-$(3)" -ldflags="-s -w"
endef

# -----------------------------------------------------------------------------

all: $(APP_NAME)
crosscompile: $(CC_BINARIES)

$(APP_NAME): version.go proxy/certs.go $(VENDOR_DIRS) $(VENDOR_FILES)
	@go build -o "$(APP_NAME)" -ldflags="-s -w"

$(APP_NAME)-%: GOOS=$(shell echo "$@" | sed 's?$(APP_NAME)-??' | cut -f1 -d-)
$(APP_NAME)-%: GOARCH=$(shell echo "$@" | sed 's?$(APP_NAME)-??' | cut -f2 -d-)
$(APP_NAME)-%: version.go proxy/certs.go $(VENDOR_DIRS) $(VENDOR_FILES) crosscompile-dir
	@$(call crosscompile,$(CC_DIR),$(GOOS),$(GOARCH))

crosscompile-dir:
	@rm -rf "$(CC_DIR)" && mkdir -p "$(CC_DIR)"

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
		git submodule foreach --recursive sh -c 'git clean -xfd && git reset --hard'

docker:
	@docker build --pull -t "$(IMAGE_NAME)" "$(ROOT_DIR)"

install_cli: install_dep install_lint

install_dep:
	@go get github.com/golang/dep/cmd/dep

install_lint:
	@go get gopkg.in/alecthomas/gometalinter.v2 && \
		$(GOMETALINTER) --install
