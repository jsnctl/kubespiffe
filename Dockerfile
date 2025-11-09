FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./kubespiffed cmd/kubespiffe/main.go
FROM scratch
COPY --from=builder /build/kubespiffed /app/
ENTRYPOINT ["/app/kubespiffed"]

