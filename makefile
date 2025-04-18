# For local use in this makefile. This does not export to sub-processes.
-include .env.default.properties
-include $(or $(CONF),.)/.env.properties

MAKEFLAGS        := --silent --always-make
MAKE_CONC        := $(MAKE) -j 128 CONC=true clear=$(or $(clear),false)

VERB_SHORT       ?= $(if $(filter false,$(verb)),,-v)
CLEAR_SHORT      ?= $(if $(filter false,$(clear)),,-c)

TAR              ?= ./
GO_CMD           ?= $(TAR)/fbads
GO_CMD_LINUX     ?= $(GO_CMD)_linux
GO_CMD_SRC       ?= ./cmd/fbads

GO_PKG           ?= $(or $(pkg),./...)
GO_FLAGS         ?= -tags=$(tags) -mod=mod -buildvcs=false
GO_RUN_ARGS      ?= $(GO_FLAGS) $(GO_CMD_SRC) $(args)
GO_VET_FLAGS     ?= -composites=false
GO_WATCH_FLAGS   ?= $(and $(pkg),-w=$(pkg))

GO_TEST_FAIL     ?= $(if $(filter false,$(fail)),,-failfast)
GO_TEST_SHORT    ?= $(if $(filter true,$(short)), -short,)
GO_TEST_FLAGS    ?= -count=1 $(GO_FLAGS) $(VERB_SHORT) $(GO_TEST_FAIL) $(GO_TEST_SHORT)
GO_TEST_PATTERNS ?= -run="$(run)"
GO_TEST_ARGS     ?= $(GO_PKG) $(GO_TEST_FLAGS) $(GO_TEST_PATTERNS)

OPENAI_API_AUTH ?= "Authorization: Bearer $(OPENAI_API_KEY)"

# Disable raw mode because it interferes with our TTY detection.
GOW_HOTKEYS := -r=false

# Repo: https://github.com/mitranim/gow.
# Install: `go install github.com/mitranim/gow@latest`.
GOW ?= gow $(CLEAR_SHORT) $(VERB_SHORT) $(GOW_HOTKEYS)

# Repo: https://github.com/mattgreen/watchexec.
# Install: `brew install watchexec`.
WATCH ?= watchexec $(CLEAR_SHORT) -d=0 -r -n --stop-timeout=1

OK = echo [$@] ok

GIT_SHA := $(shell git rev-parse --short=8 HEAD)

DOCKER_LABEL ?= agent
DOCKER_TAG_WITH_GIT_SHA ?= $(DOCKER_LABEL):$(GIT_SHA)
DOCKER_TAG_LATEST ?= $(DOCKER_LABEL):latest

# TODO: if appropriate executable does not exist, print install instructions.
ifeq ($(OS),Windows_NT)
	GO_WATCH ?= $(WATCH) $(GO_WATCH_FLAGS) -- go
else
	GO_WATCH ?= $(GOW) $(GO_WATCH_FLAGS)
endif

ifeq ($(OS),Windows_NT)
	RM_DIR = if exist "$(1)" rmdir /s /q "$(1)"
else
	RM_DIR = rm -rf "$(1)"
endif

ifeq ($(OS),Windows_NT)
	CP_INNER = if exist "$(1)" copy "$(1)"\* "$(2)" >nul
else
	CP_INNER = if [ -d "$(1)" ]; then cp -r "$(1)"/* "$(2)" ; fi
endif

ifeq ($(OS),Windows_NT)
	CP_DIR = if exist "$(1)" copy "$(1)" "$(2)" >nul
else
	CP_DIR = if [ -d "$(1)" ]; then cp -r "$(1)" "$(2)" ; fi
endif

default: go.run.w

go.run.w: # Run in watch mode
	$(GO_WATCH) run $(GO_RUN_ARGS)

go.run: # Run once
	go run $(GO_RUN_ARGS)

go.test.w: # Run tests in watch mode
	$(eval export)
	$(GO_WATCH) test $(GO_TEST_ARGS)

go.test: # Run tests once
	$(eval export)
	go test $(GO_TEST_ARGS)

go.vet.w: # Run `go vet` in watch mode
	$(GO_WATCH) vet $(GO_FLAGS) $(GO_VET_FLAGS) $(GO_PKG)

go.vet: # Run `go vet` once
	go vet $(GO_FLAGS) $(GO_VET_FLAGS) $(GO_PKG)
	$(OK)

go.build: # Build executable for current platform
	go build $(GO_FLAGS) -o=$(GO_CMD) $(GO_CMD_SRC)

go.build.linux: # Build executable for Linux
	GOOS=linux GOARCH=amd64 go build -o $(GO_CMD_LINUX) $(GO_CMD_SRC)

# TODO: keep command comments, and align them vertically.
#
# Note that if we do that, `uniq` will no longer dedup lines for commands whose
# names are repeated, usually with `<cmd_name>: export <VAR> ...`. We'd have to
# skip/ignore those lines.
help:	# Print help
	echo "Available commands are listed below"
	echo "Show this help: make help"
	echo "Show command definition: make -n <command>"
	echo
	for val in $(MAKEFILE_LIST); do grep -E '^\S+:' $$val; done | sed 's/:.*//' | uniq

openai.models: # Fetch names of current OpenAI models
	curl https://api.openai.com/v1/models -H $(OPENAI_API_AUTH)

docker: docker.build docker.clean docker.run

# TODO: only echo in verbose mode.
docker.build:
	echo "build: $(DOCKER_TAG_WITH_GIT_SHA)"
	DOCKER_BUILDKIT=0 docker build --progress=plain --build-arg LABEL=$(DOCKER_LABEL) -t $(DOCKER_TAG_WITH_GIT_SHA) -t $(DOCKER_TAG_LATEST) .

docker.run:
	docker run -it -v=$(PWD):/app -w=/app --entrypoint /bin/ash $(DOCKER_TAG_LATEST)

# Deletes all untagged images built from our project.
# TODO: only keep the latest, if any.
# TODO: un-hardcode the label both in the Dockerfile and here.
docker.clean:
	docker image prune -f --filter "label=project=$(DOCKER_LABEL)"

docker.ls:
	docker images --filter "label=project=$(DOCKER_LABEL)"
