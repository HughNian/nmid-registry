package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.etcd.io/etcd/server/v3/embed"
	yaml "gopkg.in/yaml.v2"
	"nmid-registry/pkg/loger"
	"nmid-registry/pkg/option"
	"strings"
	"sync"
	"time"
)

const (
	ServerTimeout = 10 * time.Minute

	HeartbeatTime = 5 * time.Second

	ClientDialTimeout      = 10 * time.Second
	ClientKeepAlive        = 1 * time.Minute
	ClientKeepAliveTimeout = 1 * time.Minute

	NmClusterNameKey = "/nm/cluster/name"
)

type (
	ClusterStatus struct {
		ID        string `yaml:"id"`
		State     string `yaml:"state"`
		StartTime string `yaml:"startTime"`
	}

	clusterStats struct {
		ID        string    `json:"id"`
		State     string    `json:"state"`
		StartTime time.Time `json:"startTime"`
	}
)

type Cluster interface {
	IsLeader() bool
	Put(key, value string) error
	PutUnderLease(key, value string) error
	Get(key string) (string, error)
	GetRaw(key string) (*mvccpb.KeyValue, error)
	CloseCluster(wg *sync.WaitGroup)
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
	if c.options.ClusterRole == "slave" {
		_, err := c.GetClusterClient()
		if err != nil {
			return err
		}

		err = c.CheckClusterName()
		if err != nil {
			return err
		}

		err = c.NewLease()
		if err != nil {
			return fmt.Errorf("new lease err %v", err)
		}

		go c.KeepAliveLease()

		return nil
	}

	err = c.StartClusterServer()
	if err != nil {
		return fmt.Errorf("start server failed: %v", err)
	}

	err = c.NewLease()
	if err != nil {
		return fmt.Errorf("new lease failed: %v", err)
	}

	go c.KeepAliveLease()

	return nil
}

func (c *cluster) ClusterRun() {
	var tryTimes int
	var tryReady = func() error {
		tryTimes++
		err := c.ClusterReady()
		return err
	}

	if err := tryReady(); nil != err {
		for {
			time.Sleep(HeartbeatTime)
			err := tryReady()
			if err != nil {
				loger.Loger.Errorf("failed start many times(%d)", tryTimes)
			} else {
				break
			}
		}
	}

	loger.Loger.Infof("cluster is ready")

	if c.options.ClusterRole == "master" {
		go c.DoDefrag()
	}

	go c.BackendHandle()
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

//BackendHandle 处理集群状态同步，更新成员信息
func (c *cluster) BackendHandle() {
	for {
		select {
		case <-time.After(HeartbeatTime):
			err := c.SyncStatus()
			if err != nil {
				loger.Loger.Errorf("sync status failed %v", err)
			}

			err = c.UpdateMembers()
			if err != nil {
				loger.Loger.Errorf("update members failed %v", err)
			}

		case <-c.done:
			return
		}
	}
}

//SyncStatus 同步状态
func (c *cluster) SyncStatus() error {
	status := MemberStatus{
		Options: *c.options,
	}

	if c.options.ClusterRole == "master" {
		server, err := c.GetClusterServer()
		if err != nil {
			return err
		}

		selfStats := server.Server.SelfStats()
		stats, err := newClusterStats(selfStats)
		if err != nil {
			return err
		}
		status.CStatus = stats.toClusterStatus()
	}

	status.LastHeartbeatTime = time.Now().Format(time.RFC3339)

	yamlVal, err := yaml.Marshal(status)
	if err != nil {
		return err
	}

	leaseKey := fmt.Sprintf(StatusMemberFormat, c.options.Name)
	err = c.PutUnderLease(leaseKey, string(yamlVal))
	if err != nil {
		return fmt.Errorf("put status failed: %v", err)
	}

	return nil
}

//UpdateMembers 更新成员信息
func (c *cluster) UpdateMembers() error {
	client, err := c.GetClusterClient()
	if err != nil {
		return err
	}

	resp, err := func() (*clientv3.MemberListResponse, error) {
		ctx, cancel := c.RequestContext()
		defer cancel()
		return client.MemberList(ctx)
	}()
	if err != nil {
		return err
	}

	if c.members != nil {
		c.members.UpdateClusterMembers(resp.Members)
	}

	return nil
}

func (c *cluster) RequestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.requestTimeout)
}

func (c *cluster) LongRequestContext() (context.Context, context.CancelFunc) {
	timeout := 3 * c.requestTimeout
	return context.WithTimeout(context.Background(), timeout)
}

func (c *cluster) IsLeader() bool {
	server, err := c.GetClusterServer()
	if err != nil {
		return false
	}

	return server.Server.Leader() == server.Server.ID()
}

func newClusterStats(buff []byte) (*clusterStats, error) {
	stats := clusterStats{}
	err := json.Unmarshal(buff, &stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (s *clusterStats) toClusterStatus() *ClusterStatus {
	return &ClusterStatus{
		ID:        s.ID,
		State:     strings.TrimPrefix(s.State, "State"),
		StartTime: s.StartTime.Format(time.RFC3339),
	}
}
