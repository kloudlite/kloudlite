# oci-installer image

- Dockerfile: `build-sys/images/oci-installer/Dockerfile`
- Build context: repository root
- Output image: `ghcr.io/kloudlite/kloudlite/oci-installer:<tag>`
- Pre-build artifact: `bin/oci-installer` from `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o ./bin/oci-installer ./cli/kli/oci-installer`
- Runtime: non-root OCI installer binary image.
