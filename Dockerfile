# Build stage
FROM golang:1.24.2-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o flux .

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ffmpeg imagemagick pandoc ca-certificates

COPY --from=builder /app/flux /usr/local/bin/flux

ENTRYPOINT ["flux"]
CMD ["--help"]
