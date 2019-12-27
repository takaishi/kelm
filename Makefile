APP=kelm
REGISTRY?=rtakaishi
COMMIT_SHA=$(shell git rev-parse --short HEAD)

export GO111MODULE=on

.PHONY: build
## build: build the application
build: clean
	go build -o ${APP} main.go

.PHONY: run
## run: runs go run main.go
run:
	go run -race main.go

.PHONY: clean
## clean: cleans the binary
clean:
	go clean

.PHONY: deps
## deps: Installing dependencies
deps:
	go get -d
	go mod tidy

.PHONY: depsdev
## depsdev: Installing dependencies for development
depsdev:
	GO111MODULE=off go get -u \
	golang.org/x/lint/golint

.PHONY: docker-build
## docker-build: build container image
docker-build:
	echo docker build -t ${APP} .
	echo docker tag ${APP} ${REGISTRY}/${APP}:${COMMIT_SHA}

.PHONY: docker-push
## docker-push: push container image
docker-push: docker-build
	docker push ${REGISTRY}/${APP}:${COMMIT_SHA}

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
