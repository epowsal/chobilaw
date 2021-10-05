package main

import (
	"charset"
	"errors"
	"htmlutil"
	"log"
	"netutil"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/d5/tengo/v2/stdlib"

	"github.com/d5/tengo/v2"
	"github.com/epowsal/orderfile"
)

type ChobilawSite struct {
	tengo.ObjectImpl
	homeurl                string
	taskurl                string
	no_down_catalog_url_db *orderfile.OrderFile
	no_down_item_url_db    *orderfile.OrderFile
	istest                 bool
	catataskurl            string
	itemtaskurl            string
	lasttaskhttpcode       string
	pageurl                string
	pagehtml               string
}

var speeddownsize int64
var speeddown float32
var lastspeedtime time.Time
var speedmu sync.Mutex

func OpenChobilaw(siteid int) *ChobilawSite {
	cw := &ChobilawSite{}
	opt := orderfile.DefaultOption()
	opt.FixKeyLen = 8
	cw.no_down_catalog_url_db, _ = orderfile.OpenOrderFile("allsite/"+strconv.Itoa(siteid)+"/cataurl", opt)
	opt2 := orderfile.DefaultOption()
	opt2.FixKeyLen = 8
	cw.no_down_item_url_db, _ = orderfile.OpenOrderFile("allsite/"+strconv.Itoa(siteid)+"/itemurl", opt2)
	return cw
}

func NewTestChobilaw(catataskurl, itemtaskurl string) *ChobilawSite {
	cw := &ChobilawSite{istest: true, catataskurl: catataskurl, itemtaskurl: itemtaskurl}
	return cw
}

func (ch *ChobilawSite) TypeName() string {
	return "chobilaw"
}

func (ch *ChobilawSite) String() string {
	return "chobilaw object"
}

func (ch *ChobilawSite) GetModuleMap() *tengo.ModuleMap {
	stdm := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	stdm.AddBuiltinModule("chobilaw", map[string]tengo.Object{
		"getcatatask":   &tengo.UserFunction{Name: "getcatatask", Value: stdlib.FuncARS(ch.GetCataTask)},
		"getitemtask":   &tengo.UserFunction{Name: "getitemtask", Value: stdlib.FuncARS(ch.GetItemTask)},
		"download":      &tengo.UserFunction{Name: "download", Value: ch.Download},
		"parseurl":      &tengo.UserFunction{Name: "parseurl", Value: ch.ParseUrl},
		"filterurl":     &tengo.UserFunction{Name: "filterurl", Value: ch.FilterUrl},
		"getpageurl":    &tengo.UserFunction{Name: "getpageurl", Value: stdlib.FuncARS(ch.GetPageUrl)},
		"getpagehtml":   &tengo.UserFunction{Name: "getpagehtml", Value: stdlib.FuncARS(ch.GetPageHtml)},
		"getfulltag":    &tengo.UserFunction{Name: "getfulltag", Value: ch.GetFullTag},
		"gettagbyregex": &tengo.UserFunction{Name: "gettagbyregex", Value: ch.GetTagByRegex},
	})
	return stdm
}

func (ch *ChobilawSite) GetCataTask() string {
	if ch.istest {
		return ch.catataskurl
	}
	fk := ch.no_down_catalog_url_db.RandGet()
	if fk != nil {
		ch.catataskurl = string(fk[8:])
		log.Println("cata task url", ch.no_down_catalog_url_db.Count(), ch.catataskurl)
		return ch.catataskurl
	}
	return ""
}

func (ch *ChobilawSite) GetItemTask() string {
	if ch.istest {
		return ch.itemtaskurl
	}
	// cur := []byte{}
	// for true {
	// 	next, bn := ch.no_down_item_url_db.NextKey(cur)
	// 	if bn == false {
	// 		break
	// 	}
	// 	log.Println("itemurl list:", string(next[8:]))
	// 	cur = next
	// }
	fk := ch.no_down_item_url_db.RandGet()
	if fk != nil {
		ch.itemtaskurl = string(fk[8:])
		log.Println("item task url", ch.no_down_item_url_db.Count(), ch.itemtaskurl)
		return ch.itemtaskurl
	}
	return ""
}

func (ch *ChobilawSite) Close() {
	ch.no_down_catalog_url_db.Close()
	ch.no_down_item_url_db.Close()
	ch.no_down_catalog_url_db = nil
	ch.no_down_item_url_db = nil
}

func (ch *ChobilawSite) CataUrlCount() int64 {
	return ch.no_down_catalog_url_db.Count()
}

