# Use the official Ubuntu base image
FROM golang:1.21 as builder

WORKDIR /app

COPY main.go main.go

RUN go build -o mounter main.go

FROM ubuntu:latest

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
RUN cat > /start.sh <<'EOF'
#!/bin/bash

set -o errexit
set -o pipefail

trap "echo kloudlite-entrypoint:CRASHED >&2" EXIT SIGINT SIGTERM

entrypoint_executed="/home/kl/.kloudlite_entrypoint_executed"
if [ ! -f "$entrypoint_executed" ]; then
    cp /tmp/.bashrc /home/kl/.bashrc
    cp /tmp/.profile /home/kl/.profile
    echo "successfully initialized .profile and .bashrc" >> $entrypoint_executed
fi

shift
# Read the JSON file into a variable
echo $@ | jq -r | jq -r '.packages | join(" ")'> /home/kl/.nix-shell-args
echo "$@" | jq -r > /tmp/sample.json
chown -R kl /tmp/sample.json

echo $@ | jq -r | jq -r '.envVars | map_values(. = "export \(.key)=\(.value)") | .[]' > /home/kl/.env-vars
chown -R kl /home/kl/.env-vars
sudo -u kl bash -c "/home/kl/.nix-profile/bin/nix-shell -p $(cat /home/kl/.nix-shell-args) --run 'echo done'"
chown -R kl /home/kl/.nix-shell-args

if [ -d "/tmp/ssh2" ]; then
    mkdir -p /home/kl/.ssh
    chown -R kl /home/kl/.ssh
    cp /tmp/ssh2/authorized_keys /home/kl/.ssh/authorized_keys
    chmod 600 /home/kl/.ssh/authorized_keys
    chown kl /home/kl/.ssh/authorized_keys
    echo "successfully copied ssh credentials"
fi 

sudo /mounter --conf /tmp/sample.json

trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint: SETUP_COMPLETE"
/usr/sbin/sshd -D
EOF

RUN cat > /entrypoint.sh <<'EOF'
#! /usr/bin/env bash
if [ -f "/tmp/stderr.log" ]; then
    sudo chown root /tmp/stderr.log
fi

if [ -f "/tmp/stdout.log" ]; then
    sudo chown root /tmp/stdout.log
fi

(tail -f /tmp/stdout.log) &
pid=$!
(tail -f /tmp/stderr.log) &
pid="$pid $1"

trap "eval kill -9 $pid" EXIT TERM
/start.sh $@ > /tmp/stdout.log 2> /tmp/stderr.log
EOF

RUN chmod +x /start.sh
RUN chmod +x /entrypoint.sh


ENTRYPOINT ["/entrypoint.sh"]
