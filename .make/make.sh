#!/usr/bin/env bash

set -eox pipefail

function build {
    GOOS=linux GOARCH=amd64 go build -o dist/user-svc cmd/user-svc/main.go

    docker build -f .make/user-svc.Dockerfile --tag user-svc/user-svc:latest --tag gcr.io/user-svc/user-svc:latest .
}

function lint {
    go vet ./...
    go fmt ./...
    go fix ./...
    gosec ./...
}