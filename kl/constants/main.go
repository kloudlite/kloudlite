package constants

import "os"

var (
	CmdName  = "kl"
	LoginUrl = func() string {
		if os.Getenv("BASE_URL") == "" {
			return "https://auth.kloudlite.io/cli-login"
		}

		return os.Getenv("BASE_URL") + "/cli-login"
	}()
	ServerURL = func() string {
		if os.Getenv("BASE_URL") == "" {
			return "https://auth.kloudlite.io/api/"
		}

		return os.Getenv("BASE_URL") + "/api/"
	}()
)
