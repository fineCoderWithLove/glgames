package app

import (
	"context"
	"glgames/common/config"
	"glgames/common/discovery"
	"glgames/common/logs"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 启动程序 启动grpc服务和http服务，启动日志和数据库
func Run(ctx context.Context) error {
	//1。做一个日志库，info，error，fatal，debug颜色不同显示

	//2.做一个etcd的注册中心
	register := discovery.NewRegister()
	//启动grpc服务端
	server := grpc.NewServer()
	go func() {
		lis, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			logs.Fatal("user grpc server listen err %v", err)
		}
		//注册gprc service到etcd 注册数据库mongo redis

		err2 := register.Register(config.Conf.Etcd)
		if err2 != nil {
			logs.Fatal("user grpc server register err %v", err)
		}
		//初始化数据库管理

		err = server.Serve(lis)
		if err != nil {
			logs.Fatal("user grpc failed err:%v", err)
		}
	}()
	// 优雅启停，中断，推出，暂停
	stop := func() {
		server.Stop()
		register.Close()
		time.Sleep(3 * time.Second)
		logs.Info("stop app finish")
	}
	//期望有一个优雅启停 遇到中断 退出 终止 挂断
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGHUP)
	for {
		select {
		case <-ctx.Done():
			stop()
			//time out
			return nil
		case s := <-c:
			switch s {
			case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
				stop()
				logs.Info("user app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Info("hang up!! user app quit")
				return nil
			default:
				return nil
			}
		}
	}
}
