# workmachine-manager image

- Dockerfile: `build-sys/images/workmachine-manager/Dockerfile`
- Build context: repository root
- Output image: `ghcr.io/kloudlite/kloudlite/workmachine-manager:<tag>`
- Pre-build artifact: `bin/workmachine-manager` from `task api:build:workmachine-manager`
- Runtime: runs `/app/workmachine-manager server workmachine-manager`.
