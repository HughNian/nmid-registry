package registry

import (
	"encoding/json"
	"nmid-registry/pkg/cluster"
	"sync"
)

type Registry struct {
	lock sync.Mutex

	cluster cluster.Cluster
}

func NewRegistry(cls cluster.Cluster) *Registry {
	return &Registry{
		cluster: cls,
	}
}

//Register a new service.
func (r *Registry) Register(arg *ArgRegister, ins *Instance) (err error) {
	insVal, err := json.Marshal(ins)
	if nil != err {
		return err
	}

	r.cluster.Put(arg.ServiceId, string(insVal))
	return
}
