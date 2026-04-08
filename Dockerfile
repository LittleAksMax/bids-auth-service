# Build stage
FROM --platform=$BUILDPLATFORM golang:alpine AS builder

ARG TARGETARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY internal/ internal/
COPY ["cmd/", "cmd/"]
COPY migrations/ migrations/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -installsuffix cgo -o auth-service ./cmd/auth-service

FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/auth-service .
COPY --from=builder /app/migrations ./migrations

ENV MODE=production
ENV PORT=8080

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./auth-service"]
