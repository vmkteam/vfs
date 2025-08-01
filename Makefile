LOCAL_PKG := github.com/vmkteam/vfs
MAIN := $(LOCAL_PKG)/cmd/vfssrv

PKG := `go list -f {{.Dir}} ./...`

LINT_VERSION := v2.1.5

ifeq ($(RACE),1)
	GOFLAGS=-race
endif

tools:
	@go install github.com/vmkteam/mfd-generator@latest
	@go install github.com/vmkteam/zenrpc/v2/zenrpc@latest
	@curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin ${LINT_VERSION}

fmt:
	@golangci-lint fmt

lint:
	@golangci-lint version
	@golangci-lint config verify
	@golangci-lint run

build:
	@CGO_ENABLED=0 go build $(GOFLAGS) -o vfssrv $(MAIN)

run:
	@echo "Compiling"
	@go run $(GOFLAGS) $(MAIN) -dev $(fl)

test:
	@go test $(GOFLAGS) ./...

test-short:
	@go test $(GOFLAGS) -test.short -test.run="Test[^D][^B]" ./...

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
