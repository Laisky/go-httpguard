FROM golang:1.13.5-alpine3.11 AS gobin

RUN apk update && apk upgrade && \
    apk add --no-cache gcc git build-base ca-certificates && \
    update-ca-certificates

ENV GO111MODULE=on
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

# static build
ADD . .
RUN go build -a --ldflags '-extldflags "-static"' entrypoints/main.go

FROM alpine:3.18.3
COPY --from=gobin /app/main go-httpguard
COPY --from=gobin /etc/ssl/certs /etc/ssl/certs
ENTRYPOINT ["./go-httpguard"]
