package boxpkg

func (c *client) Restart() error {
	defer c.spinner.Start("stopping container please wait")()

	cr, err := c.getContainer(map[string]string{
		CONT_MARK_KEY: "true",
	})
	if err != nil && err != notFoundErr {
		return err
	}

  _ = cr

	if err == notFoundErr {
		return c.Start()
	}

	if err := c.Stop(); err != nil {
		return err
	}

	return c.Start()
}
