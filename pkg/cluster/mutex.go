package cluster

import (
	"context"
	"go.etcd.io/etcd/client/v3/concurrency"
	"sync"
	"time"
)

//cluster level mutex.

type CMutex interface {
	Lock() error
	Unlock() error
}

type cmutex struct {
	m       sync.Mutex
	cm      *concurrency.Mutex
	timeout time.Duration
}

func (c *cluster) NewCMutex(key string) (CMutex, error) {
	session, err := c.NewClusterSession()
	if nil != err {
		return nil, err
	}

	return &cmutex{
		cm:      concurrency.NewMutex(session, key),
		timeout: c.requestTimeout,
	}, nil
}

func (cmt *cmutex) Lock() (err error) {
	cmError := true

	cmt.m.Lock()
	defer func() {
		if cmError || err != nil {
			cmt.m.Unlock()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), cmt.timeout)
	defer cancel()

	err = cmt.cm.Lock(ctx)
	cmError = false

	return
}

func (cmt *cmutex) Unlock() error {
	ctx, cancel := context.WithTimeout(context.Background(), cmt.timeout)
	defer cancel()
	defer cmt.m.Unlock()

	return cmt.cm.Unlock(ctx)
}
