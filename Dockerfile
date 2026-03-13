# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# install gcc for cgo
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# enable cgo
ENV CGO_ENABLED=1

RUN go build -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache libc6-compat

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
