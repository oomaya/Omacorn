NAME := omacorn
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.4.0")
BINDIR ?= $(HOME)/.local/bin
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build test lint static cross install update clean

all: build

build:
	go build -ldflags="$(LDFLAGS)" -o $(NAME) .

test:
	go test -v -race ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt check failed: run gofmt -w ." && exit 1)

static:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(NAME) .

cross:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(NAME)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(NAME)-linux-arm64 .

install: static
	mkdir -p $(BINDIR)
	cp $(NAME) $(BINDIR)/$(NAME)
	ln -sf $(NAME) $(BINDIR)/dpipe
	ln -sf $(NAME) dpipe
	@echo "Installed $(NAME) to $(BINDIR)/$(NAME)"

update:
	@./scripts/update.sh

clean:
	rm -f $(NAME) $(NAME)-linux-* dpipe
