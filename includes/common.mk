AUTHOR                  ?= The corba-go Authors
COPYRIGHT_YEAR          ?= 2026-
COPYRIGHT_FILES         ?= $$(find . -name "*.go" -print | grep -v "/vendor/")
GO                      ?= go
DEFAULT_GOALS           ?= fmt set-license go-licenses-check goimports lint vulncheck test
GOLANG_CI_LINT_VERSION  ?= v2.11.4
TEXTLINT_ACTION_VERSION ?= v0.1.0

.DEFAULT_GOAL = default

.PHONY: test
test:
	@echo "running 'go test'..."
	TESTACC= $(GO) test ./... $(TESTARGS) -v -timeout=120m -parallel=8 -race;

.PHONY: testacc
testacc:
	@echo "running 'go test' with TESTACC=1..."
	TESTACC=1 $(GO) test ./... $(TESTARGS) --tags=acctest -v -timeout=120m -parallel=8 ;

.PHONY: dev-tools
dev-tools:
	$(GO) install github.com/rinchsan/gosimports/cmd/gosimports@latest
	$(GO) install golang.org/x/tools/cmd/stringer@latest
	$(GO) install github.com/sacloud/addlicense@latest
	$(GO) install github.com/client9/misspell/cmd/misspell@latest
	$(GO) install github.com/google/go-licenses@latest
	$(GO) install github.com/rhysd/actionlint/cmd/actionlint@latest
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANG_CI_LINT_VERSION)

.PHONY: goimports
goimports: fmt
	@echo "running gosimports..."
	@gosimports -l -w .

.PHONY: fmt
fmt:
	@echo "running gofmt..."
	@find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

.PHONY: godoc
godoc:
	godoc -http=localhost:6060

.PHONY: lint
lint: lint-go

.PHONY: lint-go
lint-go:
	@echo "running golanci-lint..."
	@golangci-lint run --fix ./...

.PHONY: set-license
set-license:
	@addlicense -c "$(AUTHOR)" -y "$(COPYRIGHT_YEAR)" $(COPYRIGHT_FILES)

.PHONY: go-licenses-check
go-licenses-check:
	@echo "running go-licenses..."
	@go-licenses check ./.../

.PHONY: vulncheck
vulncheck:
	@echo "running govulncheck..."
	@govulncheck ./.../