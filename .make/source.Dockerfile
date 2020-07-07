FROM golang:1.14

WORKDIR /go/src/github.com/ppwfx/user-svc

COPY go.mod go.mod

RUN go mod tidy

COPY . .