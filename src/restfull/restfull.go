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
)

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
	ErrMsg     string `json:"err_msg"`
	CreativeId string `json:"creative_id"`
	Size       int64  `json:"size"`
}

func (sr *SearchResp) WriteTo(w http.ResponseWriter) int {
	b, _ := json.Marshal(sr)
	return w.Write(b)
}

func NewSearchResp(errMsg, cId, cType string, cSize int64) *SearchResp {
	return &Resp{
		ErrMsg:     errMsg,
		CreativeId: cId,
		Size:       cSize,
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
		if n, err := NewSearchResp("server err", "", "", 0).WriteTo(w); err != nil {
			s.l.Println("[Search] server error, resp write: ", n, ", error: ", err)
		}
		return
	}

	cUrl, err := url.QueryUnescape(r.Form.Get("creative_url"))
	if err != nil || len(cUrl) == 0 {
		s.l.Println("[Search] can't get creative_url, err :", err)
		if n, err := NewSearchResp("can't get creative_url", "", "", 0).WriteTo(w); err != nil {
			s.l.Println("[Search] can't get creative_url, resp write: ", n, " error: ", err)
		}
		return
	}

	cType := r.Form.Get("type")
	if len(cType) == 0 {
		cType = "1"
	}

	cId, cSize, err := cache.GetCreativeInfo(cUrl)
	// 获取到信息
	if err == nil && len(cId) > 0 {
		if n, err := NewSearchResp("", cId, cType, cSize).WriteTo(w); err != nil {
			s.l.Println("[Search] fail to response cache cId, cUrl: ", cUrl, ", resp write: ", n, ", error: ", err)
		}
		return
	}

	if n, err := NewSearchResp("cache no info", "", "", 0).WriteTo(w); err != nil {
		s.l.Println("[Search] fail to response no info err: ", err)
		return
	}

	// 没有获取信息需要后台获取
}

func (s *Service) HandleDump(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func (s *Service) Server() {
	http.HandleFunc("/cid/get_creative_id", s.HandleSearch)
	http.HandleFunc("/cid/dump", s.HandleDump)

	panic(http.ListenAndServe(":12121", nil))
}
