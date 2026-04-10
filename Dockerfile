# Minimal Dockerfile for pre-compiled Go binary
FROM alpine:3.23

# Install CA certificates for HTTPS requests if needed
# RUN apk --no-cache add ca-certificates

# Always connect as a regular user to the container and not have root previlage
RUN adduser --disabled-password --gecos "" containeruser

WORKDIR /app

# Create a data dir and give ownership to containeruser
RUN mkdir data && chown containeruser:containeruser data

# Copy the pre-built binary and database from the host
# This assumes 'make docker-build' and 'make db-seed' were run first
COPY main .
COPY .env.example .env
# WARN: Critical: sqlite needs write access to the dir to create .wal and .shm files.
COPY --chown=containeruser:containeruser local_lms.db ./data/

# Expose port 8080
EXPOSE 8080

# Switch to none root user
USER containeruser

# HealthChecks: Use wget since it comes pre-installed on Alpine
# wget -qO- silently downloads the page and outputs to stdout. 
# If it gets a 500/503 from your Go app (because DB is down), wget exits with an error code (1), failing the healthcheck.
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthcheck || exit 1

# Command to run
CMD ["./main"]
