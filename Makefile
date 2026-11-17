# Makefile for zenui
#
# Release and version-sync logic lives in .repoman.json, driven by
# `repoman relcore` and `repoman syncver`; this Makefile only wraps them.
# `make help` lists every target; `make` alone builds, vets and tests.

SHELL := /usr/bin/env bash
.SHELLFLAGS := -eu -o pipefail -c

VERSION := $(shell cat VERSION 2>/dev/null || echo 0.0.0)
# purego keeps the raylib adapter cgo-free.
TAGS    := purego
export CGO_ENABLED := 0

.PHONY: all build test vet fmt fmt-check verify release release-resume help

## all: build, vet and test (default target)
all: build vet test

## build: compile every package
build:
	go build -tags $(TAGS) ./...

## test: run the test suite
test:
	go test -tags $(TAGS) -count=1 ./...

## vet: go vet over every package
vet:
	go vet -tags $(TAGS) ./...

## fmt: gofmt every Go file in place
fmt:
	gofmt -w .

## fmt-check: fail if any Go file is not gofmt-clean
fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "not gofmt-clean:"; echo "$$out"; exit 1; fi

## verify: register, version, module and forbidden-string gates
verify:
	repoman register check
	repoman syncver show
	repoman gomod check
	repoman badcode check
	repoman guards stale

## release: run the manifest-driven release for the version in VERSION
release:
	repoman relcore $(VERSION)

## release-resume: resume an interrupted release
release-resume:
	repoman relcore $(VERSION) --resume

## help: list the targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //'
