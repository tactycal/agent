GO     := go
SOURCE := $(wildcard ./*.go)

BUILDDIR      := ./build
PKGDIR        := ./pkg
DATADIR       := ./data
CI_COMMIT     ?= dev
GIT_COMMIT    := $(shell git rev-parse --short HEAD || echo $(CI_COMMIT))
VERSION       ?= $(shell cat ./VERSION)
FLAGS         := "-X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
DISTRIBUTIONS := ubuntu debian rhel centos opensuse sles amzn
PACKAGE_MANAGER := dpkg rpm

JFROG_URL      ?= https://bintray.com/api/v1
JFROG_API_KEY  ?= THE_KEY
JFROG_USERNAME ?= USERNAME
JFROG_SUBJECT  ?= SUBJECT
JFROG_PACKAGE  ?= agent

PATH_BIN ?= /usr/bin

.DEFAULT_GOAL := help
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

test: vet 
	go test -v
	$(MAKE) -C osDiscovery test
	$(MAKE) -C packageLookup test

format: ## formats the code
	gofmt -w $(SOURCE)

vet: ## examines the go code with `go vet`
	go vet

up: $(addprefix up/,$(DISTRIBUTIONS)) ## start agents for all distributions
up/%: ## starts the agent for a specific distribution
	docker-compose --project-name=tactycal up agent$*

run/%:
	$(GO) build  -ldflags $(FLAGS) -o ./bin/tactycal-$*
	./bin/tactycal-$* -f my_conf.conf -s /state/$* -t 3s -d

$(PKGDIR): $(addprefix $(PKGDIR)/,$(PACKAGE_MANAGER)) ## creates artifacts for all distributions

# PACKAGING
$(PKGDIR)/rpm: TARGET_ARTIFACT=rpm
$(PKGDIR)/rpm: FPM_DEPENDENCIES=rpm
$(PKGDIR)/rpm: TARGET_FILE=tactycal-agent-$(VERSION)-x86_64.rpm
$(PKGDIR)/dpkg: TARGET_ARTIFACT=deb
$(PKGDIR)/dpkg: FPM_DEPENDENCIES=apt
$(PKGDIR)/dpkg: TARGET_FILE=tactycal-agent_$(VERSION)_amd64.deb
$(PKGDIR)/%: build ## creates the artifact for a specific distribution
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
		./build/=/

# BUILD
build: ;## builds the code
	mkdir -p $@$(PATH_BIN)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo \
		 -ldflags $(FLAGS) -o $@$(PATH_BIN)/tactycal

vendor: ## update vendor folder
	docker run --rm -v $(PWD):/go/src/agent -w /go/src/agent trifs/govendor fetch +missing
	docker run --rm -v $(PWD):/go/src/agent -w /go/src/agent trifs/govendor remove +unused

# watch/% support
watch/%: watchmedo-exists
	watchmedo shell-command -i "./.git/*;./bin/*;./pkg/*;./build/*;./.state/*" --recursive --ignore-directories --wait --command "make $*"

watchmedo-exists: ; @which watchmedo > /dev/null
