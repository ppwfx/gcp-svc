FROM bitnami/minideb:latest

RUN install_packages ca-certificates

COPY dist/user-svc user-svc