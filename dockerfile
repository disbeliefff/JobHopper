# # Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# # Set the working directory inside the container
# WORKDIR /app

# # Copy the Go module files and download the dependencies
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy the source code into the container
# COPY . .

# # Build the Go binary
# RUN go build -o myapp .

# # Stage 2: Create the final image with the Go binary
# FROM alpine:latest

# # Set the working directory inside the container
# WORKDIR /app

# # Copy the Go binary from the builder stage
# COPY --from=builder /app/myapp .

# # Expose the port the application will run on
# EXPOSE 8080

# # Command to run the application
# CMD ["./myapp"]
