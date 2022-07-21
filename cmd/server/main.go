package server

import (
	"nmid-registry/pkg/apiserver"
	"nmid-registry/pkg/cluster"
	"nmid-registry/pkg/envdir"
	"nmid-registry/pkg/loger"
	"nmid-registry/pkg/option"
	"nmid-registry/pkg/utils"
	"os"
	"sync"
)

func main() {
	//init option
	opt := option.New()
	msg, err := opt.Parse()
	if nil != err {
		utils.Exit(1, err.Error())
	}
	if msg != "" {
		utils.Exit(0, msg)
	}

	//init env dir
	err = envdir.InitEnvDir(opt)
	if nil != err {
		loger.Loger.Printf("failed to init env dir %v", err)
		utils.Exit(1, err.Error())
	}

	//new cluster
	cls, err := cluster.NewCluster(opt)
	if nil != err {
		loger.Loger.Errorf("new cluster failed %v", err)
		utils.Exit(1, err.Error())
	}

	//new api server
	apis, err := apiserver.NewApiServer(opt, cls)
	if nil != err {
		loger.Loger.Errorf("new cluster failed %v", err)
		utils.Exit(1, err.Error())
	}

	//close nmid-registry by signal
	sigChan := make(chan utils.Signal, 1)
	if err := utils.NotifySignal(sigChan, utils.SignalInt, utils.SignalTerm); err != nil {
		loger.Loger.Printf("failed to register signal %v", err)
		os.Exit(1)
	}
	sig := <-sigChan
	go func() {
		sig := <-sigChan
		loger.Loger.Infof("%s signal received, closing nmid-registry immediately", sig)
		os.Exit(255)
	}()
	loger.Loger.Infof("%s signal received, closing nmid-registry", sig)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	apis.CloseApiServer(wg)
	cls.CloseCluster(wg)
	wg.Wait()
}
