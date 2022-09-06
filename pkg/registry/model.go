package registry

type ArgRegister struct {
	ServiceId       string   `form:"service_id" binding:"required"`
	InFlowUrl       string   `form:"inflow_url" binding:"required"`
	OutFlowUrl      string   `form:"outflow_url" binding:"required"`
	Region          string   `form:"region"`
	Zone            string   `form:"zone" validate:"required"`
	Env             string   `form:"env" validate:"required"`
	Hostname        string   `form:"hostname" validate:"required"`
	Status          uint32   `form:"status" validate:"required"`
	Addrs           []string `form:"addrs" validate:"gt=0"`
	Version         string   `form:"version"`
	Metadata        string   `form:"metadata"`
	LatestTimestamp int64    `form:"latest_timestamp"`
	DirtyTimestamp  int64    `form:"dirty_timestamp"`
	FromZone        bool     `form:"from_zone"`
}

type ArgRenew struct {
	Zone           string `form:"zone" validate:"required"`
	Env            string `form:"env" validate:"required"`
	ServiceId      string `form:"service_id" binding:"required"`
	Hostname       string `form:"hostname" validate:"required"`
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
