export GO111MODULE=auto

ENABLE_CGO= CGO_ENABLED=0
GO_LD_FLAGS= "-s -w"
GO_BUILD_TAGS= -tags ${GOTAGS}

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))
RELEASE_DIR := ${MKFILE_DIR}bin
TARGET_SERVER= ${RELEASE_DIR}/nmid-registry

build_server:
	@echo "build nmid-registry"
	cd ${MKFILE_DIR} && \
	${ENABLE_CGO} go build ${GO_BUILD_TAGS} -v -trimpath -ldflags ${GO_LD_FLAGS} \
	-o ${TARGET_SERVER} ${MKFILE_DIR}cmd/server

build: build_server

run: build_server