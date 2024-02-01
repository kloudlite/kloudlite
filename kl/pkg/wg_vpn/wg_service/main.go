package wg_svc

const (
	serviceUrl string = "http://localhost:17171"
)

func StartVpn(configuration []byte) error {
	return sendCommand("connect", string(configuration))
}

func StopVpn(verbose bool) error {
	return sendCommand("disconnect", "")
}

func EnsureInstalled() error {
	return ensureInstalled()
}

func EnsureAppRunning() error {
	return ensureAppRunning()
}
