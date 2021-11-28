FROM docker-registry.wikimedia.org/golang1.17:latest as builder

USER 0
RUN apt-get update && apt-get install -y git ca-certificates make

WORKDIR /srv/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make build

# Runtime image
FROM scratch AS base
USER nobody

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /srv/app/delete-crashing-pods /srv/app/delete-crashing-pods

ENTRYPOINT ["/srv/app/delete-crashing-pods"]
