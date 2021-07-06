NAME := vfs
LOCAL_PKG := github.com/vmkteam/vfs
MAIN := cmd/vfssrv/main.go

PKG := `go list -mod=vendor -f {{.Dir}} ./...`

ifeq ($(RACE),1)
	GOFLAGS=-race
endif

VERSION?=$(shell git version > /dev/null 2>&1 && git describe --dirty=-dirty --always 2>/dev/null || echo NO_VERSION)
LDFLAGS=-ldflags "-X=main.version=$(VERSION)"

all: tools rebuild

tools:
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

fmt:
	@goimports -local ${LOCAL_PKG} -l -w $(PKG)

lint:
	@golangci-lint run -c .golangci.yml

rebuild:
	@CGO_ENABLED=0 go build -a $(LDFLAGS) $(GOFLAGS) -o vfssrv $(MAIN)
	@go mod vendor

build:
	@CGO_ENABLED=0 go build -mod=vendor $(LDFLAGS) $(GOFLAGS) -o vfssrv $(MAIN)

run:
	@echo "Compiling"
	@go run -mod=vendor $(LDFLAGS) $(GOFLAGS) $(MAIN)

test:
	@go test -mod=vendor $(LDFLAGS) $(GOFLAGS) ./...

test-short:
	@go test -mod=vendor $(LDFLAGS) $(GOFLAGS) -test.short -test.run="Test[^D][^B]" ./...

generate:
	@go generate

mod:
	@go mod tidy
	@go mod vendor

mfd-xml:
	@mfd-generator xml -c "postgres://postgres:postgres@localhost:5432/vfs?sslmode=disable" -m ./docs/model/vfs.mfd -n "vfs:vfsFiles,vfsFolders,vfsHashes"

mfd-model:
	@mfd-generator model -m ./docs/model/vfs.mfd -p db -o ./db

mfd-repo:
	@mfd-generator repo -m ./docs/model/vfs.mfd -p db -o ./db