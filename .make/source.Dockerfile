FROM golang:1.14

WORKDIR /go/src/github.com/ppwfx/user-svc

COPY . .

RUN go mod tidy