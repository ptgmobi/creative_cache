package cache

import (
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Conf struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type ImageInfo struct {
	Cid         string `json:"c_id"`
	Oid         string `json:"o_id"`
	Type        string `json:"type"`
	Pkgs        string `json:"pkgs"`
	Countries   string `json:"countries"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Size        int64  `json:"size"`
	Format      string `json:"format"`
	PlayTime    int    `json:"play_time"`
	OriginUrl   string `json:"origin_url"`
	OverseasUrl string `json:"overseas_url"`
	DomesticUrl string `json:"domestic_url"`
}

var defaultPool *redis.Pool

func Init(cf *Conf) {
	defaultPool = &redis.Pool{
		MaxIdle:     512,               // 最大的闲置链接
		IdleTimeout: 300 * time.Second, // 闲置链接在多久后被关闭
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(cf.Host + ":" + cf.Port)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < 10*time.Second {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// 在缓存中添加一条日志
func AddInfo() {
}

// 在缓存中读取日志
func GetInfo(cUrl string) (string, int64, error) {
	c := defaultPool.Get()
	defer c.Close()

	cInfo, err := redis.String(c.Do("Get", cUrl))
	if err != nil {
		return "", 0, err
	}

	//TODO 拆解序列化字符串
	info := strings.Split(cInfo, "_")
	if len(info) != 2 {
		return "", 0, errors.New("invalid info")
	} else {
	}
}
