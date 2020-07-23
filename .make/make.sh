#!/usr/bin/env bash

set -eox pipefail

function build-docker {
    docker build -f .make/user-svc.Dockerfile \
        --tag user-svc/user-svc:latest \
        --tag gcr.io/user-svc/user-svc:latest .
}

function push-docker {
    gcloud auth configure-docker

    docker push gcr.io/user-svc/user-svc:latest
}

function migrate-database {
    cd ./.make
    terraform init
    terraform plan -target=null_resource.migrate_user-svc
    terraform apply -target=null_resource.migrate_user-svc -auto-approve
}

function deploy {
    cd ./.make
    terraform init
    terraform plan -target=module.user-svc
    terraform apply -target=module.user-svc -auto-approve
}

function lint {
    go vet ./...
    go fmt ./...
    go fix ./...
    gosec ./...
}

function test {
    go test ./... -tags="unit,integration" -v
}