FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY ntfy-bridge /app/ntfy-bridge
ENTRYPOINT ["/app/ntfy-bridge"]
