FROM golang:1.16.2-alpine3.13 as builder

RUN apk add --no-cache make build-base
RUN apk add --no-cache git

WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source files
COPY . .

# Build microservices binaries
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -race -o server ./cryptocompare/cmd

FROM alpine:latest

ARG APP_DATA_DIR=/app/data
RUN mkdir -p ${APP_DATA_DIR}

# Copy microservices binaries to clear image
COPY --from=builder /app/server .

EXPOSE 8000

# Mount data and certificate directories
VOLUME ["${APP_DATA_DIR}"]

# Run unmodifiable binary with gathered local server
ENTRYPOINT ["./server", "-c", "/app/data/config.json"]
