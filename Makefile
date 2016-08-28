# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all push container clean

BIN = spartakus
GO_PKG = k8s.io/spartakus

REGISTRY ?= gcr.io/google_containers
IMAGE = $(REGISTRY)/$(BIN)-$(ARCH)

COMMIT = $(shell git rev-parse HEAD)
TAG = $(shell git describe --exact-match --abbrev=0 --tags $(COMMIT) 2>/dev/null)
ifeq ($(TAG),)
    TAG = git-$(COMMIT)
endif
DIRTY = $(shell test -z "$$(git diff --shortstat 2>/dev/null)" || echo +dirty)
VERSION = $(TAG)$(DIRTY)

# Architectures supported: amd64, arm, arm64 and ppc64le
ARCH ?= amd64

# TODO: get a base image for non-x86 archs
#       arm arm64 ppc64le
ALL_ARCH = amd64

BUILD_IMAGE ?= golang:1.7.0-alpine

# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: all-build

sub-container-%:
	$(MAKE) ARCH=$* container

sub-push-%:
	$(MAKE) ARCH=$* push

all-build: $(addprefix bin/$(BIN)-,$(ALL_ARCH))

all-container: $(addprefix sub-container-,$(ALL_ARCH))

all-push: $(addprefix sub-push-,$(ALL_ARCH))

build: bin/$(BIN)-$(ARCH)

bin/$(BIN)-$(ARCH): FORCE
	mkdir -p bin/$(ARCH)
	mkdir -p .go/src/$(GO_PKG) .go/pkg .go/bin .go/std/$(ARCH)
	docker run                                                             \
	    -u $$(id -u):$$(id -g)                                             \
	    -v $$(pwd)/.go:/go                                                 \
	    -v $$(pwd):/go/src/$(GO_PKG)                                       \
	    -v $$(pwd)/bin/$(ARCH):/go/bin                                     \
	    -v $$(pwd)/.go/std/$(ARCH):/usr/local/go/pkg/linux_$(ARCH)_static  \
	    $(BUILD_IMAGE)                                                     \
	    /bin/sh -c "                                                       \
	        cd /go/src/$(GO_PKG) &&                                        \
	        CGO_ENABLED=0                                                  \
	        go install                                                     \
	        -installsuffix static                                          \
	        -ldflags '-X main.VERSION=$(VERSION)'                          \
	        ./...                                                          \
	    "

container: .container-$(ARCH)
.container-$(ARCH): bin/$(BIN)-$(ARCH)
	docker build -t $(IMAGE):$(VERSION) --build-arg ARCH=$(ARCH) .
	touch $@

push: .push-$(ARCH)
.push-$(ARCH): .container-$(ARCH)
	gcloud docker push $(IMAGE):$(VERSION)
	touch $@

clean:
	rm -rf .container-* .push-* .go bin

FORCE:
