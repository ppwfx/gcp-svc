SOURCE_MAKE=. ./.make/make.sh
SHELL := /bin/bash

lint:
	@${SOURCE_MAKE} && lint

build:
	@${SOURCE_MAKE} && build

test-integration:
	@${SOURCE_MAKE} && test-integration

terraform-apply:
	@${SOURCE_MAKE} && terraform-apply