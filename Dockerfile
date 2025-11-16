FROM golang:1.24.10-alpine AS builder

WORKDIR /app

# install dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o pr-service ./cmd/main.go

# Stage 2: final
FROM alpine:latest

RUN apk --no-cache add curl

WORKDIR /root

# copy compiled service from previous stage
COPY --from=builder /app/pr-service .

# app start
CMD [ "./pr-service" ]