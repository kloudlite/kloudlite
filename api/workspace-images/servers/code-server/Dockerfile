FROM kloudlite/workspace-base:latest

USER root

# Install code-server
RUN curl -fsSL https://code-server.dev/install.sh | sh

USER workspace

# Configure code-server
RUN mkdir -p /home/workspace/.config/code-server && \
    echo "bind-addr: 0.0.0.0:8080" > /home/workspace/.config/code-server/config.yaml && \
    echo "auth: none" >> /home/workspace/.config/code-server/config.yaml && \
    echo "cert: false" >> /home/workspace/.config/code-server/config.yaml

EXPOSE 8080

CMD ["/usr/bin/code-server", "--bind-addr", "0.0.0.0:8080", "--auth", "none", "/workspace"]
