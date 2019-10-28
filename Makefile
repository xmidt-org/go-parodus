DEFAULT: build

GO           ?= go
GOFMT        ?= $(GO)fmt
APP          := parodus
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
BINARY    	 := $(FIRST_GOPATH)/bin/$(APP)

PROGVER = $(shell git describe --tags `git rev-list --tags --max-count=1` | tail -1 | sed 's/v\(.*\)/\1/')
RPM_VERSION=$(shell echo $(PROGVER) | sed 's/\(.*\)-\(.*\)/\1/')
RPM_RELEASE=$(shell echo $(PROGVER) | sed -n 's/.*-\(.*\)/\1/p'  | grep . && (echo "$(echo $(PROGVER) | sed 's/.*-\(.*\)/\1/')") || echo "1")
BUILDTIME = $(shell date -u '+%Y-%m-%d %H:%M:%S')
GITCOMMIT = $(shell git rev-parse --short HEAD)

.PHONY: go-mod-vendor
go-mod-vendor:
	GO111MODULE=on $(GO) mod vendor

.PHONY: build
build: go-mod-vendor
	$(GO) build -o $(APP) -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"

.PHONY: version
version:
	@echo $(PROGVER)

# If the first argument is "update-version"...
ifeq (update-version,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "update-version"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

.PHONY: update-version
update-version:
	@echo "Update Version $(PROGVER) to $(RUN_ARGS)"
	git tag v$(RUN_ARGS)

.PHONY: install
install: go-mod-vendor
	$(GO) install -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"

.PHONY: release-artifacts
release-artifacts: go-mod-vendor
	mkdir -p ./.ignore/binaries

	# create binaries
	GOOS=darwin  GOARCH=amd64 $(GO) build -o ./.ignore/binaries/$(APP)-$(PROGVER).darwin-amd64  -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"
	GOOS=darwin  GOARCH=386   $(GO) build -o ./.ignore/binaries/$(APP)-$(PROGVER).darwin-386    -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"
	GOOS=linux   GOARCH=amd64 $(GO) build -o ./.ignore/binaries/$(APP)-$(PROGVER).linux-amd64   -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"
	GOOS=linux   GOARCH=arm   $(GO) build -o ./.ignore/binaries/$(APP)-$(PROGVER).linux-arm     -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"
	GOOS=linux   GOARCH=386   $(GO) build -o ./.ignore/binaries/$(APP)-$(PROGVER).linux-386     -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"
	GOOS=windows GOARCH=amd64 $(GO) build -o ./.ignore/binaries/$(APP)-$(PROGVER).windows-amd64 -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"
	GOOS=windows GOARCH=386   $(GO) build -o ./.ignore/binaries/$(APP)-$(PROGVER).windows-386   -ldflags "-X 'main.BuildTime=$(BUILDTIME)' -X main.GitCommit=$(GITCOMMIT) -X main.Version=$(PROGVER)"

	# Create tars
	tar -czf ./.ignore/$(APP)-$(PROGVER).darwin-amd64.tar.gz  ./.ignore/binaries/$(APP)-$(PROGVER).darwin-amd64  NOTICE LICENSE CHANGELOG.md
	tar -czf ./.ignore/$(APP)-$(PROGVER).darwin-386.tar.gz    ./.ignore/binaries/$(APP)-$(PROGVER).darwin-386    NOTICE LICENSE CHANGELOG.md
	tar -czf ./.ignore/$(APP)-$(PROGVER).linux-amd64.tar.gz   ./.ignore/binaries/$(APP)-$(PROGVER).linux-amd64   NOTICE LICENSE CHANGELOG.md
	tar -czf ./.ignore/$(APP)-$(PROGVER).linux-arm.tar.gz     ./.ignore/binaries/$(APP)-$(PROGVER).linux-arm     NOTICE LICENSE CHANGELOG.md
	tar -czf ./.ignore/$(APP)-$(PROGVER).linux-386.tar.gz     ./.ignore/binaries/$(APP)-$(PROGVER).linux-386     NOTICE LICENSE CHANGELOG.md
	tar -czf ./.ignore/$(APP)-$(PROGVER).windows-amd64.tar.gz ./.ignore/binaries/$(APP)-$(PROGVER).windows-amd64 NOTICE LICENSE CHANGELOG.md
	tar -czf ./.ignore/$(APP)-$(PROGVER).windows-386.tar.gz   ./.ignore/binaries/$(APP)-$(PROGVER).windows-386   NOTICE LICENSE CHANGELOG.md

	# create checksums
	touch ./.ignore/sha256sums.txt
	cd .ignore/; shasum -a 256 $(APP)-$(PROGVER).darwin-amd64.tar.gz  >> sha256sums.txt
	cd .ignore/; shasum -a 256 $(APP)-$(PROGVER).darwin-386.tar.gz    >> sha256sums.txt
	cd .ignore/; shasum -a 256 $(APP)-$(PROGVER).linux-amd64.tar.gz   >> sha256sums.txt
	cd .ignore/; shasum -a 256 $(APP)-$(PROGVER).linux-arm.tar.gz     >> sha256sums.txt
	cd .ignore/; shasum -a 256 $(APP)-$(PROGVER).linux-386.tar.gz     >> sha256sums.txt
	cd .ignore/; shasum -a 256 $(APP)-$(PROGVER).windows-amd64.tar.gz >> sha256sums.txt
	cd .ignore/; shasum -a 256 $(APP)-$(PROGVER).windows-386.tar.gz   >> sha256sums.txt

.PHONY: style
style:
	! $(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

.PHONY: test
test: go-mod-vendor
	GO111MODULE=on $(GO) test -v -race  -coverprofile=cover.out ./...

.PHONY: test-cover
test-cover: test
	$(GO) tool cover -html=cover.out

.PHONY: codecov
codecov: test
	curl -s https://codecov.io/bash | bash

.PHONY: clean
clean:
	rm -rf ./$(APP) ./.ignore ./coverage.txt ./vendor
