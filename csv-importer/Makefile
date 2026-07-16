GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOFMT = $(GOCMD) fmt
GOVET = $(GOCMD) vet
PACKAGENAME = libra-etd-proquest-import
BINNAME = $(PACKAGENAME)

build: darwin-bin

linux: linux-bin

all: darwin-bin linux-bin

darwin-bin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -race -o bin/$(BINNAME).darwin *.go

linux-bin:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/$(BINNAME).linux *.go

clean:
	$(GOCLEAN) .
	rm -rf bin

dep:
	$(GOGET) -u
	$(GOMOD) tidy
	$(GOMOD) verify

fmt:
	$(GOFMT)

vet:
	$(GOVET)

check:
	go install honnef.co/go/tools/cmd/staticcheck
	$(HOME)/go/bin/staticcheck -checks all,-S1002,-ST1003 *.go
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	$(GOVET) -vettool=$(HOME)/go/bin/shadow ...
