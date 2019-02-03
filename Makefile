PKG := `go list vfs/... | grep -v /vendor/`
MAIN := cmd/vfssrv/main.go

ifeq ($(RACE),1)
	GOFLAGS=-race
endif

VERSION?=$(shell git version > /dev/null 2>&1 && git describe --dirty=-dirty --always 2>/dev/null || echo NO_VERSION)
LDFLAGS=-ldflags "-X=main.version=$(VERSION)"

all: tools rebuild

tools:
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

fmt:
	@gofmt -l -w -s `go list -f {{.Dir}} vfs/... | grep -v /vendor/`

vet:
	@go vet $(PKG)

lint:
	@golangci-lint run -c .golangci.yml

rebuild:
	@go build -a $(LDFLAGS) $(GOFLAGS) -o vfssrv $(MAIN)

build:
	@go build $(LDFLAGS) $(GOFLAGS) -o vfssrv $(MAIN)

run:
	@echo "Compiling"
	@go run $(LDFLAGS) $(GOFLAGS) $(MAIN) -verbose

test:
	@go test $(LDFLAGS) $(GOFLAGS) $(PKG)

test-short:
	@go test $(LDFLAGS) $(GOFLAGS) -test.short -test.run="Test[^D][^B]" $(PKG)

gen:
	go generate