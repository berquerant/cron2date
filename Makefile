GOMOD = go mod
GOBUILD = go build -trimpath -race -v
GOTEST = go test -v -cover -race

BIN = dist/cron2date
CMD = "./cmd/cron2date"

.PHONY: build
build: $(BIN)

.PHONY: $(BIN)
$(BIN):
	$(GOBUILD) -o $@ $(CMD)

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: init
init:
	$(GOMOD) tidy

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	go tool golangci-lint run ./...

.PHONY: clean
clean:
	rm -rf dist
