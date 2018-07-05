package lib

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"time"
	"github.com/garyburd/redigo/redis"
)

func MysqlInit(user,pwd,host string,port int) *sql.DB {
	u := user
	p := pwd
	h := host
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/gamebb?charset=utf8", u, p, h, port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	return db
}

func NewPool(host string, idle,port int, timeout int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     idle,
		IdleTimeout: time.Duration(timeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
			if err != nil {
				panic(err)
			}
			return c, err
		},
	}
}

