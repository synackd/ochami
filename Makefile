
# Set path to commands
GO      ?= $(shell command -v go)
GIT     ?= $(shell command -v git)
INSTALL ?= $(shell command -v install)
SCDOC   ?= $(shell command -v scdoc)
SHELL   ?= /bin/sh

INSTALL_PROGRAM ?= $(INSTALL) -Dm755
INSTALL_DATA    ?= $(INSTALL) -Dm644

IMPORT := github.com/OpenCHAMI/ochami/

prefix ?= /usr/local
exec_prefix ?= $(prefix)
bindir ?= $(exec_prefix)/bin
mandir ?= $(exec_prefix)/man

# Check that commands are present
ifeq ($(GO),)
$(error '$(GO)' command not found.)
endif
ifeq ($(GIT),)
$(error '$(GIT)' command not found.)
endif
ifeq ($(INSTALL),)
$(error '$(INSTALL)' command not found.)
endif
ifeq ($(SCDOC),)
$(error '$(SCDOC)' command not found.)
endif
ifeq ($(SHELL),)
$(error '$(SHELL)' command not found.)
endif

NAME    ?= ochami
VERSION ?= $(shell git describe --tags --always --dirty --broken --abbrev=0)
BUILD   ?= $(shell git rev-parse --short HEAD)
LDFLAGS := -s -X=$(IMPORT)internal/version.Version=$(VERSION) -X=$(IMPORT)internal/version.Commit=$(BUILD) -X=$(IMPORT)internal/version.Date=$(shell date -Iseconds)

INTERNAL := $(wildcard internal/*)
MANSRC   := $(wildcard man/*.sc)
MANBIN   := $(subst .sc,,$(MANSRC))
MAN1BIN  := $(filter %.1,$(MANBIN))
MAN5BIN  := $(filter %.5,$(MANBIN))

.PHONY: all
all: binaries

.PHONY: binaries
binaries: $(NAME)

.PHONY: clean
clean:
	$(GO) clean -i -x

.PHONY: clean-man
clean-man:
	rm -f $(MANBIN)

.PHONY: clean-completions
clean-completions:
	rm -rf completions/

completions: $(NAME)
	./scripts/completions.sh

.PHONY: distclean
distclean: clean clean-completions clean-man

.PHONY: install
install: install-prog install-completions install-man

.PHONY: install-prog
install-prog: $(NAME)
	$(INSTALL_PROGRAM) $(NAME) $(DESTDIR)$(bindir)/$(NAME)

.PHONY: install-completions
install-completions: completions
	$(INSTALL_DATA) ./completions/ochami.bash $(DESTDIR)/usr/share/bash-completion/completions/ochami
	$(INSTALL_DATA) ./completions/ochami.fish $(DESTDIR)/usr/share/fish/vendor_completions.d/ochami.fish
	$(INSTALL_DATA) ./completions/ochami.zsh $(DESTDIR)/usr/share/zsh/site-functions/_ochami

.PHONY: install-man
install-man: $(MANBIN)
	mkdir -p $(DESTDIR)$(mandir)/man1
	mkdir -p $(DESTDIR)$(mandir)/man5
	$(INSTALL_DATA) $(MAN1BIN) $(DESTDIR)$(mandir)/man1/
	$(INSTALL_DATA) $(MAN5BIN) $(DESTDIR)$(mandir)/man5/

.PHONY: man
man: $(MANBIN)

man/%: man/%.sc
	$(SCDOC) < $< > $@

.PHONY: uninstall
uninstall: uninstall-prog uninstall-completions uninstall-man

.PHONY: uninstall-prog
uninstall-prog:
	rm -f $(DESTDIR)$(bindir)/$(NAME)

.PHONY: uninstall-completions
uninstall-completions:
	rm -f $(DESTDIR)/usr/share/bash-completion/completions/ochami
	rm -f $(DESTDIR)/usr/share/fish/vendor_completions.d/ochami.fish
	rm -f $(DESTDIR)/usr/share/zsh/site-functions/_ochami

.PHONY: uninstall-man
uninstall-man:
	rm -f $(foreach man1page,$(subst man/,,$(MAN1BIN)),$(DESTDIR)$(mandir)/man1/$(man1page))
	rm -f $(foreach man5page,$(subst man/,,$(MAN5BIN)),$(DESTDIR)$(mandir)/man5/$(man5page))

$(NAME): *.go cmd/*.go $(foreach file,$(INTERNAL),$(wildcard $(file)/*.go))
	$(GO) build -v -ldflags="$(LDFLAGS)"
