# Dockerfile for GoReleaser builds (dockers_v2). Same image template for
# every binary in this repo (`service` and `migrate`). GoReleaser passes
# BINARY via --build-arg per dockers_v2 entry; the binary is copied under a
# fixed name inside the container so ENTRYPOINT can stay in exec form
# (which preserves signal forwarding for graceful shutdown).
#
# GoReleaser dockers_v2 places binaries in: linux/<arch>/<binary>
# Using TARGETPLATFORM ARG to select the correct platform.

FROM alpine:3.24

# Install ca-certificates for HTTPS connections and tzdata for timezone support
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Build arguments
ARG TARGETPLATFORM
ARG BINARY=service

# Copy the pre-built binary from GoReleaser's build context to a fixed
# name so ENTRYPOINT can use exec form (no shell = signals work).
COPY ${TARGETPLATFORM}/${BINARY} /app/entrypoint

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Run the binary
ENTRYPOINT ["/app/entrypoint"]
