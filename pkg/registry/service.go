package registry

import (
	"encoding/json"
	"time"
)

//instance status
const (
	InstanceOk = iota
	InstanceError
)

var (
	now = time.Now().UnixNano()
)

type Service struct {
	ServiceId   string
	InFlowAddr  string
	OutFlowAddr string
	Instances   []*Instance

	LatestTimestamp int64
}

type Instance struct {
	ServiceId string
	Region    string
	Zone      string
	Env       string
	HostName  string
	Addrs     []string
	Version   string
	Metadata  map[string]string

	Status uint32

	//timestamp
	RegTimestamp   int64
	UpTimestamp    int64
	RenewTimestamp int64
	DirtyTimestamp int64

	LatestTimestamp int64
}

func NewService(arg *ArgRegister) *Service {
	return &Service{
		ServiceId:       arg.ServiceId,
		InFlowAddr:      arg.InFlowAddr,
		OutFlowAddr:     arg.OutFlowAddr,
		Instances:       make([]*Instance, 0),
		LatestTimestamp: now,
	}
}

func NewInstance(arg *ArgRegister) *Instance {
	ins := &Instance{
		ServiceId:       arg.ServiceId,
		Region:          arg.Region,
		Zone:            arg.Zone,
		Env:             arg.Env,
		HostName:        arg.Hostname,
		Addrs:           arg.Addrs,
		Version:         arg.Version,
		Status:          arg.Status,
		RegTimestamp:    now,
		UpTimestamp:     now,
		RenewTimestamp:  now,
		DirtyTimestamp:  now,
		LatestTimestamp: now,
	}

	metaData := make(map[string]string)
	if err := json.Unmarshal([]byte(arg.Metadata), &metaData); err == nil {
		ins.Metadata = metaData
	}

	return ins
}
