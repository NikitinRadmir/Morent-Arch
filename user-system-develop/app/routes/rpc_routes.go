package routes

import (
	"user-system/app/rpc"

	"github.com/gin-gonic/gin"
)

func RegisterRPCRoutes(router *gin.Engine, rpcServer *rpc.RPCServer) {
	router.POST("/api/rpc", gin.WrapH(rpcServer))
}
