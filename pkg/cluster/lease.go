package cluster

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"nmid-registry/pkg/loger"
	"strconv"
	"time"
)

const (
	LeaseFormat = "/leases/%s"         //+cluster member name
	LeaseTTL    = clientv3.MaxLeaseTTL // 9000000000Second=285Year
	MinTTL      = 5
)

func (c *cluster) GetLeaseKey() string {
	return fmt.Sprintf(LeaseFormat, c.options.Name)
}

func (c *cluster) GetLeaseStr() (string, error) {
	leaseKey := c.GetLeaseKey()
	return c.Get(leaseKey)
}

func (c *cluster) NewLease() error {
	lease, err := c.GetLeaseStr()
	if nil != err {
		return err
	}

	var leaseID *clientv3.LeaseID
	if len(lease) > 0 {
		leaseID, err = str2Lease(lease)
		if nil != err {
			return err
		}
	}

	client, err := c.GetClusterClient()
	if err != nil {
		return fmt.Errorf("lease get client err %v", err)
	}

	if leaseID != nil {
		resp, err := func() (*clientv3.LeaseTimeToLiveResponse, error) {
			ctx, cancel := c.RequestContext()
			defer cancel()

			return client.Lease.TimeToLive(ctx, *leaseID)
		}()
		if nil != err || resp.TTL < MinTTL {
			return c.GenNewLease()
		}

		c.lease = leaseID
		loger.Loger.Infof("lease is ready (use existed: %x)", *c.lease)

		return nil
	}

	return c.GenNewLease()
}

func (c *cluster) GenNewLease() error {
	c.leaseMutex.Lock()
	defer c.leaseMutex.Unlock()

	client, err := c.GetClusterClient()
	if err != nil {
		return fmt.Errorf("lease get client err %v", err)
	}

	grantResp, err := func() (*clientv3.LeaseGrantResponse, error) {
		ctx, cancel := c.RequestContext()
		defer cancel()
		return client.Lease.Grant(ctx, LeaseTTL)
	}()
	if err != nil {
		return err
	}

	_, err = func() (*clientv3.PutResponse, error) {
		ctx, cancel := c.RequestContext()
		defer cancel()
		return client.Put(ctx, c.GetLeaseKey(), fmt.Sprintf("%x", grantResp.ID), clientv3.WithLease(grantResp.ID))
	}()

	if err != nil {
		func() (*clientv3.LeaseRevokeResponse, error) {
			ctx, cancel := c.RequestContext()
			defer cancel()
			return client.Lease.Revoke(ctx, grantResp.ID)
		}()

		return fmt.Errorf("put lease to %s failed: %v", c.GetLeaseKey(), err)
	}

	lease := grantResp.ID
	c.lease = &lease

	loger.Loger.Infof("lease is ready (grant new one: %x)", *c.lease)

	return nil
}

func (c *cluster) KeepAliveLease() {
	for {
		select {
		case <-c.done:
			return
		case <-time.After(c.requestTimeout):
			client, err := c.GetClusterClient()
			if err != nil {
				loger.Loger.Errorf("get client failed: %v", err)
				continue
			}

			leaseID, err := c.GetLease()
			if err != nil {
				loger.Loger.Errorf("get lease failed: %v", err)
				err := c.GenNewLease()
				if err != nil {
					loger.Loger.Errorf("grant new lease failed: %v", err)
				}
				continue
			}

			// KeepAliveOnce renews the lease once. In most of the cases, KeepAlive
			_, err = func() (*clientv3.LeaseKeepAliveResponse, error) {
				ctx, cancel := c.RequestContext()
				defer cancel()
				return client.Lease.KeepAliveOnce(ctx, leaseID)
			}()
			if err != nil {
				loger.Loger.Errorf("keep alive for lease %x failed: %v", leaseID, err)
				err := c.GenNewLease()
				if err != nil {
					loger.Loger.Errorf("grant new lease failed: %v", err)
				}
				continue
			}
		}
	}
}

func (c *cluster) GetLease() (clientv3.LeaseID, error) {
	c.leaseMutex.RLock()
	defer c.leaseMutex.RUnlock()

	if c.lease == nil {
		return 0, fmt.Errorf("lease is not ready")
	}

	return *c.lease, nil
}

func str2Lease(leaseStr string) (*clientv3.LeaseID, error) {
	leaseNum, err := strconv.ParseInt(leaseStr, 16, 64)
	if err != nil {
		return nil, err
	}

	leaseID := clientv3.LeaseID(leaseNum)

	return &leaseID, nil
}
