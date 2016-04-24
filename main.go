package main

import (
	"flag"
	"github.com/coreos/etcd/clientv3"
	"github.com/gocomm/config"
	"github.com/golang/glog"
	"github.com/infrmods/xbus/api"
	"github.com/infrmods/xbus/comm"
	"github.com/infrmods/xbus/configs"
	"github.com/infrmods/xbus/services"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Etcd     comm.ETCDConfig
	Services services.Config
	Configs  configs.Config
	Api      api.Config
}

var cfg_path = flag.String("config", "", "config file path")

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	var cfg Config
	if *cfg_path == "" {
		if err := config.DefaultConfig(&cfg); err != nil {
			glog.Errorf("set default config file fail: %v", err)
			return
		}
	} else if err := config.LoadFromFileF(*cfg_path, &cfg, yaml.Unmarshal); err != nil {
		glog.Errorf("load config file fail: %v", err)
		return
	}

	etcdConfig := clientv3.Config{
		Endpoints:   cfg.Etcd.Endpoints,
		DialTimeout: cfg.Etcd.Timeout,
		TLS:         cfg.Etcd.TLS}
	etcdClient, err := clientv3.New(etcdConfig)
	if err != nil {
		glog.Errorf("create etcd clientv3 fail: %v", err)
		return
	}

	services := services.NewServices(&cfg.Services, etcdClient)
	apiServer := api.NewAPIServer(&cfg.Api, services)
	if err := apiServer.Start(); err != nil {
		glog.Errorf("start api_sersver fail: %v", err)
		return
	}
	if err := apiServer.Wait(); err != nil {
		glog.Errorf("wait api_server fail: %v", err)
		return
	}
}