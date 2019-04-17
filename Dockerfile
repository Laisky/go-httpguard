FROM golang:1.12.1-alpine3.9 AS gobin

RUN apk update && apk upgrade && \
    apk add --no-cache gcc git build-base ca-certificates && \
    update-ca-certificates

RUN go get -u github.com/Laisky/go-httpguard
WORKDIR /go/src/github.com/Laisky/go-httpguard

RUN go mod download
RUN go build -a --ldflags '-extldflags "-static"' entrypoints/main.go

FROM alpine:3.9
COPY --from=gobin /go/src/github.com/Laisky/go-httpguard/main go-httpguard
COPY --from=gobin /etc/ssl/certs /etc/ssl/certs
CMD ["./go-httpguard"]
