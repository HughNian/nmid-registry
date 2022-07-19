package registry

import (
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
	ServiceId  string
	InFlowUrl  string
	OutFlowUrl string
	Instances  map[string]*Instance

	LatestTimestamp int64
}

type Instance struct {
	Region      string
	Zone        string
	Env         string
	ServiceId   string
	ServiceName string
	HostName    string
	Addrs       []string
	Version     string
	Metadata    map[string]string

	Status uint32

	//timestamp
	RegTimestamp   int64
	UpTimestamp    int64
	RenewTimestamp int64
	DirtyTimestamp int64

	LatestTimestamp int64
}

func NewService(arg *ArgRegister) *Service {
	return &Service{}
}

func NewInstance(arg *ArgRegister) *Instance {
	return &Instance{}
}
