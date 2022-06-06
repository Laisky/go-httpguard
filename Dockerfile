FROM golang:1.17.8-bullseye AS gobuild

# install dependencies
RUN apt-get update \
    && apt-get install -y --no-install-recommends g++ make gcc git build-essential ca-certificates curl \
    && update-ca-certificates

ENV GO111MODULE=on
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

# static build
ADD . .
RUN go build -a -ldflags '-w -extldflags "-static"' -o main ./cmd/gohttpguard


# copy executable file and certs to a pure container
FROM debian:bullseye

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates haveged \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=gobuild /etc/ssl/certs /etc/ssl/certs
COPY --from=gobuild /app/main /app/gohttpguard

WORKDIR /app
RUN chmod +rx -R /app && \
    adduser --disabled-password --gecos '' laisky
USER laisky

ENTRYPOINT [ "/app/gohttpguard" ]
CMD [ "-c", "/etc/gohttpguard/config.yml" ]
