package cluster

import "go.etcd.io/etcd/client/v3/concurrency"

func (c *cluster) GetClusterSession() (session *concurrency.Session, err error) {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()

	return nil, nil
}
