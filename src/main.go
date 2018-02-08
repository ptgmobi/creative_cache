package main

import (
	"fmt"

	"github.com/brg-liuwei/gotools"

	"restful"
)

type Conf struct {
	restfulConf *restful.Conf
}

var conf Conf

func main() {
	if err := gotools.DecodeJsonFile("conf/cid.conf", &conf); err != nil {
		panic(err)
	}

	restfullService, err := restful.Init(&conf.restfulConf)
	if err != nil {
		panic(err)
	}
	go restfullService.Server()
}
