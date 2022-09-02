WORK_DIR = $(shell pwd)
OWNER = feature
REPOSITORY = 685249416972.dkr.ecr.ap-south-1.amazonaws.com/$(OWNER)
APP_NAME = wallmount-job
SERVICE = wallmount-job
MODULE_NAME = futuner
GLOBAL_VERSION = latest
BASE_VERSION = base
GIT = $(shell which git)
GO := $(shell which go)
DOCKER := $(shell which docker)
EXEC = $(WORK_DIR)/
ifeq ($(APP_VERSION),)
	APP_VERSION = $(shell $(GIT) describe --tags --abbrev=0 2>/dev/null)
endif
TAG_VERSION = $(REPOSITORY)/$(APP_NAME):$(APP_VERSION)
BASE_TAG = $(REPOSITORY)/$(APP_NAME):$(BASE_VERSION)-$(APP_VERSION)
BASE_GLOBAL = $(REPOSITORY)/$(APP_NAME):$(BASE_VERSION)-$(GLOBAL_VERSION)
TAG_GLOBAL = $(REPOSITORY)/$(APP_NAME):$(GLOBAL_VERSION)
COPY = $(shell which cp)

prepare_build:
	@$(GO) get -d -v
	@$(GO) build -o $(SERVICE) .

build_docker:
# 	@$(DOCKER) pull $(BASE_GLOBAL)
# 	@$(DOCKER) tag $(BASE_GLOBAL) $(MODULE_NAME):$(BASE_VERSION)-$(GLOBAL_VERSION)
	@$(DOCKER) build --build-arg APP_VERSION=$(APP_VERSION) --build-arg MODULE_NAME=$(MODULE_NAME) --build-arg APP_NAME=$(APP_NAME) -t $(TAG_VERSION) -t $(TAG_GLOBAL) .

run_docker:
	@$(DOCKER) run -d -p 8080:8080 --env-file env $(TAG_GLOBAL)

push_docker: build_docker
	@$(DOCKER) push $(TAG_VERSION)
	@$(DOCKER) push $(TAG_GLOBAL)
