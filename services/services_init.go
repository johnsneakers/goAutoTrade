package services

import (
	"runtime"
	"fmt"
	"goAutoTrade/conf"
	"reflect"
	"unsafe"
	"os"
)

type Service interface {
	Name() string
	Version() string
	Config(conf *conf.ServiceConf) error
	Start()
}

func RunService(service Service) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	conf, err := conf.LoadServiceConf(service.Name(), "")
	if err != nil {
		panic(err)
	}

	extraProgName := fmt.Sprintf("-version:%s -port:%d", service.Version(), conf.Server.Port)
	ServicesInfo(extraProgName)
	service.Config(conf)
	service.Start()
}


func ServicesInfo(ver string) {
	argv0str := (*reflect.StringHeader)(unsafe.Pointer(&os.Args[0]))
	argv0 := (*[1 << 30]byte)(unsafe.Pointer(argv0str.Data))[:]
	line := os.Args[0]
	for i := 1; i < len(os.Args); i++ {
		line += (" " + os.Args[i])
	}
	line += (" " + ver)
	copy(argv0, line)
	argv0[len(line)] = 0
}
