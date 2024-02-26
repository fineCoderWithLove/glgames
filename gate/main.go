package main

import (
	"context"
	"flag"
	"fmt"
	"glgames/common/config"
	"glgames/common/metrics"
	"glgames/gate/app"
	"log"
	"os"
)

var configFile = flag.String("config", "application.yml", "config file")

func main() {
	//1加载配置
	flag.Parse()
	config.InitConfig(*configFile)
	//2.启动监控
	go func() {
		err := metrics.Serve(fmt.Sprintf("0.0.0.0:%d", config.Conf.MetricPort))
		if err != nil {
			panic(err)
		}
	}()
	//3启动grpc
	err := app.Run(context.Background())
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}
