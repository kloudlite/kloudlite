# nix image

- Dockerfile: `build-sys/images/nix/Dockerfile`
- Build context: repository root
- Output image: `ghcr.io/kloudlite/kloudlite/nix:<tag>`
- Pre-build artifact: none
- Runtime: helper image containing a single-user Nix install for seeding shared Nix volumes.
