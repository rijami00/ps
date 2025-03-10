# Stage 1: Build Go binary
FROM golang:1.24-alpine AS build
WORKDIR /app

# Install necessary dependencies
RUN apk add --no-cache git

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source code
COPY . .

# Set environment variables
ARG GOOS=linux
ARG GOARCH=amd64

# Run code generation
RUN go run cmd/generate/main.go && templ generate

# Stage 2: Process TailwindCSS (Full Source)
FROM node:22-alpine AS css
WORKDIR /app

# Copy package.json and install dependencies
COPY package.json package-lock.json ./
RUN npm install

# Install TailwindCSS
RUN npm install -g tailwindcss

# Check if Tailwind is installed
RUN npx tailwindcss -v || (echo "TailwindCSS is not installed!" && exit 1)

# Copy full source (needed for Tailwind's scanning)
COPY --from=build /app /app

# Generate TailwindCSS output
RUN npx tailwindcss -o ./public/static/css/tw.css --minify

# Stage 3: Build final Go binary
FROM golang:1.24-alpine AS final-build
WORKDIR /app

# Install necessary dependencies
RUN apk add --no-cache git

# Copy Go module files and dependencies (to reuse caching)
COPY --from=build /app/go.mod /app/go.sum ./
RUN go mod download

# Copy full source again for the final Go build
COPY --from=build /app /app

# Copy built Tailwind CSS
COPY --from=css /app/public/static/css/tw.css /app/public/static/css/tw.css

# Build the Go binary
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/main cmd/server/main.go

# Stage 4: Create minimal runtime container
FROM alpine:latest
WORKDIR /app

# Install CA certificates if needed
RUN apk add --no-cache ca-certificates

# Copy built binary
COPY --from=final-build /app/bin/main /app/main

# Ensure the binary is executable
RUN chmod +x /app/main

# Copy static assets and generated content
COPY --from=final-build /app/public /app/public
COPY --from=final-build /app/generated /app/generated
COPY --from=final-build /app/content /app/content

# Expose the application port
EXPOSE 8082

CMD ["/app/main"]
