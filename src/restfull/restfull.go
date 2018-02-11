package restful

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/brg-liuwei/gotools"

	"cache"
	"loop"
	"source"
)

var ErrNil = "redigo: nil returned"

type Conf struct {
	LogPath         string `json:"log_path"`
	LogRotateBackup int    `json:"log_rotate_backup"`
	LogRotateLines  int    `json:"log_rotate_lines"`
}

type Service struct {
	conf *Conf
	l    *gotools.RotateLogger
}

type SearchResp struct {
	ErrMsg      string `json:"err_msg"`
	CreativeId  string `json:"creative_id"`
	CreativeUrl string `json:"creative_url"`
	Size        int64  `json:"size"`

	MoreInfo *cache.ImageInfo `json:"more_info,omitempty"`
}

func (sr *SearchResp) WriteTo(w http.ResponseWriter) (int, error) {
	b, _ := json.Marshal(sr)
	return w.Write(b)
}

func NewSearchResp(errMsg, cId, cUrl string, cSize int64, moreInfo *cache.ImageInfo) *SearchResp {
	return &SearchResp{
		ErrMsg:      errMsg,
		CreativeId:  cId,
		CreativeUrl: cUrl,
		Size:        cSize,
		MoreInfo:    moreInfo,
	}
}

func Init(cf *Conf) (*Service, error) {
	l, err := gotools.NewRotateLogger(cf.LogPath, "", log.LUTC|log.LstdFlags, cf.LogRotateBackup)
	if err != nil {
		return nil, err
	}
	l.SetLineRotate(cf.LogRotateLines)

	srv := &Service{
		conf: cf,
		l:    l,
	}
	return srv, nil
}

func (s *Service) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := r.ParseForm(); err != nil {
		s.l.Println("[Search] ParseForm err: ", err)
		if err := NewSearchResp("server err", "", "", 0, nil).WriteTo(w); err != nil {
			s.l.Println("[Search] server error: ", err)
		}
		return
	}

	cUrl, err := url.QueryUnescape(r.Form.Get("creative_url"))
	cId := r.Form.Get("c_id")
	more := r.Form.Get("more") // more: 1: 需要详细信息，2: 不需要

	getInfo := func(key string) *SearchResp {
		easyInfo, err := cache.GetEasyInfo(key)
		if err != nil {
			if err.Error() == ErrNil { // 没有相关信息
				// TODO
			} else {
				s.l.Printf("[Search] get easy info err: %v, key: %s", err, key)
				return NewSearchResp("get easy info err", "", "", 0, nil)
			}
		}

		if len(more) == 0 || more == "2" {
			return NewSearchResp("", easyInfo.Cid, easyInfo.Curl, easyInfo.Size, nil), nil
		} else if more == "1" {
			moreInfo, err := cache.GetMoreInfo(easyInfo.MoreKey)
			if err != nil {
				s.l.Println("[Search] get more info err: ", err, " key: ", key)
				return NewSearchResp("get more info err", "", "", 0, nil)
			}
			return NewSearchResp("", easyInfo.Cid, easyInfo.Curl, easyInfo.Size, moreInfo), nil
		}
	}

	if len(cUrl) > 0 { // 根据url查询
		if _, err := getInfo(cUrl).WriteTo(w); err != nil {
			s.l.Println("[Search] get info with url err: ", err)
		}
		return
	} else if len(cId) > 0 { // 根据cid查询
		if _, err := getInfo(cId).WriteTo(w); err != nil {
			s.l.Println("[Search] get info with url err: ", err)
		}
		return
	} else {
		s.l.Println("[Search] can't get creative_url or cid, err :", err)
		if _, err := NewSearchResp("can't get creative_url or cid", "", "", 0).WriteTo(w); err != nil {
			s.l.Println("[Search] fail to response get creative_url error: ", err)
		}
		return
	}

	// 缓存里获取到信息
	// 去db里获取信息

	// 没有获取信息需要后台获取
}

func (s *Service) Server() {
	http.HandleFunc("/cid/get_creative_id", s.HandleSearch)

	panic(http.ListenAndServe(":12121", nil))
}
