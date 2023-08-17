
# Go parameters
BINARY_NAME=jira2gitlab
BINARY_UNIX=unix_$(BINARY_NAME)

all: test build
build:
			  go build -o $(BINARY_NAME) -v
				chmod +x $(BINARY_NAME)
test:
				go test -v ./...
clean:
			  go clean
				rm -f $(BINARY_NAME)
				rm -f $(BINARY_UNIX)
run:
				go clean
				go build -o $(BINARY_NAME) -v
				chmod +x $(BINARY_NAME)
				./$(BINARY_NAME)

# Cross compilation
# build-linux:
# 				CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
# docker-build:
# 				docker run --rm -it -v "$(GOPATH)":/go -w /go/src/github.com/patmizi/go-cli-boilerplate golang:latest go build -o "$(BINARY_UNIX)" -v