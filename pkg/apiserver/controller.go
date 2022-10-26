package apiserver

import (
	"encoding/json"
	"fmt"
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

	err := re.Register(c, arg, ins)
	fmt.Println(`err`, err)
	if nil != err {
		c.JSON(false, nil)
		return
	}

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
	arg := new(registry.ArgFetchAll)
	if err := c.Bind(arg); err != nil {
		return
	}

	ret, err := re.FetchAll(c, arg)

	c.JSON(ret, err)
}

func DoWatch(c *bm.Context) {
	arg := new(registry.ArgDoWatch)
	if err := c.Bind(arg); err != nil {
		return
	}

	ret, err := re.DoWatch(c, arg)

	c.JSON(ret, err)
}
