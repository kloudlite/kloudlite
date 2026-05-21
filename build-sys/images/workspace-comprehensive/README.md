# workspace-comprehensive image

- Dockerfile: `build-sys/images/workspace-comprehensive/Dockerfile`
- Build context: repository root
- Output image: `ghcr.io/kloudlite/kloudlite/workspace-comprehensive:<tag>`
- Pre-build artifact: `bin/kl-linux` from `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -trimpath -o bin/kl-linux cli/kl/main.go`
- Runtime: full developer workspace image built from `workspace-base`.
