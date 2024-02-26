package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"glgames/common/logs"
	"glgames/common/rpc"
	"glgames/user/pb"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}
func (u *UserHandler) Register(ctx *gin.Context) {
	response, err := rpc.UserCilent.Register(context.TODO(), &pb.RegisterParams{
		Account:       "",
		Password:      "",
		LoginPlatform: 0,
		SmsCode:       "",
	})
	if err != nil {

	}
	uid := response.Uid
	// gen token
	logs.Info("uid:%s", uid)
}
