package boxpkg

func (c *client) Stop() error {
	return c.stopContainer(c.cwd)
}
