.PHONY: all clean fmt tidy install get build
.PHONY: FORCE

GO ?= go
GOFMT ?= gofmt
GOFMT_FLAGS = -w -l -s

GOPATH ?= $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin

export GOPATH GOBIN

TMPDIR ?= $(CURDIR)/.tmp
OUTDIR ?= $(TMPDIR)

REVIVE_VERSION ?= v1.5.1
REVIVE_CONF ?= revive.toml
REVIVE_RUN_ARGS ?= -config $(REVIVE_CONF) -formatter friendly
REVIVE_URL ?= github.com/mgechev/revive@$(REVIVE_VERSION)
REVIVE ?= $(GO) run $(REVIVE_URL)

V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell if [ "$$(tput colors 2> /dev/null || echo 0)" -ge 8 ]; then printf "\033[34;1m▶\033[0m"; else printf "▶"; fi)

MODULE   = $(shell $(GO) list)
DATE    ?= $(shell date +%F)
VERSION ?= $(shell git describe --tags --always --dirty=-dev --match=v* 2> /dev/null || \
			cat .version 2> /dev/null || echo v0)

GO_BUILD_CMD_LDFLAGS = \
	-X $(MODULE)/version.Version=$(VERSION) \
	-X $(MODULE)/version.BuildDate=$(DATE)
GO_BUILD_CMD_FLAGS = -o "$(OUTDIR)/$(BINARY_NAME)" -ldflags "$(GO_BUILD_CMD_LDFLAGS)"

GO_BUILD = $(GO) build -v
GO_BUILD_CMD = $(GO_BUILD) $(GO_BUILD_CMD_FLAGS)

all: get tidy build

clean: ; $(info $(M) cleaning…)
	rm -rf $(TMPDIR)

fmt: ; $(info $(M) reformatting sources…)
	$Q find . -name '*.go' | xargs -r $(GOFMT) $(GOFMT_FLAGS)

install:
	$Q $(GO) install -v -ldflags "$(GO_BUILD_CMD_LDFLAGS)" ./cmd/...

get: ; $(info $(M) getting dependencies…)
	$Q $(GO) mod tidy
	$Q $(GO) mod download

build: ; $(info $(M) building…)
	$Q $(GO_BUILD_CMD)

tidy: | fmt ; $(info $(M) tidy: root)
	$(Q) $(GO) mod tidy
	$(Q) $(GO) vet ./...
	$(Q) $(REVIVE) $(REVIVE_RUN_ARGS) ./...
