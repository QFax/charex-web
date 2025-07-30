# ---- Build Stage ----
# Use a specific version of the golang image for reproducibility.
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container.
WORKDIR /app

# Copy the Go module files and download dependencies. This is done
# as a separate step to leverage Docker's layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application's source code.
COPY . .

# Build the Go application.
# - CGO_ENABLED=0 disables Cgo to create a static binary.
# - GOOS=linux specifies the target operating system.
# - -a forces rebuilding of packages that are already up-to-date.
# - -installsuffix cgo is used to avoid reusing cached packages.
# - -o specifies the output file name.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/charex-web cmd/charex-web/main.go

# ---- Final Stage ----
# Use a minimal, "distroless" image for the final container.
# It contains only the application and its runtime dependencies.
FROM gcr.io/distroless/static-debian11

# Copy the compiled binary from the builder stage.
COPY --from=builder /app/charex-web /charex-web

# Copy the static assets from the builder stage.
COPY --from=builder /app/web/static /web/static

# Set the command to run when the container starts.
CMD ["/charex-web"]