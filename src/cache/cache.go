package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Conf struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// EasyInfo.MoreKey -> ImageInfo
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

// cid|url -> EasyInfo
type EasyInfo struct {
	Cid         string `json:"c_id"`
	Oid         string `json:"o_id"`
	OverseasUrl string `json:"o_url"`
	DemosticUrl string `json:"d_url"`

	Size    int64  `json:"size"`
	MoreKey string `json:"-"`
}

var defaultPool *redis.Pool

func Init(cf *Conf) error {
	defaultPool = &redis.Pool{
		MaxIdle:     512,               // 最大的闲置链接
		IdleTimeout: 300 * time.Second, // 闲置链接在多久后被关闭
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cf.Host+":"+cf.Port)
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
	return nil
}

func Get(key string) (string, error) {
	if len(key) == 0 {
		return "", fmt.Errorf("key is nil")
	}

	c := defaultPool.Get()
	defer c.Close()

	info, err := redis.String(c.Do("Get", key))
	if err != nil {
		return "", err
	}
	return info, nil
}

func Set(key, value string, expire int) error {
	if len(key) == 0 || len(value) == 0 {
		return fmt.Errorf("redis set key is nil")
	}

	if expire == 0 {
		expire = 259200 // 60 * 60 * 24 * 3 // 三天
	}

	c := defaultPool.Get()
	defer c.Close()

	if err := c.Send("SET", key, value); err != nil {
		return err
	}
	if err := c.Send("EXPIRE", key, expire); err != nil {
		return err
	}

	return c.Flush()
}

func GetEasyInfo(key string) (*EasyInfo, error) {
	info, err := Get(key)
	if err != nil {
		return nil, err
	}

	var easyInfo EasyInfo
	if err := json.Unmarshal([]byte(info), &easyInfo); err != nil {
		return nil, fmt.Errorf("get info err: %v, key: %s", err, key)
	}
	return &easyInfo, nil
}

func GetMoreInfo(key string) (*ImageInfo, error) {
	info, err := Get(key)
	if err != nil {
		return nil, err
	}

	var moreInfo ImageInfo
	if err := json.Unmarshal([]byte(info), &moreInfo); err != nil {
		return nil, fmt.Errorf("get more info err: %v , key: %s", err, key)
	}
	return &moreInfo, nil
}
