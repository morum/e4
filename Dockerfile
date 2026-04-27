FROM golang:1.25-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/e4 ./cmd/e4

FROM debian:bookworm-slim

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates \
	&& rm -rf /var/lib/apt/lists/* \
	&& useradd --system --create-home --home-dir /home/e4 e4 \
	&& mkdir -p /data \
	&& chown -R e4:e4 /data /home/e4

COPY --from=build /out/e4 /usr/local/bin/e4

USER e4
EXPOSE 2222

ENTRYPOINT ["e4"]
CMD ["serve", "--listen", ":2222", "--host-key", "/data/e4_host_key"]
