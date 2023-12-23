package constants

import "os"

var (
	CMD_NAME = "kl"
	LoginUrl = func() string {
		if os.Getenv("BASE_URL") == "" {
			return "https://auth.kloudlite.io/cli-login"
		}

		return os.Getenv("BASE_URL") + "/cli-login"
	}()
	SERVER_URL = func() string {
		if os.Getenv("BASE_URL") == "" {
			return "https://auth.kloudlite.io/api/"
		}

		return os.Getenv("BASE_URL") + "/api/"
	}()
)
