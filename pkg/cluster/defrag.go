package cluster

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"nmid-registry/pkg/loger"
	"time"
)

//客户端存储碎片处理
const (
	DefragNormalTime = 1 * time.Hour
	DefragFailedTime = 1 * time.Minute
)

func (c *cluster) DoDefrag() {
	defragtime := DefragNormalTime
	for {
		select {
		case <-time.After(defragtime):
			defragtime = c.RunDefrag()
		case <-c.done:
			return
		}
	}
}

func (c *cluster) RunDefrag() time.Duration {
	client, err := c.GetClusterClient()
	if err != nil {
		loger.Loger.Errorf("defrag failed: get client failed: %v", err)
		return DefragFailedTime
	}

	defragmentURL, err := c.options.GetFirstAdvertiseClientURL()
	if err != nil {
		loger.Loger.Errorf("defrag err %v", err)
		return DefragNormalTime
	}
	_, err = func() (*clientv3.DefragmentResponse, error) {
		ctx, cancel := c.LongRequestContext()
		defer cancel()
		return client.Defragment(ctx, defragmentURL)
	}()
	if err != nil {
		loger.Loger.Errorf("defrag failed %v", err)
		return DefragFailedTime
	}

	loger.Loger.Infof("defrag successfully")
	return DefragNormalTime
}
