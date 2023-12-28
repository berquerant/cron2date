GOMOD = go mod
GOBUILD = go build -trimpath -race -v
GOTEST = go test -v -cover -race

ROOT = $(shell git rev-parse --show-toplevel)
BIN = dist/cron2date
CMD = "./cmd/cron2date"

.PHONY: $(BIN)
$(BIN):
	$(GOBUILD) -o $@ $(CMD)

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

DOCKER_RUN = docker run --rm -v "$(ROOT)":/usr/src/myapp -w /usr/src/myapp
DOCKER_IMAGE = golang:1.21

.PHONY: docker-test
docker-test:
	$(DOCKER_RUN) $(DOCKER_GO_IMAGE) $(GOTEST) ./...

.PHONY: docker-dist
docker-dist:
	$(DOCKER_RUN) $(DOCKER_GO_IMAGE) $(GOBUILD) -o $(BIN) $(CMD)
