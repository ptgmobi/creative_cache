package source

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"cache"
)

// http://13.250.109.164:12222/creative?cid=img.21knofin2jnsnf & inurl= & furl= & pkgs=

var creativeCenterUrl = "http://127.0.0.1:12222/creative?"
var creativeSearchUrl = "http://127.0.0.1:12222/search/url?"
var domesticUrl = "cdn.cn.ctcnpa.com"
var overseasUrl = "cloudmobi-creative-center.s3.ap-southeast"

const (
	DOMESTIC_URL = 1
	OVERSEAS_URL = 2
	OTHER_URL    = 3
)

type Creative struct {
	Cid          string   `json:"c_id"`
	Oid          string   `json:"o_id"`
	CreativeType string   `json:"type"`
	Pkgs         []string `json:"pkgs"`
	Countries    []string `json:"countries"`
	Offers       []string `json:"offers"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Size         int64    `json:"size"`
	PlayTime     int      `json:"play_time"`
	Format       string   `json:"format"`
	OriginUrl    string   `json:"origin_url"`
	OverseasUrl  string   `json:"overseas_url"`
	DomesticUrl  string   `json:"domestic_url"`
}

type CreativeResponse struct {
	ErrMsg    string     `json:"err_msg"`
	Creatives []Creative `json:"creatives"`
}

// 1：domesticUrl 2:overseasUrl 3: other
func whereUrl(url string) int {
	if strings.Contains(url, domesticUrl) {
		return DOMESTIC_URL
	} else if strings.Contains(url, overseasUrl) {
		return OVERSEAS_URL
	}
	return OTHER_URL
}

func RequestCreative(url string) *Creative {
	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Println("[RequestCreative] request get err: ", err, " url: ", url)
		return nil
	}

	var cr CreativeResponse

	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		log.Println("[RequestCreative] decode body err: ", err)
		return nil
	}

	if len(cr.ErrMsg) > 0 {
		log.Println("[RequestCreative] request err: ", cr.ErrMsg, " url: ", url)
		return nil
	}
	if len(cr.Creatives) == 1 {
		return &cr.Creatives[0]
	}
	log.Printf("[RequestCreative] unknown err, %#v", cr)
	return nil

}

func uploadNewCreative(url, cType string, region int) *Creative {
	type Info struct {
		Url    string `json:"url"`
		CType  string `json:"type"`
		Region int    `json:"region"`
	}

	var info = Info{
		Url:    url,
		CType:  cType,
		Region: region,
	}
	body, _ := json.Marshal(&info)

	resp, err := http.Post(
		creativeCenterUrl,
		"application/json",
		strings.NewReader(string(body)),
	)
	if err != nil {
		log.Println("[uploadNewCreative] post err: ", err)
		return nil
	}
	defer resp.Body.Close()

	var cr CreativeResponse

	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		log.Println("[uploadNewCreative] decode body err: ", err)
		return nil
	}
	if len(cr.ErrMsg) > 0 {
		log.Println("[uploadNewCreative] response err: ", cr.ErrMsg)
		return nil
	}
	if len(cr.Creatives) == 1 {
		return &cr.Creatives[0]
	}
	log.Println("[uploadNewCreative] no creative")
	return nil
}

func GetWithCidOrUrl(cid, url, cType string, region int) *Creative {
	var requestUrl = creativeCenterUrl
	if len(cid) > 0 {
		requestUrl += "cid=" + cid
		return RequestCreative(requestUrl)
	} else if len(url) > 0 {
		w := whereUrl(url)
		switch w {
		case DOMESTIC_URL:
			requestUrl += "inurl=" + url
			return RequestCreative(requestUrl)
		case OVERSEAS_URL:
			requestUrl += "furl=" + url
			return RequestCreative(requestUrl)
		case OTHER_URL:
			return uploadNewCreative(url, cType, region)
		default:
			log.Println("[GetWithCidOrUrl] unknown url: ", url)
		}
	}
	return nil
}

func (c *Creative) SerializeEasyInfo() string {
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

func SearchWithUrl(cUrl string) *Creative {
	var requestUrl = creativeSearchUrl
	url := requestUrl + "ourl=" + cUrl
	c := RequestCreative(url)
	if c != nil {
		// 将简化信息写入redis
		if err := cache.Set(c.OriginUrl, c.SerializeEasyInfo(), 432000); err != nil {
			log.Println("[SearchWithUrl] set redis err: ", err, " cid: ", c.Cid)
		}
	}
	return c
}
