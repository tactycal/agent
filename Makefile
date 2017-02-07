GO     := go
SOURCE := $(wildcard ./*.go)

BUILDDIR      := ./build
PKGDIR        := ./pkg
DATADIR       := ./data
CI_COMMIT     ?= dev
GIT_COMMIT    := $(shell git rev-parse --short HEAD || echo $(CI_COMMIT))
VERSION       := $(shell cat ./VERSION)
FLAGS         := "-X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"
DISTRIBUTIONS := ubuntu debian rhel centos

JFROG_URL      ?= https://bintray.com/api/v1
JFROG_API_KEY  ?= THE_KEY
JFROG_USERNAME ?= USERNAME
JFROG_SUBJECT  ?= SUBJECT
JFROG_PACKAGE  ?= agent

PATH_BIN ?= /usr/bin

ifeq ($(CI),)
DOCKER_RUN := docker run --rm -it
else
DOCKER_RUN := docker run --rm
endif

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
	/bin/rm -f Dockerfile.qa

build: $(addprefix build/,$(DISTRIBUTIONS)) ## builds the code for all distributions

test: $(addprefix test/,$(DISTRIBUTIONS)) ## runs tests for all distributions in a docker container

test/%: TEST_FLAGS ?= -v
test/%: qaContainer vet ## runs tests for a specific distributions in a docker container
	@echo "*\n* Testing for $* in a container\n*"
	$(DOCKER_RUN) -w /go/src/agent tactycal-agent-qa go test -tags $* $(TEST_FLAGS)

qaContainer: Dockerfile.qa
	docker build -q --rm -t tactycal-agent-qa -f Dockerfile.qa .

Dockerfile.qa:
	cp -f Dockerfile.debian Dockerfile.qa
	echo "ADD . /go/src/agent" >> Dockerfile.qa

testLocal: $(addprefix testLocal/,$(DISTRIBUTIONS)) ## runs tests locally for all distributions

testLocal/%: TEST_FLAGS ?= -v
testLocal/%: vetLocal ## runs unit tests locally for a specific distribution
	@echo "*\n* Testing for $* locally \n*"
	go test $(TEST_FLAGS) -tags $*

format: ## formats the code
	gofmt -w $(SOURCE)

vet: qaContainer ## examines the go code with `go vet` in a docker container
	$(DOCKER_RUN) -w /go/src/agent tactycal-agent-qa go vet $(SOURCE)

vetLocal: ## examines the go code with `go vet` locally
	go vet $(SOURCE)

up: $(addprefix up/,$(DISTRIBUTIONS)) ## start agents for all distributions
up/%: ## starts the agent for a specific distribution
	docker-compose --project-name=tactycal up agent$*

run/%:
	$(GO) build -tags $* -ldflags $(FLAGS) -o ./bin/tactycal-$*
	./bin/tactycal-$* -f my_conf.conf -s /state/$* -t 3s -d

$(PKGDIR): $(addprefix $(PKGDIR)/,$(DISTRIBUTIONS)) ## creates artifacts for all distributions

# PACKAGING
# Generates a DEB or RPM
$(PKGDIR)/rhel: FPM_DEPENDENCIES=yum
$(PKGDIR)/centos: FPM_DEPENDENCIES=yum
$(PKGDIR)/rhel: TARGET_ARTIFACT=rpm
$(PKGDIR)/centos: TARGET_ARTIFACT=rpm
$(PKGDIR)/%: FPM_DEPENDENCIES=apt
$(PKGDIR)/%: TARGET_ARTIFACT=deb
$(PKGDIR)/%: EXISTING_CNT = $(shell docker ps -a | grep tactycal-fpm-packager-$* | awk '{print $$1}')
$(PKGDIR)/%: packager/%  ## creates the artifact for a specific distribution
	mkdir -p $(PKGDIR)/$*

	# delete previous containers if found
	@$(if $(EXISTING_CNT),docker rm -f $(EXISTING_CNT) >/dev/null,)

	# build the artifact
	docker run --name tactycal-fpm-packager-$* tactycal-fpm-packager-$* -s dir -t $(TARGET_ARTIFACT) \
		--name tactycal-agent \
		--package /tactycal-agent.$(TARGET_ARTIFACT) \
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
		--after-install /scripts/after-install.sh \
		--before-install /scripts/before-install.sh \
		--after-remove /scripts/after-remove.sh \
		--before-remove /scripts/before-remove.sh \
		--after-upgrade /scripts/after-upgrade.sh \
		--before-upgrade /scripts/before-upgrade.sh \
		/build/=/

	# copy the artifact from container
	docker cp tactycal-fpm-packager-$*:/tactycal-agent.$(TARGET_ARTIFACT) $(PKGDIR)/$*/
	# remove the container
	docker rm tactycal-fpm-packager-$* >/dev/null

# PACKAGER
# Prepares an image with FPM installed and the build folder
packager/rhel: DOCKER_IMAGE=sigscicmd/fpm-rpm
packager/centos: DOCKER_IMAGE=sigscicmd/fpm-rpm
packager/rhel: DOCKERFILE = Dockerfile.packager.rpm
packager/centos: DOCKERFILE = Dockerfile.packager.rpm
packager/%: DOCKER_IMAGE=sigscicmd/fpm-deb
packager/%: EXISTING_IMG = $(shell docker images | grep tactycal-fpm-packager-$* | awk '{print $$1}')
packager/%: DOCKERFILE = Dockerfile.packager.deb
packager/%: $(BUILDDIR)/% ## Prepares a docker image with FPM installed and the build folder
	# copy over the packager dockerfile
	cp $(DOCKERFILE) $(BUILDDIR)/$*/Dockerfile

	# copy over scripts
	cp -r scripts $(BUILDDIR)/$*/scripts

	# delete previous images
	$(if $(EXISTING_IMG),docker rmi -f $(EXISTING_IMG) >/dev/null,)

	# generate a new image
	docker build -q --tag tactycal-fpm-packager-$* $(BUILDDIR)/$*

# BUILD
# Builds the go code and prepares empty folders for packages
build/%: ;## builds the code for a specific distribution
$(BUILDDIR)/%: EXISTING_CNT = $(shell docker ps -a | grep tactycal-agent-builder-$* | awk '{print $$1}')
$(BUILDDIR)/%: EXISTING_IMG = $(shell docker images | grep tactycal-agent-builder-$* | awk '{print $$1}')
$(BUILDDIR)/%:
	# prepare empty folders
	mkdir -p $@/pkg$(PATH_BIN)

	# delete previous containers if found
	@$(if $(EXISTING_IMG),docker rmi -f $(EXISTING_IMG) >/dev/null 2>/dev/null,)
	@$(if $(EXISTING_CNT),docker rm -f $(EXISTING_CNT) >/dev/null 2>/dev/null,)

	# prepare the image
	docker build -q --tag tactycal-agent-builder-$* --file Dockerfile.build .

	# build the agent
	docker run --name tactycal-agent-builder-$* --env GOOS=linux --env GOARCH=amd64 --env CGO_ENABLED=0 \
		tactycal-agent-builder-$* go build -a -installsuffix cgo -tags $* -ldflags $(FLAGS) -o /tactycal

	# copy over the binary
	docker cp tactycal-agent-builder-$*:/tactycal $@/pkg$(PATH_BIN)/tactycal

	# cleanup docker image and container
	docker rm -f tactycal-agent-builder-$* >/dev/null
	docker rmi -f tactycal-agent-builder-$* >/dev/null

# PUBLISH
# posts the packages to registry
publish: $(addprefix publish/,$(DISTRIBUTIONS)) ## publishes all artifacts to bintray
publish/%: CONTENT_PATH = $(JFROG_SUBJECT)/$*/$(JFROG_PACKAGE)/$(VERSION)
publish/%: FILE_PATH = tactycal-agent_$(VERSION)_amd64.deb
publish/%: HEADERS =
publish/%: SOURCE_FILE = pkg/$*/tactycal-agent.deb

publish/ubuntu: HEADERS = -H "X-Bintray-Debian-Distribution: tactycal" -H "X-Bintray-Debian-Component: main" -H 'X-Bintray-Debian-Architecture: amd64'
publish/debian: HEADERS = -H "X-Bintray-Debian-Distribution: tactycal" -H "X-Bintray-Debian-Component: main" -H 'X-Bintray-Debian-Architecture: amd64'

publish/rhel: SOURCE_FILE = pkg/rhel/tactycal-agent.rpm
publish/centos: SOURCE_FILE = pkg/centos/tactycal-agent.rpm
publish/rhel: FILE_PATH = tactycal-agent-$(VERSION)-x86_64.rpm
publish/centos: FILE_PATH = tactycal-agent-$(VERSION)-x86_64.rpm
publish/%: $(PKGDIR)/% ## publishes the artifact for a specific distribution
	# Check if version exists
	@versionStatus=$$(curl -s -w "%{http_code}" -o /dev/null \
		--user "$(JFROG_USERNAME):$(JFROG_API_KEY)" -H "Content-Type: application/json" \
		$(JFROG_URL)/packages/$(JFROG_SUBJECT)/$*/$(JFROG_PACKAGE)/versions/$(VERSION)); \
	if [ "$$versionStatus" = "404" ]; then \
		echo Create a new version; \
		curl -f --user "$(JFROG_USERNAME):$(JFROG_API_KEY)" -H "Content-Type: application/json"\
			-d '{"name": "$(VERSION)", "description": "Release $(VERSION)"}' \
			$(JFROG_URL)/packages/$(JFROG_SUBJECT)/$*/$(JFROG_PACKAGE)/versions; \
		echo ""; \
		echo ""; \
		\
		echo "Upload the package (or delete the version on failure)"; \
		curl -f -v --user "$(JFROG_USERNAME):$(JFROG_API_KEY)" -X PUT $(HEADERS) \
			-T $(SOURCE_FILE) "$(JFROG_URL)/content/$(CONTENT_PATH)/$(FILE_PATH)?override=1&publish=1" || \
		(curl -v --user "$(JFROG_USERNAME):$(JFROG_API_KEY)" -XDELETE \
			$(JFROG_URL)/packages/$(JFROG_SUBJECT)/$*/$(JFROG_PACKAGE)/versions/$(VERSION) && exit 0) && \
		echo "" && \
		echo "" && \
		\
		echo "Publish version (or delete the version on failure)" && \
		curl -f -v --user "$(JFROG_USERNAME):$(JFROG_API_KEY)" \
			-H "Content-Type: application/json" -d '{"publish_wait_for_secs": -1}' \
			$(JFROG_URL)/content/$(JFROG_SUBJECT)/$*/$(JFROG_PACKAGE)/$(VERSION)/publish || \
		(curl -v --user "$(JFROG_USERNAME):$(JFROG_API_KEY)" -XDELETE \
			$(JFROG_URL)/packages/$(JFROG_SUBJECT)/$*/$(JFROG_PACKAGE)/versions/$(VERSION) && exit 0) && \
		echo "" && \
		echo ""; \
	else \
		echo "WARNING: Version $(VERSION) already exists; Skipping"; \
	fi

# watch/% support
watch/%: watchmedo-exists
	watchmedo shell-command -i "./.git/*;./bin/*;./pkg/*;./build/*;./.state/*;./Dockerfile.qa" --recursive --ignore-directories --wait --command "make $*"

watchmedo-exists: ; @which watchmedo > /dev/null

vendor: ## update vendor folder
	docker run --rm -v $(PWD):/go/src/agent -w /go/src/agent trifs/govendor fetch +missing
	docker run --rm -v $(PWD):/go/src/agent -w /go/src/agent trifs/govendor remove +unused

# integration/%:
# 	@cp -f Dockerfile.$* Dockerfile.qa
# 	@echo "ADD . /go/src/agent" >> Dockerfile.qa
# 	@echo "*\n* Testing for $* in a container\n*"
# 	docker build -q --rm -t tactycal-agent-qa -f Dockerfile.qa .
# 	docker run --rm -it -w /go/src/agent tactycal-agent-qa go test -tags $* $(TEST_FLAGS)
