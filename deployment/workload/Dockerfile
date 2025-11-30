FROM alpine:latest
RUN apk add --no-cache curl openssl jq
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
