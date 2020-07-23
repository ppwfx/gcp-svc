SOURCE_MAKE=. ./.make/make.sh
SHELL := /bin/bash

lint:
	@${SOURCE_MAKE} && lint

build:
	@${SOURCE_MAKE} && build

test:
	@${SOURCE_MAKE} && test

push-docker:
	@${SOURCE_MAKE} && push-docker

migrate-database:
	@${SOURCE_MAKE} && migrate-database

deploy:
	@${SOURCE_MAKE} && deploy