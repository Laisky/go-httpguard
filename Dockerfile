FROM golang:1.10.1-alpine3.7 AS gobin
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh ca-certificates && \
    update-ca-certificates
RUN mkdir -p /go/src/pateo.com/go-httpguard
ADD . /go/src/pateo.com/go-httpguard
WORKDIR /go/src/pateo.com/go-httpguard
RUN go build ./entrypoints/main.go

FROM alpine:3.7
COPY --from=gobin /go/src/pateo.com/go-httpguard/main go-httpguard
COPY --from=gobin /etc/ssl/certs /etc/ssl/certs
CMD ["./go-httpguard"]
