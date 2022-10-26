package registry

type ArgRegister struct {
	ServiceId       string   `form:"service_id" binding:"required"`
	InFlowAddr      string   `form:"inflow_addr" binding:"required"`
	OutFlowAddr     string   `form:"outflow_addr" binding:"required"`
	Region          string   `form:"region"`
	Zone            string   `form:"zone" binding:"required"`
	Env             string   `form:"env" binding:"required"`
	Hostname        string   `form:"hostname" binding:"required"`
	Status          uint32   `form:"status" binding:"required"`
	Addrs           []string `form:"addrs"` //validate:"gt=0"
	Version         string   `form:"version"`
	Metadata        string   `form:"metadata"`
	LatestTimestamp int64    `form:"latest_timestamp"`
	DirtyTimestamp  int64    `form:"dirty_timestamp"`
	FromZone        bool     `form:"from_zone"`
}

type ArgRenew struct {
	ServiceId      string `form:"service_id" binding:"required"`
	InFlowAddr     string `form:"inflow_addr" binding:"required"`
	OutFlowAddr    string `form:"outflow_addr" binding:"required"`
	Zone           string `form:"zone" validate:"required"`
	Env            string `form:"env" validate:"required"`
	Hostname       string `form:"hostname" validate:"required"`
	Status         uint32 `form:"status" validate:"required"`
	DirtyTimestamp int64  `form:"dirty_timestamp"`
	FromZone       bool   `form:"from_zone"`
}

type ArgLogOff struct {
	Zone            string `form:"zone" validate:"required"`
	Env             string `form:"env" validate:"required"`
	ServiceId       string `form:"service_id" binding:"required"`
	Hostname        string `form:"hostname" validate:"required"`
	FromZone        bool   `form:"from_zone"`
	LatestTimestamp int64  `form:"latest_timestamp"`
}

type ArgFetchAll struct {
	ServiceId string `form:"service_id" binding:"required"`
}

type ArgDoWatch struct {
	ServiceId string `form:"service_id" binding:"required"`
}
