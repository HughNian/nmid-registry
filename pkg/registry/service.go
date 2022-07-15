package registry

import "sync"

//instance status
const (
	InstanceOk = iota
	InstanceError
)

type Service struct {
	sync.RWMutex

	ServiceId string
	Zone      string
	Instances map[string]*Instance

	LatestTimestamp int64
}

type Instance struct {
	Region   string
	Zone     string
	Env      string
	AppID    string
	HostName string
	Addrs    []string
	Version  string
	Metadata map[string]string

	Status uint32

	//timestamp
	RegTimestamp   int64
	UpTimestamp    int64
	RenewTimestamp int64
	DirtyTimestamp int64

	LatestTimestamp int64
}
