package models

import (
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var DB *gorm.DB
var db gorm.DB

var RedisPool *redis.Pool

type BaseModel struct {
	Id int `gorm:"primary_key",sql:"AUTO_INCREMENT"`
}

func (b *BaseModel) NewRecord() bool {
	return b.Id <= 0
}

func (b *BaseModel) Destory() error {
	return DB.Delete(b).Error
}

// 调用游戏服务器
func GetGameServer(url string) ([]byte, error) {
	log.Printf("url = %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

// 调用游戏服务器
func PostFormGameServer(url string, data url.Values) ([]byte, error) {
	log.Printf("url = %s\n", url)
	resp, err := http.PostForm(url, data)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func InitDatabase() {
	adapter := revel.Config.StringDefault("gorm.adapter", "mysql")
	databaseURI := revel.Config.StringDefault("gorm.database_uri", "")
	var err error
	db, err = gorm.Open(adapter, databaseURI)
	if err != nil {
		panic(err)
	}
	DB = &db
	db.LogMode(false)
	logger = Logger{log.New(os.Stdout, "  ", 0)}
	db.SetLogger(logger)
	if err := db.AutoMigrate(&Rank{}).Error; err != nil {
		panic(err)
	}
	db.LogMode(true)
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
