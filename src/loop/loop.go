package loop

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"cache"
	"source"
)

var loopQueue *Queue
var waitGroup sync.WaitGroup

type creativeInfo struct {
	url    string // 素材链接
	cType  string // img: 图片, mp4: 视频
	region int    // 1:国内， 2:国外，3:国内外 默认：3
}

func Init() error {
	loopQueue = NewQueue()
	go LoopQueue()
	return nil
}

// url: 素材链接
// ctype： img:图片，mp4:视频
// region: 1: 国内, 2:国外，默认：国外
func AddUploadQueue(url, cType string, region int) error {
	if loopQueue == nil {
		return fmt.Errorf("loopQueue is nil, add error")
	}
	ci := &creativeInfo{
		url:    url,
		cType:  cType,
		region: region,
	}

	loopQueue.Add(ci)
	return nil
}

func TopUploadQueue() (interface{}, error) {
	if loopQueue == nil {
		return nil, fmt.Errorf("loopQueue is nil, top error")
	}

	return loopQueue.Top(), nil
}

func SerializeEasyInfo(c *source.Creative) string {
	ei := cache.EasyInfo{
		Cid:         c.Cid,
		Oid:         c.Oid,
		OverseasUrl: c.OverseasUrl,
		DomesticUrl: c.DomesticUrl,
		Size:        c.Size,
	}

	eiBytes, err := json.Marshal(&ei)
	if err != nil {
		log.Println("SerializeEasyInfo marshal err: ", err, " cid: ", c.Cid)
		return ""
	}
	return string(eiBytes)

}

func LoopQueue() {
	update := func() {
		copyQueue := loopQueue.CopyQueue()
		if copyQueue.Length() == 0 {
			log.Println("[LoopQueue] queue's Length is zero")
			return
		}
		// 数据库里查询
		for ciInter := copyQueue.Top(); ciInter != nil; ciInter = copyQueue.Top() {
			ci, ok := ciInter.(*creativeInfo)
			if !ok {
				log.Printf("[LoopQueue] queue elem err, type is %s", reflect.TypeOf(ci).Name())
				continue

			}

			// 起4个线程去上传
			for copyQueue.Length()%4 == 0 {
				waitGroup.Wait()
			}

			waitGroup.Add(1)
			go func() {
				creative := source.GetWithCidOrUrl("", ci.url, ci.cType, ci.region)
				log.Println("[LoopQueue] get info with url: ", ci.url, " type: ", ci.cType, " region: ", ci.region, " queueSize:", copyQueue.Length())
				if creative == nil {
					log.Println("[LoopQueue] get info with url failed! url: ", ci.url)
					return
				}
				// 将简化信息写入redis
				if err := cache.Set(ci.url, SerializeEasyInfo(creative), 432000); err != nil {
					//if err := cache.Set(creative.Cid, SerializeEasyInfo(creative), 259200); err != nil {
					log.Println("[LoopQueue] set redis err: ", err, " cid: ", creative.Cid)
					return
				}
			}()

		}
	}

	for {
		update()
		time.Sleep(time.Second * 1)
	}
}
