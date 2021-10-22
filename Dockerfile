FROM migrate/migrate:v4.15.0 as dbmigrations

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY db/migrations /migrations

USER "nobody"

# Blacklister build image
FROM golang:1.16.9-buster as build

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY . /go/src/blacklister
WORKDIR /go/src/blacklister
RUN make build-all

# Blacklister production image
FROM debian:10-slim as prod

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# hadolint ignore=DL3008,DL3009
RUN apt-get update &&\
  apt-get install -y --no-install-recommends \
    ca-certificates

COPY --from=build /go/src/blacklister/build/blacklister-linux-amd64 /usr/local/bin/blacklister
RUN chmod u+x /usr/local/bin/blacklister

ENTRYPOINT ["/usr/local/bin/blacklister"]
USER "nobody"
