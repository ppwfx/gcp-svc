SOURCE_MAKE=. ./.make/make.sh
SHELL := /bin/bash

lint:
	@${SOURCE_MAKE} && lint

build-docker:
	@${SOURCE_MAKE} && build-docker

test:
	@${SOURCE_MAKE} && test

push-docker:
	@${SOURCE_MAKE} && push-docker

migrate-database:
	@${SOURCE_MAKE} && migrate-database

docker-compose-user-svc:
	@${SOURCE_MAKE} && docker-compose-user-svc

deploy:
	@${SOURCE_MAKE} && deploy