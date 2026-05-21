# api-server image

- Dockerfile: `build-sys/images/api-server/Dockerfile`
- Build context: repository root
- Output image: `ghcr.io/kloudlite/kloudlite/api-server:<tag>`
- Pre-build artifact: `bin/api-server` from `task api:build:api-server`
- Runtime: runs `/app/api-server server api`.
