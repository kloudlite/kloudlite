# Use the official Ubuntu base image
FROM --platform=$TARGETPLATFORM golang:1.21 as builder

WORKDIR /app

COPY main.go main.go

RUN go mod init mounter.io

RUN go build -o mounter main.go

FROM --platform=$TARGETPLATFORM ubuntu:latest

RUN apt-get update && apt-get install -y openssh-server sudo curl xz-utils jq iproute2

RUN mkdir /var/run/sshd

RUN useradd -m kl && usermod -aG sudo kl && chsh -s /bin/bash kl && \
  echo "kl ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/kl

# Optionally change the SSH port and other settings
#RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
# SSH login fix. Otherwise user is kicked off after login
RUN sed -i 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' /etc/pam.d/sshd

EXPOSE 22

ENV HOSTNAME box
RUN mkdir -m 0755 /nix && chown kl /nix
USER kl
RUN curl -L https://nixos.org/nix/install | sh -s -- --no-daemon --no-modify-profile

USER root
RUN chmod -x /etc/update-motd.d/*
RUN cat > /etc/motd <<'EOF'
Welcome to KloudLite's Development Environment
EOF

COPY .bashrc .profile /tmp/

COPY --from=builder /app/mounter /mounter

SHELL ["/bin/bash", "-c"]

COPY ./start.sh /start.sh
COPY ./entrypoint.sh /entrypoint.sh

RUN chmod +x /start.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
