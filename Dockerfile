# Build stage
FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY internal/ internal/
COPY ["cmd/", "cmd/"]

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service ./cmd/auth-service

FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/auth-service .

ENV MODE=production
ENV PORT=8080

ENV DATABASE_HOST=
ENV DATABASE_PORT=
ENV DATABASE_USER=
ENV DATABASE_PASSWORD=
ENV DATABASE_NAME=
ENV ACCESS_TOKEN_SECRET=
ENV REFRESH_TOKEN_SECRET=
ENV ACCESS_TOKEN_TTL=
ENV REFRESH_TOKEN_TTL=
ENV VALIDATION_API_KEY=
ENV REDIS_HOST=
ENV REDIS_PORT=
ENV REDIS_PASSWORD=

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./auth-service"]

