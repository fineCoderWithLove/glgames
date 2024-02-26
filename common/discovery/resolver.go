package discovery

import (
	"glgames/common/config"
	"glgames/common/logs"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type Resolver struct {
	conf        config.EtcdConf
	etcdCli     *clientv3.Client //etcd连接
	key         string
	DialTimeout int //超时时间
	closeCh     chan struct{}
	cc          resolver.ClientConn
	srvAddrList []resolver.Address //地址
	watchCh     clientv3.WatchChan
}

func (r Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	//获取到调用的key(user/v1) 链接etcd获取val
	//建立etcd的连接
	r.cc = cc
	var err error
	r.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.conf.Addrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		logs.Fatal("grpc client connect err%v", err)
	}
	r.closeCh = make(chan struct{})
	//2.根据key获取val
	r.key = target.URL.Path
	if err := r.sync(); err != nil {
		return nil, err
	}
	go r.watch()
	return nil, nil
}

func (r Resolver) Scheme() string {
	return "etcd"
}

func (r Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.conf.RWTimeout)*time.Second)
	defer cancel()
	res, err := r.etcdCli.Get(ctx, r.key, clientv3.WithPrefix())
	if err != nil {
		logs.Error("get etcd failed,name = %s,err=%v", r.key, err)
		return err
	}
	r.srvAddrList = []resolver.Address{}
	for _, v := range res.Kvs {
		server, err := ParseValue(v.Value)
		if err != nil {
			logs.Error("grpc client update failed,name=%s,err:%v", r.key, err)
			continue
		}
		r.srvAddrList = append(r.srvAddrList, resolver.Address{
			Addr:       server.Addr,
			Attributes: attributes.New("weight", server.Weight),
		})
	}
	if len(r.srvAddrList) == 0 {
		return nil
	}
	err1 := r.cc.UpdateState(resolver.State{
		Addresses: r.srvAddrList,
	})
	if err1 != nil {
		logs.Error("get etcd failed,name = %s,err=%v", r.key, err)
		return err
	}
	return nil
}

func (r *Resolver) watch() {
	//1. 定时 1分钟同步一次数据
	//2. 监听节点的事件 从而触发不同的操作
	//3. 监听Close事件 关闭 etcd
	ticker := time.NewTicker(time.Minute)
	r.watchCh = r.etcdCli.Watch(context.Background(), r.key, clientv3.WithPrefix())
	for {
		select {
		case <-r.closeCh:
			//close
			r.Close()
		case res, ok := <-r.watchCh:
			if ok {
				//
				r.update(res.Events)
			}

		case <-ticker.C:
			if err := r.sync(); err != nil {
				logs.Error("watch sync failed,err:%v", err)
			}
		}
	}
}

func (r Resolver) update(events []*clientv3.Event) {
	for _, event := range events {
		switch event.Type {
		case clientv3.EventTypePut:
			//put key value
			server, err := ParseValue(event.Kv.Value)
			if err != nil {
				logs.Error("grpc client update(EventTypePut) parse etcd value failed, name=%s,err:%v", r.key, err)
			}
			addr := resolver.Address{
				Addr:       server.Addr,
				Attributes: attributes.New("weight", server.Weight),
			}
			if !Exist(r.srvAddrList, addr) {
				r.srvAddrList = append(r.srvAddrList, addr)
				err = r.cc.UpdateState(resolver.State{
					Addresses: r.srvAddrList,
				})
				if err != nil {
					logs.Error("grpc client update(EventTypePut) UpdateState failed, name=%s,err:%v", r.key, err)
				}
			}
		case clientv3.EventTypeDelete:
			server, err := ParseKey(string(event.Kv.Key))
			if err != nil {
				logs.Error("get etcd update(eventTypeDelete),name = %s,err=%v", r.key, err)
			}
			addr := resolver.Address{
				Addr: server.Addr,
			}
			//remove操作
			if list, ok := Remove(r.srvAddrList, addr); ok {
				r.srvAddrList = list
			}
			err1 := r.cc.UpdateState(resolver.State{
				Addresses: r.srvAddrList,
			})
			if err1 != nil {
				logs.Error("get etcd update(eventTypeDelete),name = %s,err=%v", r.key, err)
			}

		}
	}
}

func (r *Resolver) Close() {
	if r.etcdCli != nil {
		err := r.etcdCli.Close()
		if err != nil {
			logs.Error("Resolver close etcd err:%v", err)
		}
		logs.Info("close etcd...")
	}
}

func Exist(list []resolver.Address, addr resolver.Address) bool {
	for i := range list {
		if list[i].Addr == addr.Addr {
			return true
		}
	}
	return false
}

func Remove(list []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range list {
		if list[i].Addr == addr.Addr {
			list[i] = list[len(list)-1]
			return list[:len(list)-1], true
		}
	}
	return nil, false
}

func NewResolver(conf config.EtcdConf) *Resolver {
	return &Resolver{
		conf:        conf,
		DialTimeout: conf.DialTimeout,
	}
}
