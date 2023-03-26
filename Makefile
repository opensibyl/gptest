# General
WORKDIR = $(PWD)

# Go parameters
GOCMD = go
GOTEST = $(GOCMD) test

default:
	go build ./cmd/gptest

build_linux:
	GOOS=linux GOARCH=amd64 ${GOCMD} build -o gptest_linux ./cmd/gptest

build_windows:
	GOOS=windows GOARCH=amd64 ${GOCMD} build -o gptest_windows.exe ./cmd/gptest

build_macos:
	GOOS=darwin GOARCH=amd64 ${GOCMD} build -o gptest_macos ./cmd/gptest

test:
	$(GOTEST) ./...
