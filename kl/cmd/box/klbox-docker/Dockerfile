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
RUN cat > /entrypoint.sh <<'EOF'
#!/bin/bash
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
sudo -u kl bash -c "/home/kl/.nix-profile/bin/nix-shell -p $(cat /home/kl/.nix-shell-args) --run 'echo done' | tee /tmp/exec.log"
chown -R kl /home/kl/.nix-shell-args
chown -R kl /home/kl/.ssh

sudo /mounter --conf /tmp/sample.json

/usr/sbin/sshd -D
EOF
RUN chmod +x /entrypoint.sh
# Run SSHD on container start
# ENTRYPOINT ["bash", "/entrypoint.sh"]
ENTRYPOINT ["/entrypoint.sh"]