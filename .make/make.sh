#!/usr/bin/env bash

set -eox pipefail

function build {
    go build -o dist/user-svc cmd/user-svc/main.go
}