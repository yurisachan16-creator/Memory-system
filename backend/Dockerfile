FROM golang:1.23-alpine AS builder

WORKDIR /src/backend

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM alpine:3.20

WORKDIR /app/backend

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/server ./server
COPY backend/migrations ./migrations
COPY backend/config ./config

EXPOSE 8080

CMD ["./server"]

