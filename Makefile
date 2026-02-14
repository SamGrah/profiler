SHELL := /bin/bash

.PHONY: build test test-one fmt vet run check

build:
	dagger call build

test:
	dagger call test

test-one:
	@if [[ -z "$(PKG)" || -z "$(NAME)" ]]; then \
		echo "Usage: make test-one PKG=./internal/service NAME='^TestCarServiceUpdate$$'"; \
		exit 1; \
	fi
	dagger call test-one --pkg="$(PKG)" --name="$(NAME)"

fmt:
	dagger call fmt export --path=.

vet:
	dagger call vet

run:
	dagger call run up --ports=8080:8080

check:
	dagger check
