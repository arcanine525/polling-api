# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/main.go

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p uploads

ENV PORT=8080
EXPOSE 8080

ENV GIN_MODE=release

CMD ["./server"]
