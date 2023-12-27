# Use a minimal base image with Go installed
FROM golang:alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the local code to the container
COPY . .

# Build the Go application
RUN go build -o main .

# Use a minimal base image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose port 8000 for the Go application
EXPOSE 8000

# Run the Go application
CMD ["./main"]
