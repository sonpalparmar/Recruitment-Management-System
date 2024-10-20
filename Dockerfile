# Use the official Go image as the builder
FROM golang:1.21 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code and .env file into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/main.go

# Start a new stage from scratch
FROM alpine:latest  

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy the .env file
COPY --from=builder /app/.env ./

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
