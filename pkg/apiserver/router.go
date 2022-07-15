package apiserver

import (
	"errors"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
)

var (
	writeOnly     bool
	errProtectMsg = errors.New("registry in protect mode & only can do register")
)

func HttpRouter(apiServer *ApiServer) {
	writeOnly = apiServer.IsWriteOnly()

	httpServer := apiServer.server
	group := httpServer.Group("/registry")
	{
		group.POST("/register", Register)
		group.POST("/renew", Renew)
		group.POST("/logoff", LogOff)
		group.GET("/fetch/all", WriteOnly, FetchAll)
		group.GET("/fetch", WriteOnly, Fetch)
		group.GET("/fetchs", WriteOnly, Fetchs)
		group.GET("/poll", WriteOnly, Poll)
		group.GET("/polls", WriteOnly, Polls)
	}
}

func WriteOnly(c *bm.Context) {
	if writeOnly {
		c.JSON(nil, errProtectMsg)
		c.AbortWithStatus(503)
	}
}
