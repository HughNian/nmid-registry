package apiserver

import (
	"errors"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
	"nmid-registry/pkg/registry"
)

var (
	re        *registry.Registry
	writeOnly bool
	errMsg    = errors.New("registry in protect mode & only can do register")
)

func DoApiServer(apiServer *ApiServer) {
	writeOnly = apiServer.IsWriteOnly()

	re = registry.NewRegistry(apiServer.cluster)

	HttpRouter(apiServer.server)
}

func HttpRouter(httpServer *bm.Engine) {
	group := httpServer.Group("/registry")
	{
		group.POST("/register", Register)
		group.POST("/renew", Renew)
		group.POST("/logoff", LogOff)
		group.GET("/fetch/all", WriteOnly, FetchAll)
		group.POST("/watch", WriteOnly, DoWatch)
	}
}

//WriteOnly if route write only can't do read operator like as fetch, fetchs fetchAll
func WriteOnly(c *bm.Context) {
	if writeOnly {
		c.JSON(nil, errMsg)
		c.AbortWithStatus(503)
	}
}