func (ch *ChobilawSite) ItemUrlCount() int64 {
	return ch.no_down_item_url_db.Count()
}

func (ch *ChobilawSite) Download(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	url, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	ctt, _, _, httpcode, redi := netutil.UrlGet(url, nil, false, nil, nil, 8*time.Second, 8*time.Second, nil, ".*?text/html.*", "")
	if httpcode == 200 {
		if redi == "" {
			redi = url
		}

		if charset.IsGb18030(ctt) {
			ctt = []byte(charset.Convert(string(ctt), "GB18030", "UTF-8"))
		}
		if !charset.IsUtf8(ctt) {
			name := charset.DetectCharset(ctt)
			if name != "" {
				ctt = []byte(charset.Convert(string(ctt), name, "UTF-8"))
			}
		}

		ch.pageurl = redi
		ret := make(map[string]tengo.Object, 0)
		ret["html"] = &tengo.String{Value: string(ctt)}
		ret["redi"] = &tengo.String{Value: string(redi)}
		ret["httpcode"] = &tengo.String{Value: strconv.Itoa(httpcode)}
		ch.lasttaskhttpcode = strconv.Itoa(httpcode)
		speedmu.Lock()
		speeddownsize += int64(len(ctt))
		if time.Since(lastspeedtime).Seconds() > 5 {
			speeddown = float32(speeddownsize) / float32(time.Since(lastspeedtime).Seconds())
			lastspeedtime = time.Now()
			speeddownsize = 0
		}
		speedmu.Unlock()
		return &tengo.Map{Value: ret}, nil
	} else {
		log.Println("download error", httpcode, url)
	}
	ret := make(map[string]tengo.Object, 0)
	ret["httpcode"] = &tengo.String{Value: strconv.Itoa(httpcode)}
	ch.lasttaskhttpcode = strconv.Itoa(httpcode)
	return &tengo.Map{Value: ret}, nil
}

func (ch *ChobilawSite) ParseUrl(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	url, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "PrseUrl arg1",
			Expected: "string(compatible)",
			Found:    args[0].TypeName(),
		}
	}
	pagehtml, ok2 := tengo.ToString(args[1])
	if !ok2 {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "PrseUrl arg2",
			Expected: "string(compatible)",
			Found:    args[1].TypeName(),
		}
	}

	urlmp := htmlutil.GetAllLinkUrlForTengo(url, []byte(pagehtml))
	return &tengo.Map{Value: urlmp}, nil
}

func (ch *ChobilawSite) FilterUrl(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	urlmp := args[0].(*tengo.Map)
	partern, ok2 := tengo.ToString(args[1])
	if !ok2 {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string(compatible)",
			Found:    args[1].TypeName(),
		}
	}

	c, ce := regexp.Compile(partern)
	if ce != nil {
		return nil, errors.New("regex partern error")
	}
	ret := make(map[string]tengo.Object, 0)
	for key, _ := range urlmp.Value {
		if c.Match([]byte(key)) {
			ret[key] = tengo.TrueValue
		}
	}
	return &tengo.Map{Value: ret}, nil
}

func (ch *ChobilawSite) GetPageUrl() string {
	return string(ch.pageurl)
}

func (ch *ChobilawSite) GetPageHtml() string {
	return string(ch.pagehtml)
}

func (ch *ChobilawSite) GetFullTag(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	html, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "html",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	tagrestr, ok2 := tengo.ToString(args[1])
	if !ok2 {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "tag start regex with group 1 tag name",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	taghtml, err := htmlutil.GetFullTag([]byte(html), tagrestr)
	if err != nil {
		return nil, err
	}
	return &tengo.String{Value: string(taghtml)}, nil
}

func (ch *ChobilawSite) GetTagByRegex(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	html, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "html",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	tagrestr, ok2 := tengo.ToString(args[1])
	if !ok2 {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "tag start regex with group 1 tag name",
			Expected: "string",
			Found:    args[1].TypeName(),
		}
	}

	re, ree := regexp.Compile(tagrestr)
	if ree != nil {
		return nil, errors.New("regex express have error")
	}

	allind := re.FindAllSubmatchIndex([]byte(html), -1)
	retstr := ""
	for i := 0; i < len(allind); i++ {
		for j := 2; j < len(allind[i]); j += 2 {
			retstr += html[allind[i][j]:allind[i][j+1]]
		}
	}
	return &tengo.String{Value: retstr}, nil
}
