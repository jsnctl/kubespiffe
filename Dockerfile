FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY . .
FROM scratch
COPY --from=builder /build/kubespiffed /app/
ENTRYPOINT ["/app/kubespiffed"]

