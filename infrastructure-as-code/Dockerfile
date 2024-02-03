#syntax=docker/dockerfile:1.4
FROM alpine:3.16
RUN apk add bash curl gettext zip jq
# RUN apk add terraform helm kubectl --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community
RUN apk add helm kubectl --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community
RUN curl -L0 https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip > tf.zip && unzip tf.zip && mv terraform /usr/local/bin && rm tf.zip
RUN adduser --disabled-password --home="/app" --uid 1717 nonroot
USER nonroot
WORKDIR /app
COPY --chown=nonroot ./terraform ./terraform
RUN mkdir -p infrastructure-templates
COPY --chown=nonroot ./infrastructure-templates ./infrastructure-templates
ENV TF_PLUGIN_CACHE_DIR="/app/.terraform.d/plugin-cache"
RUN mkdir -p $TF_PLUGIN_CACHE_DIR
# SHELL ["/bin/bash", "-c"]
# RUN cat > script.sh <<'EOF'
#   for dir in $(ls -d ./infrastructures/templates/*); do
#     pushd $dir
#     terraform init -backend=false
#     popd
#   done
#   tdir=$(basename $(dirname $TF_PLUGIN_CACHE_DIR))
#   zip terraform.zip -r $tdir && rm -rf $tdir
# EOF
RUN mkdir build-scripts
COPY .ci/scripts/terraform-module-cache.sh ./build-scripts/terraform-module-cache.sh
RUN bash build-scripts/terraform-module-cache.sh ./infrastructure-templates
ENV TERRAFORM_ZIPFILE="/app/terraform.zip"
ENV TEMPLATES_DIR="/app/infrastructure-templates"
