package cluster

import (
	"fmt"
	"go.etcd.io/etcd/client/v3/concurrency"
	"nmid-registry/pkg/loger"
)

func (c *cluster) NewClusterSession() (session *concurrency.Session, err error) {
	c.sessionMutex.RLock()
	if c.session != nil {
		session := c.session
		c.sessionMutex.RUnlock()
		return session, nil
	}
	c.sessionMutex.RUnlock()

	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()

	client, err := c.GetClusterClient()
	if err != nil {
		loger.Loger.Warnf("get cluster client err %v", err)
		return nil, err
	}

	lease, err := c.GetLease()
	if err != nil {
		loger.Loger.Warnf("get cluster lease err %v", err)
		return nil, err
	}

	session, err = concurrency.NewSession(client, concurrency.WithLease(lease))
	if err != nil {
		loger.Loger.Warnf("new client session err %v", err)
		return nil, fmt.Errorf("create session failed: %v", err)
	}

	c.session = session
	loger.Loger.Infof("session is ready")

	return session, nil
}

func (c *cluster) CloseClusterSession() {
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()

	if nil == c.session {
		return
	}

	err := c.session.Close()
	if nil != err {
		loger.Loger.Errorf("close cluster session err %v", err)
	}

	c.session = nil
}
