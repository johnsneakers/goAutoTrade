package conf

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
	"encoding/json"
)

type ServiceConf struct {
	Name   string
	Server *ServerConf
	Redis  *RedisConf
	Mysql  *MysqlConf
}

type ServerConf struct {
	Host      string
	Port      int
	Endpoints []string
	Timeout   int
}

type MysqlConf struct {
	User     string
	Password string
	Host     string
	Port     int
}

type RedisConf struct {
	Host     string
	Password string
	Port     int
	Timeout  int
}



var (
	rootConfigPath = flag.String("c", "../../conf", "service root config path")
)


type BitmexConf struct {

}

func LoadBitmexConf() (a *BitmexConf) {
	path := *rootConfigPath + "/bitmex.json"
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	var buf []byte
	if buf, err = ioutil.ReadAll(file); err != nil {
		panic(err)
	}

	conf := &BitmexConf{}
	err = json.Unmarshal(buf,conf)
	if err != nil {
		panic(err)
	}

	return conf
}

var CONF_PATH = ""
func LoadServiceConf(name string,srv_path string) (conf *ServiceConf, err error) {
	path := ""
	if srv_path != "" {
		path = srv_path + "/" + name + ".conf"
	} else {
		flag.Parse()
		path = *rootConfigPath + "/" + name + ".conf"
		CONF_PATH = *rootConfigPath
	}

	return conf.loadconfig(path, name)
}

func (conf *ServiceConf) loadconfig(path, name string) (*ServiceConf, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buf []byte
	if buf, err = ioutil.ReadAll(file); err != nil {
		return nil, err
	}

	conf = &ServiceConf{}
	if err = toml.Unmarshal(buf, conf); err != nil {
		return nil, err
	}

	return conf, err
}
