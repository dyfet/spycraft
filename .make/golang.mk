# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2025 David Sugar <tychosoft@gmail.com>

.PHONY: lint vet fix test cover stage release upgrade

ifndef	BUILD_MODE
BUILD_MODE := default
endif

ifndef	GO
GO := go
endif

STATIC_CHECK	:= $(shell which staticcheck 2>/dev/null || true )
ifeq ($(STATIC_CHECK),)
STATIC_CHECK	:= true
endif

GOVULN_CHECK	:= $(shell which govulncheck 2>/dev/null || true)
ifeq ($(GOVULN_CHECK),)
GOVULN_CHECK	:= true
endif

GOVER=$(shell grep ^go <go.mod)
TARGET := $(CURDIR)/target
export GOCACHE := $(TARGET)/cache
export PATH := $(TARGET)/debug:${PATH}

docs:	required
	@rm -rf target/docs
	@install -d target/docs
	@doc2go -out target/docs ./...

lint:	required
	@$(GO) fmt ./...
	@$(GO) mod tidy
	@$(STATIC_CHECK) ./...

vet:	required
	@$(GO) vet ./...
	@$(GOVULN_CHECK) ./...

fix:	required
	@$(GO) fix ./...

test:
	@$(GO) test ./...

stage:
	@rm -rf target/stage
	@install -d target/stage
	@$(MAKE) DESTDIR=$(CURDIR)/target/stage install

cover:
	@$(GO) test -coverprofile=coverage.out ./...

go.sum:	go.mod
	@$(GO) mod tidy

upgrade: required
	@.make/upgrade.sh

# if no vendor directory (clean) or old in git checkouts
vendor:	go.sum
	@if test -d .git ; then \
		rm -rf vendor ;\
		$(GO) mod vendor ;\
	elif test ! -d vendor ; then \
		$(GO) mod vendor ;\
	else \
		touch vendor ;\
	fi
