.PHONY: build clean test run install

BINARY_NAME=goani
VERSION?=dev
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS=-ldflags "-s -w \
	-X github.com/Yyyangshenghao/goani-cli/internal/version.Version=$(VERSION) \
	-X github.com/Yyyangshenghao/goani-cli/internal/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/Yyyangshenghao/goani-cli/internal/version.BuildDate=$(BUILD_DATE)"

# 默认构建当前平台
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/goani

# 构建所有平台
build-all:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/goani
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/goani
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/goani
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/goani
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/goani

# 运行测试
test:
	go run test/source/main.go
	go run test/player/main.go

# 运行
run:
	go run cmd/goani/main.go

# 安装到 GOPATH
install:
	go install $(LDFLAGS) ./cmd/goani

# 清理
clean:
	rm -rf bin/

# 格式化代码
fmt:
	go fmt ./...

# 静态检查
lint:
	go vet ./...
