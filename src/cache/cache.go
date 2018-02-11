package cache

import (
	"encoding/json"
	"fmt"
	"strings"
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
	Cid     string `json:"c_id"`
	Curl    string `json:"c_url"`
	Size    int64  `json:"size"`
	MoreKey string `json:"-"`
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

// 在缓存中读取日志
func GetCreativeInfo(cUrl string) (string, int64, error) {
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
