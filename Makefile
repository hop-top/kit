.PHONY: lint lint-go lint-ts lint-py lint-docs lint-config lint-links \
	test test-go test-go-integration test-ts test-py test-parity \
	proto openapi clients clients-ts clients-php clients-rs clients-test api \
	job-test job-integration-hatchet job-integration-restate job-integration-temporal \
	test-workflow test-hook test-release promote promote-alpha promote-beta promote-rc promote-release check

check: lint test ## Run all linters and tests (full gate)

test: test-go test-ts test-py test-workflow test-hook ## Run all tests

test-go: ## Go tests (skips long-running container tests)
	@find go cmd contracts engine examples incubator -name "go.mod" -execdir go test -short ./... -count=1 \;

test-go-integration: ## Go tests including testcontainer integration
	@find go cmd contracts engine examples incubator -name "go.mod" -execdir go test ./... -count=1 -timeout 300s \;

test-ts: ## TypeScript tests
	cd sdk/ts && pnpm vitest run --exclude src/sqlstore.test.ts

test-py: ## Python tests
	cd sdk/py && uv sync --all-extras -q && uv run pytest
	cd engine/sdk/py-kit-engine && uv sync --all-extras -q && uv run pytest

test-parity: ## Cross-language parity tests
	go test -tags parity ./go/console/cli/... -timeout 300s -count=1

lint: lint-go lint-ts lint-py lint-docs lint-config lint-links ## Run all linters

lint-go: ## Go: golangci-lint
	@find go cmd contracts engine examples incubator -name "go.mod" -execdir golangci-lint run ./... \;

lint-ts: ## TypeScript: eslint
	cd sdk/ts && pnpm eslint src/

lint-py: ## Python: ruff check + format
	cd sdk/py && uv run ruff check . && uv run ruff format --check .

lint-docs: ## Markdown: markdownlint
	npx markdownlint-cli2 "README.md" "CHANGELOG.md" "RELEASING.md" "AGENTS.md" "docs/**/*.md" "cmd/kit/README.md" "incubator/**/*.md" --config examples/spaced/.markdownlint.yaml

lint-config: ## Validate configuration files and check for broken paths
	@echo "Validating configuration files..."
	@# Check all JSON files for syntax errors
	@find . -name "*.json" -not -path "*/node_modules/*" -not -path "*/vendor/*" -exec jq . {} + > /dev/null
	@# Check release-please for broken paths
	@for p in $$(jq -r '.packages | keys[]' .github/release-please-config.json); do \
		if [ "$$p" != "." ] && [ ! -d "$$p" ]; then \
			echo "Error: release-please-config.json references non-existent path: $$p"; \
			exit 1; \
		fi; \
	done
	@# Check pnpm-workspace.yaml for broken paths
	@for p in $$(jq -r '.packages[]' pnpm-workspace.yaml 2>/dev/null || yq -r '.packages[]' pnpm-workspace.yaml 2>/dev/null || grep -E '^- ' pnpm-workspace.yaml | sed 's/^- //'); do \
		if [ ! -d "$$p" ]; then \
			echo "Error: pnpm-workspace.yaml references non-existent path: $$p"; \
			exit 1; \
		fi; \
	done
	@echo "Config validation passed."

lint-links: ## Check for broken links in documentation
	lychee --offline docs/ README.md

proto: ## Generate protobuf + Connect/gRPC stubs
# Generated files are committed for go-get compatibility.
# Re-run after changing .proto files.
	cd contracts/proto/routellm/v1 && buf generate
	cd contracts/proto/crud/v1 && buf generate

openapi: ## Print OpenAPI extraction instructions (requires running server)
	@echo "Start server, then: curl http://localhost:8080/openapi.json > openapi.json"

clients: clients-ts clients-php clients-rs ## Build all polyglot clients

clients-ts: ## Build TypeScript client
	cd sdk/ts && pnpm install && pnpm build

clients-php: ## Install PHP client dependencies
	cd sdk/experimental/php && composer install

clients-rs: ## Build Rust client
	cd sdk/experimental/rs && cargo build --features api

clients-test: ## Test all polyglot clients
	cd sdk/ts && pnpm test
	cd sdk/experimental/php && composer test
	cd sdk/experimental/rs && cargo test --features api

api: proto clients ## Generate protos + build all clients

test-workflow: ## Run bats unit tests for cli-demo-media workflow shell logic
	bats .github/tests/cli-demo-media.bats

test-hook: ## Run bats tests for pre-push hook
	bats .github/tests/pre-push-hook.bats

job-test:
	go test ./go/runtime/job/... -count=1

job-integration-hatchet:
	docker compose -f go/runtime/job/hatchet/testdata/docker-compose.yml up -d --wait --wait-timeout 60 || \
		(docker compose -f go/runtime/job/hatchet/testdata/docker-compose.yml down -v; exit 1)
	go test -tags hatchet ./go/runtime/job/hatchet/... -count=1 -timeout 120s || \
		(docker compose -f go/runtime/job/hatchet/testdata/docker-compose.yml down -v; exit 1)
	docker compose -f go/runtime/job/hatchet/testdata/docker-compose.yml down -v

job-integration-restate:
	docker compose -f go/runtime/job/restate/testdata/docker-compose.yml up -d --wait --wait-timeout 60 || \
		(docker compose -f go/runtime/job/restate/testdata/docker-compose.yml down -v; exit 1)
	go test -tags restate ./go/runtime/job/restate/... -count=1 -timeout 120s || \
		(docker compose -f go/runtime/job/restate/testdata/docker-compose.yml down -v; exit 1)
	docker compose -f go/runtime/job/restate/testdata/docker-compose.yml down -v

job-integration-temporal:
	docker compose -f go/runtime/job/temporal/testdata/docker-compose.yml up -d --wait --wait-timeout 60 || \
		(docker compose -f go/runtime/job/temporal/testdata/docker-compose.yml down -v; exit 1)
	go test -tags temporal ./go/runtime/job/temporal/... -count=1 -timeout 120s || \
		(docker compose -f go/runtime/job/temporal/testdata/docker-compose.yml down -v; exit 1)
	docker compose -f go/runtime/job/temporal/testdata/docker-compose.yml down -v

test-release: ## Run e2e tests for release scripts
	bash scripts/test-release-e2e.sh

promote: ## Interactive release promotion
	@./scripts/promote-release.sh

promote-alpha: ## Promote main to alpha
	./scripts/promote-release.sh alpha

promote-beta: ## Promote alpha to beta
	./scripts/promote-release.sh beta

promote-rc: ## Promote beta to rc
	./scripts/promote-release.sh rc

promote-release: ## Promote rc to stable release
	./scripts/promote-release.sh release
