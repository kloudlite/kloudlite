package constants

import "os"

var (
	LoginUrl = func() string {
		if os.Getenv("BASE_URL") == "" {
			return "https://auth.kloudlite.io/cli-login"
		}

		return os.Getenv("BASE_URL") + "/cli-login"
	}()
	ServerURL = func() string {
		if os.Getenv("BASE_URL") == "" {
			return "https://auth.devc.kloudlite.io/api/"
		}

		return os.Getenv("BASE_URL") + "/api/"
	}()

	UpdateURL = func() string {
		if os.Getenv("Update_URL") == "" {
			return "https://i.jpillora.com/kloudlite/kl"
		}

		return os.Getenv("Update_URL")
	}()
)
