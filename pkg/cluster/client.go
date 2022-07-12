package cluster

import clientv3 "go.etcd.io/etcd/client/v3"

func (c *cluster) GetClusterClient() (client *clientv3.Client, err error) {
	c.clientMutex.Lock()
	c.clientMutex.Unlock()

	return nil, nil
}
