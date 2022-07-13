package cluster

import (
	"fmt"
	"go.etcd.io/etcd/server/v3/embed"
	"nmid-registry/pkg/loger"
	"sync"
	"time"
)

func (c *cluster) StartClusterServer() (err error) {
	c.serverMutex.Lock()
	defer c.serverMutex.Unlock()

	if nil != c.server {
		return nil
	}

	var etcdConfig *embed.Config
	if c.options.IsUseInitialCluster() {
		etcdConfig, err = CreateEtcdConfig(c.options)
	} else {
		etcdConfig, err = CreateEtcdConfigAddMember(c.options, c.members)
	}
	if nil != err {
		return err
	}

	server, err := embed.StartEtcd(etcdConfig)
	if err != nil {
		return err
	}

	go c.EtcdServerHandle(server)

	return nil
}

func (c *cluster) EtcdServerHandle(server *embed.Etcd) {
	select {
	case <-server.Server.ReadyNotify():
		c.server = server
		if c.server.Config().IsNewCluster() {
			err := c.Put(NmClusterNameKey, c.options.ClusterName)
			if err != nil {
				err = fmt.Errorf("register cluster name %s failed: %v",
					c.options.ClusterName, err)
				loger.Loger.Errorf("%v", err)
				panic(err)
			}
		}

		go func(s *embed.Etcd) {
			select {
			case err, ok := <-s.Err():
				if ok {
					loger.Loger.Errorf("etcd server %s serve failed: %v",
						c.server.Config().Name, err)
					CloseClusterServer(s)
				}
			}
		}(server)

		loger.Loger.Infof("server is ready")
	case <-time.After(ServerTimeout):
		CloseClusterServer(server)
	}
}

func (c *cluster) GetClusterServer() (*embed.Etcd, error) {
	c.serverMutex.RLock()
	defer c.serverMutex.RUnlock()

	if c.server == nil {
		return nil, fmt.Errorf("server is not ready")
	}
	return c.server, nil
}

func (c *cluster) CloseServer() {
	c.serverMutex.Lock()
	defer c.serverMutex.Unlock()

	if c.server == nil {
		return
	}

	CloseClusterServer(c.server)
	c.server = nil
}

func (c *cluster) Close(wg *sync.WaitGroup) {
	defer wg.Done()

	close(c.done)
	c.CloseClusterSession()
	c.CloseClusterClient()
	c.CloseServer()
}

func CloseClusterServer(server *embed.Etcd) {
	select {
	case <-server.Server.ReadyNotify():
		server.Close()
		<-server.Server.StopNotify()
	default:
		server.Server.HardStop()

		for _, client := range server.Clients {
			if client != nil {
				client.Close()
			}
		}

		for _, peer := range server.Peers {
			if peer != nil {
				peer.Close()
			}
		}

		loger.Loger.Infof("hard stop server")
	}
}
