FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/reward-system ./cmd/server

FROM alpine:3.20
RUN adduser -D -g '' appuser && apk add --no-cache ca-certificates
USER appuser
WORKDIR /app
COPY --from=builder /out/reward-system /app/reward-system
EXPOSE 8080
ENTRYPOINT ["/app/reward-system"]