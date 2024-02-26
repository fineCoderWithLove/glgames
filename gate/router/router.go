package router

import (
	"github.com/gin-gonic/gin"
	"glgames/common/config"
	"glgames/common/rpc"
	"glgames/gate/api"
)

// 注册路由
func RegisterRouter() *gin.Engine {
	if config.Conf.Log.Level == "DEBUG" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	//初始化grpc的client，然后去调用服务user的grpc服务
	rpc.Init()
	engine := gin.Default()
	userhandler := api.NewUserHandler()
	engine.POST("register", userhandler.Register)
	return engine
}
