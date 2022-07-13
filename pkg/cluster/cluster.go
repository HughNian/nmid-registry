package cluster

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/server/v3/embed"
	"nmid-registry/pkg/loger"
	"nmid-registry/pkg/option"
	"sync"
	"time"
)

const (
	ServerTimeout = 10 * time.Minute

	HeartbeatTime    = 5 * time.Second
	DefragNormalTime = 1 * time.Hour
	DefragFailedTime = 1 * time.Minute

	ClientDialTimeout      = 10 * time.Second
	ClientKeepAlive        = 1 * time.Minute
	ClientKeepAliveTimeout = 1 * time.Minute
)

const (
	StatusMemberPrefix = "/status/members/"
	StatusMemberFormat = "/status/members/%s" // +memberName
	NmClusterNameKey   = "/nm/cluster/name"
)

type Cluster interface {
	IsLeader() bool
}

type cluster struct {
	options        *option.Options
	requestTimeout time.Duration

	serverMutex  sync.RWMutex
	clientMutex  sync.RWMutex
	leaseMutex   sync.RWMutex
	sessionMutex sync.RWMutex

	server  *embed.Etcd
	client  *clientv3.Client
	lease   *clientv3.LeaseID
	session *concurrency.Session
	members *Members

	done chan struct{}
}

func NewCluster(opt *option.Options) (Cluster, error) {
	var members *Members
	var err error

	requestTimeout, err := time.ParseDuration(opt.ClusterRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid cluster request timeout: %v", err)
	}

	if len(opt.GetPeerUrls()) == 0 {
		members, err = NewMembers(opt)
		if err != nil {
			return nil, fmt.Errorf("new members failed: %v", err)
		}
	}

	clu := &cluster{
		options:        opt,
		members:        members,
		requestTimeout: requestTimeout,
		done:           make(chan struct{}),
	}

	return clu, nil
}

func (c *cluster) ClusterReady() (err error) {

	return nil
}

func (c *cluster) ClusterRun() {

}

func (c *cluster) CheckClusterName() error {
	value, err := c.Get(NmClusterNameKey)
	if nil != err {
		return fmt.Errorf("check cluster name err %v", err)
	}

	if len(value) > 0 {
		if c.options.ClusterName != value {
			loger.Loger.Errorf("clustername check mismatch local(%s) != exist(%s)", c.options.ClusterName, value)
			panic(err)
		}
	} else {
		return fmt.Errorf("key %s not found", NmClusterNameKey)
	}

	return nil
}

func (c *cluster) RequestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.requestTimeout)
}

func (c *cluster) IsLeader() bool {
	server, err := c.GetClusterServer()
	if err != nil {
		return false
	}

	return server.Server.Leader() == server.Server.ID()
}
