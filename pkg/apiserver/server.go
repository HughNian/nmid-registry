package apiserver

import (
	"context"
	"fmt"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
	"nmid-registry/pkg/cluster"
	"nmid-registry/pkg/loger"
	"nmid-registry/pkg/option"
	"sync"
	"time"
)

const (
	ApiPrefix = ""
	LockKey   = "/config/lock"
)

func AboutMe() string {
	return fmt.Sprintf(`Copyright© 2021 - %d Nmid-Registry(http://www.niansong.top). Nmid-Registry is Nmid Register Center All rights reserved.Apache License 2.0.`, time.Now().Year())
}

type (
	ApiServer struct {
		clock cluster.CMutex
		lock  sync.Mutex

		option *option.Options

		writeOnly bool

		server  *bm.Engine
		cluster cluster.Cluster
	}
)

func NewApiServer(opt *option.Options, cls cluster.Cluster) *ApiServer {
	apiServer := &ApiServer{
		option:    opt,
		writeOnly: opt.WriteOnly,
		cluster:   cls,
	}

	//http server
	httpServer := bm.DefaultServer(&bm.ServerConfig{
		Network: "http",
		Addr:    opt.ApiAddr,
	})
	apiServer.server = httpServer

	HttpRouter(apiServer)

	if err := apiServer.server.Start(); err != nil {
		loger.Loger.Errorf("http server error %v", err)
		panic(err)
	}
	loger.Loger.Infof("http server start Listening on: %s", opt.ApiAddr)

	return apiServer
}

func (as *ApiServer) CloseApiServer(wg *sync.WaitGroup) {
	defer wg.Done()

	err := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return as.server.Shutdown(ctx)
	}()
	if nil != err {
		loger.Loger.Errorf("http server shutdown failed %v", err)
	}
}

func (as *ApiServer) IsWriteOnly() bool {
	return as.writeOnly
}
