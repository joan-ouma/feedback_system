# Final stage - FIXED: Remove postgresql-client
FROM alpine:latest

RUN apk --no-cache add ca-certificates
# Removed: postgresql-client (not needed for MongoDB)

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/bin/server .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Copy static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

EXPOSE 8080

CMD ["./server"]
