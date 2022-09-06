package apiserver

import (
	"encoding/json"
	"github.com/go-kratos/kratos/pkg/ecode"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
	"nmid-registry/pkg/loger"
	"nmid-registry/pkg/registry"
)

func Register(c *bm.Context) {
	arg := new(registry.ArgRegister)
	if err := c.Bind(arg); err != nil {
		return
	}
	ins := registry.NewInstance(arg)
	if err := c.Bind(ins); err != nil {
		return
	}

	if ins.Status == 0 || ins.Status > registry.InstanceError {
		loger.Loger.Error("params status invalid")
		return
	}

	if arg.Metadata != "" {
		// check the metadata type is json
		if !json.Valid([]byte(arg.Metadata)) {
			c.JSON(nil, ecode.RequestErr)
			loger.Loger.Error("register params metadata(%v) invalid json", arg.Metadata)
			return
		}
	}

	re.Register(c, arg, ins)

	c.JSON(true, nil)
}

func Renew(c *bm.Context) {
	arg := new(registry.ArgRenew)
	if err := c.Bind(arg); err != nil {
		return
	}

	c.JSON(re.Renew(c, arg))
}

func LogOff(c *bm.Context) {
	arg := new(registry.ArgLogOff)
	if err := c.Bind(arg); err != nil {
		return
	}

	c.JSON(nil, re.LogOff(c, arg))
}

func FetchAll(c *bm.Context) {

}

func Fetch(c *bm.Context) {

}

func Fetchs(c *bm.Context) {

}
