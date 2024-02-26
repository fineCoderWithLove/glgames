package rpc

import (
	"context"
	"fmt"
	"glgames/common/config"
	"glgames/common/discovery"
	"glgames/common/logs"
	"glgames/user/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

var (
	UserCilent pb.UserServiceClient
)

func Init() {
	//etcd解析，链接的时候触发,通过提供的etcd地址进行解析
	r := discovery.NewResolver(config.Conf.Etcd)
	resolver.Register(r)
	userdomain := config.Conf.Domain["user"]
	initClient(userdomain.Name, userdomain.LoadBalance, &UserCilent)
}

func initClient(name string, loadBalance bool, client interface{}) {
	addr := fmt.Sprintf("etcd:///%s", name)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials())}
	if loadBalance {
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, "round_robin")))
	}
	conn, err := grpc.DialContext(context.TODO(), addr, opts...)
	if err != nil {
		logs.Fatal("rpc connect etcd  err %v", err)
	}

	switch c := client.(type) {
	case *pb.UserServiceClient:
		*c = pb.NewUserServiceClient(conn)
	default:
		logs.Fatal("unsupported client type")
	}

}
