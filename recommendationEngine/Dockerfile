# Build stage AS builder 1.20.1-alpine3.17
FROM golang:alpine AS builder

LABEL maintainer="Alabi Mujeeb"

WORKDIR /engine

# Copy go mod and sum files 
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed 
# RUN go mod download 

COPY . .

RUN go build -o main main.go

# Run stage
# FROM alpine:latest
# WORKDIR /engine
# COPY --from=builder /engine/main .

# RUN main --no-cache add ca-certificates

EXPOSE 5000

CMD ["/engine/main"]

