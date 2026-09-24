FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG BUILD_PATH=./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o server ${BUILD_PATH}

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

ENV PORT=34005

EXPOSE 34005

ENTRYPOINT ["./server"]