#syntax=docker/dockerfile:1
FROM alpine:3.16
RUN apk add bash curl gettext zip
RUN apk add terraform --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community
RUN adduser --disabled-password --home="/app" --uid 1717 nonroot
# RUN chown -R nonroot:nonroot /app
USER nonroot
WORKDIR /app
COPY --chown=nonroot ./terraform ./terraform
RUN mkdir infrastructures
COPY --chown=nonroot ./infrastructures ./infrastructures
ENV TF_PLUGIN_CACHE_DIR="/app/.terraform.d/plugin-cache"
RUN mkdir -p $TF_PLUGIN_CACHE_DIR
RUN cat > script.sh <<'EOF'
for dir in $(ls -d ./infrastructures/templates/*); do
  pushd $dir
  terraform init -backend=false
  popd
done
tdir=$(basename $(dirname $TF_PLUGIN_CACHE_DIR))
zip terraform.zip -r $tdir && rm -rf $tdir
EOF
RUN bash script.sh
