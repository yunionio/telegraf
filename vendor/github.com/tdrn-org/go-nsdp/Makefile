MAKEFLAGS += --no-print-directory

GOBIN ?= $(shell go env GOPATH)/bin

.DEFAULT_GOAL := check

.PHONY: deps
deps:
	go mod download -x

.PHONY: testdeps
testdeps: deps
	go install honnef.co/go/tools/cmd/staticcheck@2024.1.1

.PHONY: tidy
tidy:
	go mod verify
	go mod tidy

.PHONY: test
test: testdeps
	go test -v -coverpkg=./... -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: vet
vet: testdeps
	go vet ./...

.PHONY: staticcheck
staticcheck: testdeps
	$(GOBIN)/staticcheck ./...

.PHONY: check
check: test vet staticcheck

.PHONY: clean
clean:
	go clean ./...
