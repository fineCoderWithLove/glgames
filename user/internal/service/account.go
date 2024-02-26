package service

import (
	"context"
	"glgames/common/logs"
	"glgames/core/repo"
	"glgames/user/pb"
)

type AccountService struct {
	pb.UnimplementedUserServiceServer
}

func NewAccountService(manager *repo.Manager) *AccountService {
	return &AccountService{}
}
func (a *AccountService) Register(context.Context, *pb.RegisterParams) (*pb.RegisterResponse, error) {
	//写业务逻辑方法
	logs.Info("register server called.....")
	return &pb.RegisterResponse{Uid: "10000"}, nil

}
