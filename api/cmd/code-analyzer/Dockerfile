FROM python:3.11-alpine

# Install Semgrep and required dependencies
RUN apk --no-cache add ca-certificates git && \
    pip install --no-cache-dir semgrep

WORKDIR /

# Copy the binary
COPY ./bin/code-analyzer /usr/local/bin/code-analyzer

RUN chmod +x /usr/local/bin/code-analyzer

# Create directories for reports
RUN mkdir -p /var/lib/kloudlite/code-analysis

EXPOSE 8082

ENTRYPOINT ["/usr/local/bin/code-analyzer"]
