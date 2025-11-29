FROM alpine:latest
RUN apk add --no-cache curl openssl
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
