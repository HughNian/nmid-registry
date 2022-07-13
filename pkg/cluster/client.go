package cluster

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"nmid-registry/pkg/loger"
	"time"
)

const (
	AutoSyncTime         = 1 * time.Minute
	DialTimeout          = 10 * time.Second
	DialKeepAliveTime    = 1 * time.Minute
	DialKeepAliveTimeout = 1 * time.Minute

	ClientLogFileName = "etcd_client.log"
)

func (c *cluster) GetClusterClient() (client *clientv3.Client, err error) {
	c.clientMutex.Lock()
	if nil != c.client {
		client = c.client
		c.clientMutex.Unlock()
		return client, nil
	}
	c.clientMutex.Unlock()

	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()

	var endpoints []string
	if nil == c.members {
		endpoints = c.options.GetPeerUrls()
	} else {
		endpoints = c.members.KnownPeerUrls()
	}
	loger.Loger.Infof("client connect with endpoints: %v", endpoints)

	client, err = clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		AutoSyncInterval:     AutoSyncTime,
		DialTimeout:          DialTimeout,
		DialKeepAliveTime:    DialKeepAliveTime,
		DialKeepAliveTimeout: DialKeepAliveTimeout,
		LogConfig:            ClientLoggerConfig(c.options, ClientLogFileName),
		MaxCallSendMsgSize:   c.options.Cluster.MaxCallSendMsgSize,
	})
	if nil != err {
		return nil, err
	}

	c.client = client

	return client, nil
}

func (c *cluster) CloseClusterClient() {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()

	if nil == c.client {
		return
	}

	err := c.client.Close() //client close also with lease close
	if nil != err {
		loger.Loger.Errorf("close client err: %v", err)
	}

	c.client = nil
}
