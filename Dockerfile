# Stage 1: Build frontend assets
FROM node:22-alpine AS css
WORKDIR /app

# Install TailwindCSS
RUN npm install -g tailwindcss

# Copy all project files for Tailwind processing
COPY . .

# Generate TailwindCSS output
RUN npx tailwindcss -o ./public/static/css/tw.css --minify

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS build
WORKDIR /app

# Install only necessary dependencies
RUN apk add --no-cache git  # Removed 'make'

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy and cache Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source files
COPY . .

# Copy Tailwind-generated CSS from frontend stage
COPY --from=css /app/public/static/css/tw.css /app/public/static/css/tw.css

# Set environment variables
ARG GOOS=linux
ARG GOARCH=amd64

# Run code generation
RUN go run cmd/generate/main.go && templ generate

# Build the Go binary
RUN GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-s -w" -o bin/main cmd/server/main.go

# Stage 3: Create minimal runtime container
FROM alpine:latest
WORKDIR /app

# Remove ca-certificates if not needed
# Install only if your app needs HTTPS/TLS
RUN apk add --no-cache ca-certificates

# Copy built binary
COPY --from=build /app/bin/main /app/main

# Ensure the binary is executable
RUN chmod +x /app/main

# Copy static assets and generated content
COPY --from=build /app/public /app/public
COPY --from=build /app/generated /app/generated
COPY --from=build /app/content /app/content

# Expose the application port
EXPOSE 8082

CMD ["/app/main"]
