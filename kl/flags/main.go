package flags

var (
	Version    = "development"
	CliName    = "kl"
	BasePrefix = ""
	DevMode    = "false"
)

func IsDev() bool {
	if DevMode == "false" {
		return false
	}
	return true
}
