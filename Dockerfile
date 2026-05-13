# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go

# Stage 2: Final lightweight image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/main .
RUN mkdir -p uploads/pdf

ENV TZ=Asia/Jakarta

EXPOSE 8080

CMD ["./main"]