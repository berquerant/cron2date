GOMOD = go mod
GOBUILD = go build -trimpath -race -v
GOTEST = go test -v -cover -race

ROOT = $(shell git rev-parse --show-toplevel)
BIN = dist/cron2date
CMD = "./cmd/cron2date"
RBIN = dist/date2cron
RCMD = "./cmd/date2cron"

.PHONY: build
build: $(BIN) $(RBIN)

.PHONY: $(BIN)
$(BIN):
	$(GOBUILD) -o $@ $(CMD)

.PHONY: $(RBIN)
$(RBIN):
	$(GOBUILD) -o $@ $(RCMD)

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: init
init:
	$(GOMOD) tidy

.PHONY: vuln
vuln:
	go run golang.org/x/vuln/cmd/govulncheck ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: clean
clean:
	find . -name "*_generated.go" -type f -delete
