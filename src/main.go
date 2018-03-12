package main

import (
	"github.com/brg-liuwei/gotools"
	_ "net/http/pprof"

	"cache"
	"loop"
	"restful"
)

type Conf struct {
	RestfulConf *restful.Conf `json:"restful_conf"`
	CacheConf   *cache.Conf   `json:"cache_conf"`
}

var conf Conf

func main() {
	if err := gotools.DecodeJsonFile("conf/creative.conf", &conf); err != nil {
		panic(err)
	}

	if err := cache.Init(conf.CacheConf); err != nil {
		panic(err)
	}
	if err := loop.Init(); err != nil {
		panic(err)
	}
	restfullService, err := restful.Init(conf.RestfulConf)
	if err != nil {
		panic(err)
	}
	restfullService.Server()
}
