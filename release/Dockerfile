FROM alpine:3.18 as alpine
RUN apk add -U --no-cache ca-certificates

FROM scratch
WORKDIR /app
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ntfy-bridge /app/ntfy-bridge
ENTRYPOINT ["/app/ntfy-bridge"]
