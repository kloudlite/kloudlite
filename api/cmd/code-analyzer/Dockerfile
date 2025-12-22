FROM alpine:3.19

RUN apk --no-cache add ca-certificates curl nodejs npm

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

WORKDIR /

COPY ./bin/code-analyzer /usr/local/bin/code-analyzer

RUN chmod +x /usr/local/bin/code-analyzer

# Create directories for reports
RUN mkdir -p /var/lib/kloudlite/code-analysis

EXPOSE 8082

ENTRYPOINT ["/usr/local/bin/code-analyzer"]
