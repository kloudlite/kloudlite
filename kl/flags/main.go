package flags

var (
	Version = "development"
	CliName = "kl"
	DevMode = "false"

	IsVerbose = false
	IsQuiet   = false

	ImageBase      = "ghcr.io/kloudlite/kl"
	DefaultBaseURL = "https://auth.kloudlite.io"
)

func IsDev() bool {
	if DevMode == "false" {
		return false
	}
	return true
}
