FROM golang:1.13.3
WORKDIR /go/src/github.com/ppwfx/user-svc/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /dist/user-svc cmd/user-svc/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /dist/user-svc user-svc
CMD ["./user-svc"]