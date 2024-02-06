package constants

import (
	"fmt"
	"os"

	"github.com/kloudlite/kl/flags"
)

var (
	prefix = flags.BasePrefix

	LoginUrl = func() string {
		if os.Getenv("BASE_URL") == "" {
			return fmt.Sprint("https://auth.", prefix, "kloudlite.io/cli-login")
		}

		return os.Getenv("BASE_URL") + "/cli-login"
	}()

	BaseURL = func() string {
		if os.Getenv("BASE_URL") == "" {
			return fmt.Sprint("https://auth.", prefix, "kloudlite.io")
		}

		return os.Getenv("BASE_URL") + "/api/"
	}()

	ServerURL = func() string {
		if os.Getenv("BASE_URL") == "" {
			return fmt.Sprint("https://auth.", prefix, "kloudlite.io/api/")
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

var (
	InfraCliName = "kli"
	CoreCliName  = "kl"
)
