package douban

import (
	"TVHelper/global"
	"TVHelper/internal/common"
	"encoding/base64"
	"strconv"
	"strings"
	"math"
	"go.uber.org/zap"

	"github.com/imroc/req/v3"

	"github.com/tidwall/gjson"
)

var count = 30

func CateFilter(cateType, ext, pg, douban string) (cateFilterResult common.Result, err error) {
	var resp *req.Response
	extDecodeBytes, err := base64.StdEncoding.DecodeString(ext)
	if err != nil {
		global.Logger.Error(ext, zap.Error(err))
	}
	curPage, err := strconv.Atoi(pg)
	if err != nil {
		global.Logger.Error(pg, zap.Error(err))
	}
	switch cateType {
	case "0interests":
		status := GJsonGetDefault(gjson.GetBytes(extDecodeBytes, "status"), "mark")
		subtypeTag := gjson.GetBytes(extDecodeBytes, "subtype_tag").String()
		yearTag := GJsonGetDefault(gjson.GetBytes(extDecodeBytes, "year_tag"), "全部")
		path := strings.Join([]string{"/user/", douban, "/interests"}, "")
		resp, err = global.DouBanClient.R().SetQueryParams(map[string]string{
			"type":        "movie",
			"status":      status,
			"subtype_tag": subtypeTag,
			"year_tag":    yearTag,
			"start":       strconv.Itoa((curPage - 1) * count),
			"count":       strconv.Itoa(count),
		}).Get(path)
	}
	if err != nil {
		return
	}

	respStr := resp.String()
	total, err := strconv.Atoi(gjson.Get(respStr, "total").String())
	if err != nil {
		global.Logger.Error(respStr, zap.Error(err))
	}

	cateFilterResult = common.Result{
		Page:      curPage,
		PageCount: int(math.Ceil(float64(total) / float64(count))),
		Limit:     count,
		Total:     total,
	}

	lists := make([]common.Vod, 0)
	path := "interests.#.subject"

	gjson.Get(resp.String(), path).ForEach(func(_, v gjson.Result) bool {
		itemType := v.Get("type").String()
		if itemType == "movie" || itemType == "tv" {
			rating := GJsonGetDefault(v.Get("rating.value"), v.Get("null_rating_reason"))
			honorInfos := GJsonArrayToString(v.Get("honor_infos.#.title"), " | ")
			lists = append(lists, common.Vod{
				VodId: strings.Join([]string{"msearch:", itemType, "__", v.Get("id").String()},
					""),
				VodName: GJsonGetDefault(v.Get("title"), "暂不支持展示"),
				VodPic:     v.Get("pic.normal").String(),
				VodRemarks: strings.TrimSpace(strings.Join([]string{rating, honorInfos}, " ")),
			})
		}
		return true
	})

	cateFilterResult.List = lists

	return
}
