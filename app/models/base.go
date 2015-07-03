package models

import (
	"github.com/garyburd/redigo/redis"
	"github.com/revel/revel"
	"io/ioutil"
	"labix.org/v2/mgo"
	"net/http"
	"net/url"
	"time"
)

var (
	RedisPool    *redis.Pool
	DB_NAME      = "wwa"
	COL_USER_INV = "user_invs"
	session      *mgo.Session
)

// 调用游戏服务器
func GetGameServer(url string) ([]byte, error) {
	// log.Printf("url = %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

// 调用游戏服务器
func PostFormGameServer(url string, data url.Values) ([]byte, error) {
	// log.Printf("url = %s\n", url)
	resp, err := http.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func InitDatabase() {
	mongodbUrl, ok := revel.Config.String("mongodb")
	if !ok {
		panic("app.conf中找不到mongodb")
	}
	var err error
	session, err = mgo.Dial(mongodbUrl)
	if err != nil {
		panic(err)
	}
}

// 初始化redis
func InitRedis() {
	protocol := revel.Config.StringDefault("redis.protocol", "tcp")
	address := revel.Config.StringDefault("redis.address", ":6379")
	RedisPool = &redis.Pool{
		MaxIdle:     3,
		MaxActive:   50,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(protocol, address)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// 获取mgo.Session
func Session() *mgo.Session {
	return session.Clone()
}

func Col(session *mgo.Session, colName string) *mgo.Collection {
	return session.DB(DB_NAME).C(colName)
}
