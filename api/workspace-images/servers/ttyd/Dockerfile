FROM kloudlite/workspace-base:latest

USER root

# Install ttyd and dependencies
RUN apt-get update && apt-get install -y \
    cmake \
    libjson-c-dev \
    libwebsockets-dev \
    && rm -rf /var/lib/apt/lists/*

# Build and install ttyd from source
RUN cd /tmp && \
    git clone https://github.com/tsl0922/ttyd.git && \
    cd ttyd && mkdir build && cd build && \
    cmake .. && \
    make && make install && \
    cd / && rm -rf /tmp/ttyd

USER workspace

EXPOSE 7681

CMD ["ttyd", "--writable", "--port", "7681", "/bin/bash"]
