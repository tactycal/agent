GO     := go
SOURCE := $(wildcard ./*.go)

BUILDDIR      := ./build
PKGDIR        := ./pkg
DATADIR       := ./data
CI_COMMIT     ?= dev
GIT_COMMIT    := $(shell git rev-parse --short HEAD || echo $(CI_COMMIT))
VERSION       ?= $(shell cat ./VERSION)
FLAGS         := "-X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
DISTRIBUTIONS := ubuntu debian rhel centos

JFROG_URL      ?= https://bintray.com/api/v1
JFROG_API_KEY  ?= THE_KEY
JFROG_USERNAME ?= USERNAME
JFROG_SUBJECT  ?= SUBJECT
JFROG_PACKAGE  ?= agent

PATH_BIN ?= /usr/bin

.PHONY: help vendor

all: format vet build test

help: ## displays this message
	@grep -E '^[a-zA-Z_/%\-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@printf "\n\033[36m%-30s\033[0m %s\n" "Supported distributions" "$(DISTRIBUTIONS)"

clean: ## cleans up the repository
	/bin/rm -rf ./bin
	/bin/rm -rf $(BUILDDIR)
	/bin/rm -rf $(PKGDIR)
	/bin/rm -rf $(DATADIR)

build: $(addprefix build/,$(DISTRIBUTIONS)) ## builds the code for all distributions

test: $(addprefix test/,$(DISTRIBUTIONS)) ## runs tests for all distributions

test/%: TEST_FLAGS ?= -v
test/%: vet ## runs unit tests for a specific distribution
	@echo "*\n* Testing for $* locally \n*"
	go test $(TEST_FLAGS) -tags $*

format: ## formats the code
	gofmt -w $(SOURCE)

vet: ## examines the go code with `go vet`
	go vet $(SOURCE)

up: $(addprefix up/,$(DISTRIBUTIONS)) ## start agents for all distributions
up/%: ## starts the agent for a specific distribution
	docker-compose --project-name=tactycal up agent$*

run/%:
	$(GO) build -tags $* -ldflags $(FLAGS) -o ./bin/tactycal-$*
	./bin/tactycal-$* -f my_conf.conf -s /state/$* -t 3s -d

$(PKGDIR): $(addprefix $(PKGDIR)/,$(DISTRIBUTIONS)) ## creates artifacts for all distributions

# PACKAGING
$(PKGDIR)/rhel: FPM_DEPENDENCIES=yum
$(PKGDIR)/centos: FPM_DEPENDENCIES=yum
$(PKGDIR)/rhel: TARGET_ARTIFACT=rpm
$(PKGDIR)/rhel: TARGET_FILE=tactycal-agent-$(VERSION)-x86_64.rpm
$(PKGDIR)/centos: TARGET_ARTIFACT=rpm
$(PKGDIR)/%: FPM_DEPENDENCIES=apt
$(PKGDIR)/%: TARGET_ARTIFACT=deb
$(PKGDIR)/%: TARGET_FILE=tactycal-agent_$(VERSION)_amd64.deb
$(PKGDIR)/%: build/% ## creates the artifact for a specific distribution
	mkdir -p $(PKGDIR)/$*
	fpm -s dir -t $(TARGET_ARTIFACT) \
		--name tactycal-agent \
		--package ./pkg/$*/$(TARGET_FILE) \
		--force \
		--category admin \
		--epoch $(shell /bin/date +%s) \
		--iteration $(GIT_COMMIT) \
		--deb-compression bzip2 \
		--url https://tactycal.com \
		--description "Tactycal Agent" \
		--maintainer "Tactycal <support@tactycal.com>" \
		--license "Apache-2.0" \
		--vendor "tactycal" \
		--version $(VERSION) \
		--architecture amd64 \
		--depends $(FPM_DEPENDENCIES) \
		--after-install ./scripts/after-install.sh \
		--before-install ./scripts/before-install.sh \
		--after-remove ./scripts/after-remove.sh \
		--before-remove ./scripts/before-remove.sh \
		--after-upgrade ./scripts/after-upgrade.sh \
		--before-upgrade ./scripts/before-upgrade.sh \
		./build/$*/=/

# BUILD
build/%: ;## builds the code for a specific distribution
$(BUILDDIR)/%:
	mkdir -p $@$(PATH_BIN)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo \
		-tags $* -ldflags $(FLAGS) -o $@$(PATH_BIN)/tactycal

vendor: ## update vendor folder
	docker run --rm -v $(PWD):/go/src/agent -w /go/src/agent trifs/govendor fetch +missing
	docker run --rm -v $(PWD):/go/src/agent -w /go/src/agent trifs/govendor remove +unused

# watch/% support
watch/%: watchmedo-exists
	watchmedo shell-command -i "./.git/*;./bin/*;./pkg/*;./build/*;./.state/*" --recursive --ignore-directories --wait --command "make $*"

watchmedo-exists: ; @which watchmedo > /dev/null
