#!/usr/bin/env bash

set -eox pipefail

function build {
    GOOS=linux GOARCH=amd64 go build -o dist/user-svc cmd/user-svc/main.go

    docker build -f .make/user-svc.Dockerfile --tag user-svc/user-svc:latest .

    docker build -f .make/source.Dockerfile --tag user-svc/source:latest .
}

function test-integration {
    docker-compose -f .make/test-integration.yaml up --abort-on-container-exit

    docker-compose -f .make/test-integration.yaml down

    docker-compose -f .make/test-integration.yaml rm -f -v
}