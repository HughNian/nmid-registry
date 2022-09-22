package cluster

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"nmid-registry/pkg/loger"
)

type Watcher interface {
	Watch(key string) (chan<- WatchRet, error)
}

type watcher struct {
	w    clientv3.Watcher
	done chan struct{}
}

type WatchRet struct {
	WType mvccpb.Event_EventType
	WKey  string
}

func (c *cluster) NewWatcher() (Watcher, error) {
	client, err := c.GetClusterClient()
	if nil != err {
		return nil, err
	}

	w := clientv3.NewWatcher(client)

	return &watcher{
		w:    w,
		done: make(chan struct{}),
	}, nil
}

func (w *watcher) Watch(key string) (chan<- WatchRet, error) {
	//can't use context with timeout here
	ctx, cancel := context.WithCancel(context.Background())
	wResp := w.w.Watch(ctx, key)

	keyChan := make(chan<- WatchRet, 10)

	go func() {
		defer cancel()
		defer close(keyChan)

		for {
			select {
			case <-w.done:
				return
			case resp := <-wResp:
				if resp.Canceled {
					loger.Loger.Infof("watch key %s canceled: %v", key, resp.Err())
					return
				}
				if resp.IsProgressNotify() {
					continue
				}
				for _, event := range resp.Events {
					switch event.Type {
					case mvccpb.PUT:
						wRet := WatchRet{
							WType: mvccpb.PUT,
							WKey:  string(event.Kv.Value),
						}
						keyChan <- wRet
					case mvccpb.DELETE:
						wRet := WatchRet{
							WType: mvccpb.DELETE,
							WKey:  string(event.Kv.Value),
						}
						keyChan <- wRet
					default:
						loger.Loger.Errorf("key %s received unknown event type %v", key, event.Type)
					}
				}
			}
		}
	}()

	return keyChan, nil
}

func (w *watcher) Close() {
	close(w.done)

	err := w.w.Close()

	if err != nil {
		loger.Loger.Errorf("close watcher failed: %s", err.Error())
	}
}
