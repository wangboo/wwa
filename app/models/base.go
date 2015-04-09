package models

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"log"
	"os"
)

var DB *gorm.DB
var db gorm.DB

var Redis redis.Conn

type BaseModel struct {
	Id int `gorm:"primary_key",sql:"AUTO_INCREMENT"`
}

func (b *BaseModel) NewRecord() bool {
	return b.Id <= 0
}

func (b *BaseModel) Destory() error {
	return DB.Delete(b).Error
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
	conn, err := redis.Dial(protocol, address)
	if err != nil {
		panic(fmt.Sprintf("redis连接错误", err.Error()))
	}
	log.Println("Init redis success!!!")
	Redis = conn
}
