package cluster

func (c *cluster) Put(key, value string) error {
	client, err := c.GetClusterClient()
	if err != nil {
		return err
	}

	ctx, cancel := c.RequestContext()
	defer cancel()
	_, err = client.Put(ctx, key, value)
	return err
}
