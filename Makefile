SOURCE_MAKE=. ./.make/make.sh
SHELL := /bin/bash

build:
	@${SOURCE_MAKE} && build

test-integration:
	@${SOURCE_MAKE} && test-integration