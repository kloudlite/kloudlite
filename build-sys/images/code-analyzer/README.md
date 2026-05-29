# code-analyzer image

- Dockerfile: `build-sys/images/code-analyzer/Dockerfile`
- Build context: repository root
- Output image: `ghcr.io/kloudlite/kloudlite/code-analyzer:<tag>`
- Pre-build artifact: `bin/code-analyzer` from `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o ./bin/code-analyzer ./cli/code-analyzer`
- Runtime: Semgrep-backed code analyzer service on port `8082`.
