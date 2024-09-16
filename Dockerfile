FROM golang:1.23.0-alpine AS builder

# Move to working directory (/build).
WORKDIR /build

# Copy the code into the container.
COPY backend/. .
RUN go mod download

# Set necessary environment variables needed for our image and build the API server.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o apiserver .

FROM scratch

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/build", "/"]

# Command to run when starting the container.
ENTRYPOINT ["/backend"]
