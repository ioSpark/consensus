.PHONY: all
all: lint fmt test tailwind

.PHONY: test
test:
	pre-commit run test

.PHONY: fmt
fmt:
	pre-commit run format-html
	pre-commit run golangci-lint-fmt

.PHONY: lint
lint:
	pre-commit run golangci-lint-config-verify
	pre-commit run golangci-lint-full

.PHONY: tailwind
tailwind:
	pre-commit run tailwind

.PHONY: tailwind-watch
tailwind-watch:
	tailwindcss -i config.css -o static/style.css --watch

.PHONY: build-binary
build:
	go build -o consensus

.PHONY: run-binary
run-binary: build-binary
	./consensus

.PHONY: run
run:
	cd dev/ && docker compose up

# TODO: Be able to specify a registry
# TODO: Properly build multi-arch image
.PHONY: build-image
build-image:
	docker build -t local/consensus:latest .

.PHONY: push
push: build-image
	echo placeholder
