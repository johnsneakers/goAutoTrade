package bitmex

import (
	cconf "goAutoTrade/conf"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/garyburd/redigo/redis"
	"goAutoTrade/services"
	"github.com/fatih/color"
)

type service struct {
	conf        *cconf.ServiceConf
	db          *sql.DB
	rp          *redis.Pool
	bitmex *Bitmex
}

const (
	VERSION = "0.0.1"
)

func NewService() services.Service {
	return new(service)
}

func (s *service) Name() string {
	return "bitmex"
}

func (s *service) Version() string {
	return VERSION
}

func (s *service) Config(conf *cconf.ServiceConf) error {
	s.bitmex = new(Bitmex)
	s.bitmex.Configuration = cconf.LoadBitmexConf()
	return nil
}

func (s *service) Start() {
	color.Green("%s now start!", s.Name())
	PricesTimer()
	go s.bitmex.WebsocketClient()
	go deal()
	select{}
}

