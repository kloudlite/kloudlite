FROM amazon/aws-cli:latest

# Install gzip for backup compression
RUN yum install -y gzip || microdnf install -y gzip

# Set minimal labels
LABEL org.opencontainers.image.source="https://github.com/kloudlite/kloudlite"
LABEL org.opencontainers.image.description="K3s backup image with AWS CLI and gzip"

# Use the default entrypoint from aws-cli image
