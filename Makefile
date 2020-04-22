.PHONY: test clean qtest deploy dist linter
APP_VERSION:=$(shell cat VERSION | head -1)
GIT_COMMIT:=$(shell git describe --dirty --always)
GIT_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD -- | head -1)
BUILD_USER:=$(shell whoami)
BUILD_DATE:=$(shell date +"%Y-%m-%d")
BUILD_OS:=linux
BUILD_ARCH:=amd64
ALIAS=go-get-aws-keys
BINARY:=go-get-aws-keys
VERBOSE:=-v
PROJECT=github.com/greenpau/$(ALIAS)
#PKG_DIR=pkg/$(ALIAS)
PKG_DIR=pkg/client
CLI_DIR=cmd/client
DIST_DIR=./dist/$(BINARY)-$(APP_VERSION).$(BUILD_OS)-$(BUILD_ARCH)
DIST_DIR_NAME=$(BINARY)-$(APP_VERSION).$(BUILD_OS)-$(BUILD_ARCH)
DIST_ARCHIVE_CMD=tar -cvzf
DIST_ARCHIVE_FILE_EXT=tar.gz
DIST_ARCHIVE_EXE=
ifeq ($(BUILD_OS), windows)
  DIST_ARCHIVE_CMD=zip -r
  DIST_ARCHIVE_FILE_EXT=zip
  DIST_ARCHIVE_EXE=.exe
endif

all:
	@echo "Version: $(APP_VERSION), Branch: $(GIT_BRANCH), Revision: $(GIT_COMMIT)"
	@echo "Build for $(BUILD_OS)-$(BUILD_ARCH) on $(BUILD_DATE) by $(BUILD_USER)"
	@rm -rf ./bin/$(BUILD_OS)-$(BUILD_ARCH)
	@mkdir -p bin/$(BUILD_OS)-$(BUILD_ARCH)
	@GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) CGO_ENABLED=0 go build -o ./bin/$(BUILD_OS)-$(BUILD_ARCH)/$(BINARY)$(DIST_ARCHIVE_EXE) $(VERBOSE) \
		-gcflags="all=-trimpath=$(GOPATH)/src" \
		-asmflags="all=-trimpath $(GOPATH)/src" \
		-ldflags="-w -s -v \
		-X $(PROJECT)/$(PKG_DIR).appName=$(BINARY) \
		-X $(PROJECT)/$(PKG_DIR).appVersion=$(APP_VERSION) \
		-X $(PROJECT)/$(PKG_DIR).gitBranch=$(GIT_BRANCH) \
		-X $(PROJECT)/$(PKG_DIR).gitCommit=$(GIT_COMMIT) \
		-X $(PROJECT)/$(PKG_DIR).buildOperatingSystem=$(BUILD_OS) \
		-X $(PROJECT)/$(PKG_DIR).buildArchitecture=$(BUILD_ARCH) \
		-X $(PROJECT)/$(PKG_DIR).buildUser=$(BUILD_USER) \
		-X $(PROJECT)/$(PKG_DIR).buildDate=$(BUILD_DATE)" \
		./$(CLI_DIR)/*.go
	@echo "Done!"

linter:
	@#golint || go get -u golang.org/x/lint/golint
	@golint ./$(PKG_DIR)/*.go
	@echo "PASS: golint"

clean:
	@rm -rf bin/
	@rm -rf dist/
	@rm -rf .doc/
	@rm -rf .coverage/
	@echo "OK: clean up completed"

deploy:
	@sudo rm -rf /usr/bin/$(BINARY)
	@sudo cp ./bin/$(BUILD_OS)-$(BUILD_ARCH)/$(BINARY) /usr/bin/$(BINARY)

qtest:
	@./bin/$(BUILD_OS)-$(BUILD_ARCH)/$(BINARY) -version
	@./bin/$(BUILD_OS)-$(BUILD_ARCH)/$(BINARY) -log-level debug
	@#go test -v -run TestParseAzureAuthResponseForm ./$(PKG_DIR)/*.go
	@#go test -v -run TestParseAdfsAuthResponseForm ./$(PKG_DIR)/*.go
	@#go test -v -run TestParseAdfsAuthForm ./$(PKG_DIR)/*.go
	@#go test -v -run TestWriteAwsCredentials ./$(PKG_DIR)/*.go
	@#go test -v -run TestIsValidAwsCredentials ./$(PKG_DIR)/*.go
	@#go test -v -run TestGetVersionInfo ./$(PKG_DIR)/*.go
	@#go test -v -run TestParseAwsStsResponse ./$(PKG_DIR)/*.go
	@#./bin/$(BUILD_OS)-$(BUILD_ARCH)/$(BINARY) -log-level debug
	@#go test -v -run TestNewClient $(PKG_DIR)/*.go

dist: all
	@rm -rf  $(DIST_DIR)*
	@mkdir -p $(DIST_DIR)
	@cp ./bin/$(BUILD_OS)-$(BUILD_ARCH)/$(BINARY) $(DIST_DIR)/$(BINARY)$(DIST_ARCHIVE_EXE)
	@cp ./assets/conf/azure/$(BINARY)-config.yaml $(DIST_DIR)/
	@cp ./README.md $(DIST_DIR)/
	@cp ./LICENSE $(DIST_DIR)/
	@chmod +x $(DIST_DIR)/$(BINARY)$(DIST_ARCHIVE_EXE)
	@cd ./dist/ && $(DIST_ARCHIVE_CMD) ./$(DIST_DIR_NAME).$(DIST_ARCHIVE_FILE_EXT) ./$(DIST_DIR_NAME)

test: covdir linter
	@go test $(VERBOSE) -coverprofile=.coverage/coverage.out ./$(PKG_DIR)/*.go
	@echo "PASS: core tests"
	@echo "OK: all tests passed!"

covdir:
	@mkdir -p .coverage

coverage: covdir
	@go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html

docs:
	@mkdir -p .doc
	@godoc -html $(PROJECT)/$(PKG_DIR) > .doc/index.html
	@echo "Run to serve docs:"
	@echo "    godoc -goroot .doc/ -html -http \":5000\""

release:
	@echo "Making release"
	@if [ $(GIT_BRANCH) != "master" ]; then echo "cannot release to non-master branch $(GIT_BRANCH)" && false; fi
	@git diff-index --quiet HEAD -- || ( echo "git directory is dirty, commit changes first" && false )
	@versioned || go get -u github.com/greenpau/versioned/cmd/versioned@latest
	@versioned -patch
	@echo "Patched version"
	@git add VERSION
	@git commit -m "released v`cat VERSION | head -1`"
	@git tag -a v`cat VERSION | head -1` -m "v`cat VERSION | head -1`"
	@git push
	@git push --tags
	@#git push --delete origin vX.Y.Z
	@#git tag --delete vX.Y.Z
