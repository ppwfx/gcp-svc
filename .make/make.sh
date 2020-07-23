#!/usr/bin/env bash

set -eox pipefail

TAG=$(git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || echo "dev")

function build-docker {
    docker build -f .make/user-svc.Dockerfile \
        --tag user-svc/user-svc:$TAG \
        --tag gcr.io/user-svc/user-svc:$TAG .
}

function push-docker {
    gcloud auth configure-docker

    docker push gcr.io/user-svc/user-svc:$TAG
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
    terraform plan -target=module.user-svc -var svc-version=$TAG
    terraform apply -target=module.user-svc -auto-approve -var svc-version=$TAG
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