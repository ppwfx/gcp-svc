#!/usr/bin/env bash

set -eox pipefail

TAG=$(git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || echo "dev")

function build-docker {
    REV=$(git rev-parse HEAD)
    DATE=$(date "+%Y-%m-%d")
    VERSION=${TAG//v}

    docker build -f .make/user-svc.Dockerfile \
        --label=org.opencontainers.image.created=$DATE \
        --label=org.opencontainers.image.name=gcp-svc/user-svc \
        --label=org.opencontainers.image.revision=$REV \
        --label=org.opencontainers.image.version=$TAG \
        --label=org.opencontainers.image.source=https://github.com/ppwfx/gcp-svc \
        --label=repository=http://github.com/ppwfx/gcp-svc \
        --tag user-svc/user-svc:$TAG \
        --tag gcr.io/user-svc/user-svc:$TAG .
}

function push-docker {
    gcloud auth configure-docker

    docker push gcr.io/user-svc/user-svc:$TAG
}

function docker-compose-user-svc {
    docker-compose -f .make/docker-compose.yml up --build --remove-orphans --renew-anon-volumes
}

function test-integration-remote {
    go test ./pkg/communication -tags=integration -remote -user-svc-addr https://$(cd .make && terraform output -json | jq -r '.["user-svc-url"]["value"]')
}

function apply-terraform {
    cd ./.make
    terraform apply -auto-approve -var user-svc-version=$TAG
}

function deploy {
    cd ./.make
    terraform init
    terraform plan -target=module.user-svc -var user-svc-version=$TAG
    terraform apply -target=module.user-svc -auto-approve -var user-svc-version=$TAG
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