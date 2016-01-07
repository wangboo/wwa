package models

import (
	"fmt"
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
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// 调用游戏服务器
func PostFormGameServer(url string, data url.Values) ([]byte, error) {
	// log.Printf("url = %s\n", url)
	resp, err := http.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func InitDatabase() {
	mongodbUrl, ok := revel.Config.String("mongodb")
	fmt.Println("mongodbUrl = ", mongodbUrl)
	if !ok {
		panic("app.conf中找不到mongodb")
	}
	var err error
	session, err = mgo.Dial(mongodbUrl)
	if err != nil {
		fmt.Print("mgo connect error", err)
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

func Col(s *mgo.Session, colName string) *mgo.Collection {
	return s.DB(DB_NAME).C(colName)
}

func WwaTypeForeach(fn func(int)) {
	fn(0)
	fn(1)
	fn(2)
	fn(3)
}

func WwaTypeToName(typeOfWwa int) string {
	switch typeOfWwa {
	case 0:
		return "小试身手"
	case 1:
		return "非同小可"
	case 2:
		return "炉火纯青"
	case 3:
		return "谁与争锋"
	default:
		return "未知"
	}
}

func RankToTitle(rank int) string {
	switch rank {
	case 1:
		return "战神"
	case 2:
		return "勇士"
	case 3:
		return "精英"
	default:
		return "未知"
	}
}

func RankToBuffAddStr(rank int) string {
	switch rank {
	case 1:
		return "20%"
	case 2:
		return "15%"
	case 3:
		return "10%"
	default:
		return "0%"
	}
}

func RankToBuffEndWeekendStr(typeOfWwa int) string {
	switch typeOfWwa {
	case 1:
		return "3"
	case 2:
		return "4"
	case 3:
		return "5"
	default:
		return "6"
	}
}
