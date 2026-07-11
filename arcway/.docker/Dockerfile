ARG DEBIAN_IMAGE=arm64v8/debian:bookworm-slim
ARG GO_VERSION=1.26.1

FROM ${DEBIAN_IMAGE} AS build

ARG GO_VERSION

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=${GOROOT}/bin:${GOPATH}/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates curl git make \
	&& rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-arm64.tar.gz" -o /tmp/go.tgz \
	&& rm -rf "${GOROOT}" \
	&& tar -C /usr/local -xzf /tmp/go.tgz \
	&& rm -f /tmp/go.tgz

WORKDIR /workspace

COPY arcway /workspace/arcway
COPY core /workspace/core

WORKDIR /workspace/arcway

RUN make build

FROM ${DEBIAN_IMAGE} AS run

RUN apt-get update \
	&& apt-get install -y --no-install-recommends bash ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /workspace/arcway/.build/gateway /app/gateway
COPY --from=build /workspace/arcway/.build/gateway-controller /app/gateway-controller
COPY arcway/.docker/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/gateway /app/gateway-controller /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
