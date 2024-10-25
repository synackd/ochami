# Set path to commands
GO  ?= $(shell command -v go)
GIT ?= $(shell command -v git)

IMPORT := github.com/synackd/ochami/

# Check that commands are present
ifeq ($(GO),)
$(error 'go' command not found.)
endif
ifeq ($(GIT),)
$(error 'git' command not found.)
endif

NAME    ?= ochami
VERSION ?= $(shell git describe --tags --always --dirty --broken --abbrev=0)
BUILD   ?= $(shell git rev-parse --short HEAD)
LDFLAGS := -s -X=$(IMPORT)internal/version.Version=$(VERSION) -X=$(IMPORT)internal/version.Commit=$(BUILD) -X=$(IMPORT)internal/version.Date=$(shell date -Iseconds)

INTERNAL := $(wildcard internal/*)

.PHONY: all
all: binaries

.PHONY: binaries
binaries: $(NAME)

.PHONY: clean
clean:
	$(GO) clean -i -x

$(NAME): *.go cmd/*.go $(foreach file,$(INTERNAL),$(wildcard $(file)/*.go))
	$(GO) build -v -ldflags="$(LDFLAGS)"
