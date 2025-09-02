# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

PATH	:= $(PWD)/target/debug:${PATH}

.PHONY: all required build debug release install clean verify

all:		build		# default target debug
required:       vendor          # required to build
verify:		test		# verify build
build:		lint debug	# debug build and lint

# Define or override custom env
sinclude custom.mk

debug:	required
	@install -d target/debug
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -v -mod vendor -tags debug,$(TAGS) -ldflags '-s -w -X main.etcPrefix=$(SYSCONFDIR) -X main.workingDir=$(WORKINGDIR) -X main.logPrefix=$(LOGPREFIXDIR)' -o target/debug/ ./...

release:	required
	@install -d target/release
	@GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build --buildmode=$(BUILD_MODE) -v -mod vendor -tags release,$(TAGS) -ldflags '-s -w -X main.etcPrefix=$(SYSCONFDIR) -X main.workingDir=$(WORKINGDIR) -X main.logPrefix=$(LOGPREFIXDIR)' -o target/release/ ./...

install:        release
	@install -d -m 755 $(DESTDIR)$(WORKINGDIR)
	@install -d -m 755 $(DESTDIR)$(SYSCONFDIR)
	@install -d -m 755 $(DESTDIR)$(SBINDIR)
	@install -s -m 755 target/release/spycraft $(DESTDIR)$(SBINDIR)
	@install -s -m 755 target/release/sipdump $(DESTDIR)$(BINDIR)
	@install -s -m 755 target/release/sipfind $(DESTDIR)$(BINDIR)
	@install -m 644 etc/$(PROJECT).conf $(DESTDIR)$(SYSCONFDIR)

clean:
	@$(GO) clean ./...
	@rm -rf target vendor
	@rm -f *.out *.log go.sum

setcap:	# Used for testing
	sudo setcap cap_net_raw,cap_net_admin=eip target/debug/spycraft
	sudo setcap cap_net_raw,cap_net_admin=eip target/debug/sipdump
	sudo setcap cap_net_raw,cap_net_admin=eip target/debug/sipfind

# Optional make components we add
sinclude .make/*.mk

