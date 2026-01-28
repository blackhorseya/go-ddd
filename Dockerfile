# Dockerfile for GoReleaser builds (dockers_v2)
# This Dockerfile is designed to be used with GoReleaser, which handles the build process
# and copies the pre-built binary into the container.
#
# GoReleaser dockers_v2 places binaries in: linux/<arch>/<binary>
# Using TARGETPLATFORM ARG to select the correct binary.

FROM alpine:3.20

# Install ca-certificates for HTTPS connections and tzdata for timezone support
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Build arguments for multi-platform support
ARG TARGETPLATFORM

# Copy the pre-built binary from GoReleaser's build context
# GoReleaser places binaries in linux/<arch>/ directories
COPY ${TARGETPLATFORM}/service /app/service

# Set ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose HTTP and gRPC ports
EXPOSE 8080 9090

# Run the service
ENTRYPOINT ["/app/service"]
