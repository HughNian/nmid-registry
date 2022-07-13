package cluster

import "go.etcd.io/etcd/api/v3/mvccpb"

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

func (c *cluster) Get(key string) (string, error) {
	kv, err := c.GetRaw(key)
	if nil != err {
		return ``, nil
	}

	return string(kv.Value), nil
}

func (c *cluster) GetRaw(key string) (*mvccpb.KeyValue, error) {
	client, err := c.GetClusterClient()
	if nil != err {
		return nil, err
	}

	ctx, cancel := c.RequestContext()
	defer cancel()

	resp, err := client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0], nil
}
