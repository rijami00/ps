# Stage 1: Build Go binary
FROM golang:1.24-alpine AS generate
WORKDIR /app

# Install necessary dependencies
# RUN apk add --no-cache git

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@v0.3.898

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source code
COPY . .

# Run code generation
RUN go run cmd/generate/main.go && templ generate

# Stage 2: Process TailwindCSS (Full Source)
FROM node:22-alpine AS css
WORKDIR /app

# Copy package.json and install dependencies
COPY package.json ./
RUN npm install

# Install TailwindCSS
RUN npm install -g tailwindcss

# Check if Tailwind is installed
# RUN npx tailwindcss --help > /dev/null || (echo "TailwindCSS is not installed!" && exit 1)

# Copy full source (needed for Tailwind's scanning)
COPY --from=generate /app /app

# Generate TailwindCSS output
RUN npx tailwindcss -i input.css -o ./public/static/css/tw.css --minify

# Stage 3: Build final Go binary
FROM golang:1.24-alpine AS build
WORKDIR /app

# Install necessary dependencies
# RUN apk add --no-cache git

# Copy Go module files and dependencies (to reuse caching)
COPY --from=generate /app/go.mod /app/go.sum ./
RUN go mod download


# Copy full source again for the final Go build
COPY --from=generate /app /app

# Copy built Tailwind CSS
COPY --from=css /app/public/static/css/tw.css /app/public/static/css/tw.css

# Build the Go binary
RUN GOOS=linux GOARCH=amd64 go build -tags zoneinfodata -ldflags "-s -w" -o bin/main cmd/server/main.go

# Stage 4: Create minimal runtime container
FROM alpine:latest
WORKDIR /app

# Copy built binary
COPY --from=build /app/bin/main /app/main

# Ensure the binary is executable
RUN chmod +x /app/main

# Copy static assets and generated content
COPY --from=build /app/public /app/public
COPY --from=build /app/generated /app/generated
COPY --from=build /app/content /app/content
COPY --from=build /app/description.json /app/description.json
# Expose the application port
EXPOSE 8080

CMD ["/app/main"]
