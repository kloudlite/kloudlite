package kafka

type Conn interface {
	GetBrokers() string
	GetSASLAuth() *SASLAuth
}

type conn struct {
	brokers  string
	saslAuth *SASLAuth
}

func (c *conn) GetSASLAuth() *SASLAuth {
	return c.saslAuth
}

func (c *conn) GetBrokers() string {
	return c.brokers
}

type ConnectOpts struct {
	SASLAuth *SASLAuth
}

func Connect(brokers string, opts ConnectOpts) (Conn, error) {
	return &conn{
		brokers:  brokers,
		saslAuth: opts.SASLAuth,
	}, nil
}
