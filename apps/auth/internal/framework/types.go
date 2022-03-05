package framework

type Framework struct {
	start func() error
}

type Config struct {
	MongoDB struct {
		Uri string
		Db  string
	}
}
