package loki_server

type BasicAuth struct {
	Username string
	Password string
}

type ClientOpts struct {
	BasicAuth *BasicAuth
}
