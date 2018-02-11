package source

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// http://13.250.109.164:12222/creative?cid=img.21knofin2jnsnf & inurl= & furl= & pkgs=

var creativeCenterUrl = "http://13.250.109.164:12222/creative?"
var domesticUrl = "cloudmobi-creative-center.s3.cn-north"
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
	DemosticUrl  string   `json:"demostic_url"`
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
		log.Println("[RequestCreative] request err: ", err, " url: ", url)
		return nil
	}

	var cr CreativeResponse

	err := json.NewDecoder(resp.Body).Decode(&cr)

}

func GetWithCidOrUrl(cid, url string) *Creative {
	var requestUrl = creativeCenterUrl
	if len(cid) > 0 {
		requestUrl += "cid=" + cid
	} else if len(url) > 0 {
		w := whereUrl(url)
		switch w {
		case DOMESTIC_URL:
			requestUrl += "inurl=" + url
		case OVERSEAS_URL:
			requestUrl += "furl=" + url
		case OTHER_URL:
			// 上传新连接
		default:
		}
	}
}
