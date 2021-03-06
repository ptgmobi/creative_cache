package restful

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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
	ErrMsg        string `json:"err_msg"`
	CreativeId    string `json:"creative_id"`
	CreativeOldId string `json:"creative_old_id"`
	OverseasUrl   string `json:"overseas_url"`
	DomesticUrl   string `json:"domestic_url"`
	Size          int64  `json:"size"`

	MoreInfo *cache.ImageInfo `json:"more_info,omitempty"`
}

func (sr *SearchResp) WriteTo(w http.ResponseWriter) (int, error) {
	b, _ := json.Marshal(sr)
	return w.Write(b)
}

func NewSearchResp(errMsg, cId, oId, oUrl, dUrl string, cSize int64, moreInfo *cache.ImageInfo) *SearchResp {
	return &SearchResp{
		ErrMsg:        errMsg,
		CreativeId:    cId,
		CreativeOldId: oId,
		OverseasUrl:   oUrl,
		DomesticUrl:   dUrl,
		Size:          cSize,
		MoreInfo:      moreInfo,
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
		if _, err := NewSearchResp("server err", "", "", "", "", 0, nil).WriteTo(w); err != nil {
			s.l.Println("[Search] server error: ", err)
		}
		return
	}

	var regionInt int
	base64Url := r.Form.Get("curl")
	if len(base64Url) == 0 {
		if _, err := NewSearchResp("url is nil", "", "", "", "", 0, nil).WriteTo(w); err != nil {
			s.l.Println("[Search] fail to response get c_url error: ", err)
		}
		return
	}

	cUrlBytes, err := base64.StdEncoding.DecodeString(base64Url)
	if err != nil {
		s.l.Println("HandleSearch base64 decode curl err: ", err, " url: ", base64Url)
		if _, err := NewSearchResp("url can't base64 decode", "", "", "", "", 0, nil).WriteTo(w); err != nil {
			s.l.Println("[Search] fail to response get creative_url error: ", err)
		}
		return
	}
	cUrl := string(cUrlBytes)
	cType := r.Form.Get("ctype")   // img:????????? mp4?????????
	region := r.Form.Get("region") // 1: ????????? 2????????????, 3:????????? ??????3
	if len(region) == 0 {
		regionInt = 3
	} else {
		regionInt, err = strconv.Atoi(region)
		if err != nil {
			s.l.Println("HandleSearch region err: ", err, " region: ", region)
			regionInt = 3
		}
	}
	cId := r.Form.Get("cid")

	getInfo := func(key string, isUrl bool) *SearchResp {
		easyInfo, err := cache.GetEasyInfo(key)
		if err != nil {
			if err.Error() == ErrNil { // ??????????????????
				if isUrl { // url??????
					// ???????????????
					if c := source.SearchWithUrl(base64Url); c != nil {
						return NewSearchResp("", c.Cid, c.Oid, c.OverseasUrl, c.DomesticUrl, c.Size, nil)
					}
					// ????????????????????????
					if err := loop.AddUploadQueue(key, cType, regionInt); err != nil {
						s.l.Printf("[Search] add Upload queue err: %v, key:%s", err, key)
					}
				} else { // cid ??????
					// ??????????????????
					if c := source.GetWithCidOrUrl(key, "", cType, regionInt); c != nil {
						return NewSearchResp("", c.Cid, c.Oid, c.OverseasUrl, c.DomesticUrl, c.Size, nil)
					}
				}
				return NewSearchResp("no info", "", "", "", "", 0, nil)
			} else {
				s.l.Printf("[Search] get easy info err: %v, key: %s", err, key)
				return NewSearchResp("get easy info err", "", "", "", "", 0, nil)
			}
		}

		return NewSearchResp("", easyInfo.Cid, easyInfo.Oid, easyInfo.OverseasUrl, easyInfo.DomesticUrl, easyInfo.Size, nil)
	}

	if len(cUrl) > 0 && len(cType) > 0 { // ??????url??????
		if _, err := getInfo(cUrl, true).WriteTo(w); err != nil {
			s.l.Println("[Search] get info with url err: ", err)
		}
		return
	} else if len(cId) > 0 { // ??????cid??????
		if _, err := getInfo(cId, false).WriteTo(w); err != nil {
			s.l.Println("[Search] get info with url err: ", err)
		}
		return
	} else {
		s.l.Println("[Search] can't get creative_url or cid, cid:", cId, " ctype: ", cType, " url: ", cUrl)
		if _, err := NewSearchResp("can't get creative_url or cid", "", "", "", "", 0, nil).WriteTo(w); err != nil {
			s.l.Println("[Search] fail to response get creative_url error: ", err)
		}
		return
	}
}

func (s *Service) Server() {
	http.HandleFunc("/cache/get_cid", s.HandleSearch)

	panic(http.ListenAndServe(":12121", nil))
}
