module github.com/kloudlite/cli

go 1.23.0

toolchain go1.23.11

replace github.com/kloudlite/api => ../api

require (
	github.com/briandowns/spinner v1.23.2
	github.com/fatih/color v1.18.0
	github.com/kloudlite/api v1.1.2
	github.com/urfave/cli/v2 v2.27.7
	google.golang.org/grpc v1.74.2
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
