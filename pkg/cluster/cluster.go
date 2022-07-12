package cluster

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/server/v3/embed"
	"nmid-registry/pkg/option"
	"sync"
	"time"
)

const (
	HeartbeatTime = 5 * time.Second
	ServerTimeout = 10 * time.Minute
	LeaseTTL      = clientv3.MaxLeaseTTL // 9000000000Second=285Year
	TTLTimeout    = 5

	ClientDialTimeout      = 10 * time.Second
	ClientKeepAlive        = 1 * time.Minute
	ClientKeepAliveTimeout = 1 * time.Minute
)

const (
	LeaseFormat         = "/leases/%s" //+memberName
	StatusMemberPrefix  = "/status/members/"
	StatusMemberFormat  = "/status/members/%s" // +memberName
	NmidrClusterNameKey = "/nmidr/cluster/name"
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
}

func NewCluster(opt *option.Options) (Cluster, error) {
	var members *Members
	var err error

	// defensive programming
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
	}

	return clu, nil
}

func (c *cluster) ClusterReady() (err error) {

	return nil
}

func (c *cluster) RequestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.requestTimeout)
}

func (c *cluster) IsLeader() bool {
	return true
}
