package main

import (
	"encoding/json"
	_ "fmt"
	"io/ioutil"
	"log"
	"orderdatafile"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epowsal/orderfile"

	"github.com/d5/tengo/v2"
)

var aw *AddWebsite

type websiteCatalog struct {
	Name string
	Id   int
}

type websiteBaseInfo struct {
	Url        string
	Name       string
	CataId     int
	Id         int
	Script1    string
	Script1Url string
	Script2    string
	Script2Url string
	Script3    string
	Script4    string
	Script4Url string
	TestResult string

	NextStartTime              time.Time `json:"-"`
	StopTaskTime               time.Time `json:"-"`
	RunScript3Once             bool      `json:"-"`
	CataUrlCount               int
	ItemUrlCount               int
	StopRun                    bool
	Runing                     bool `json:"-"`
	DownloadOneStopMillisecond int
	NextDownloadTime           time.Time `json:"-"`

	Script1Compile    *tengo.Compiled         `json:"-"`
	Script2Compile    *tengo.Compiled         `json:"-"`
	Script3Compile    *tengo.Compiled         `json:"-"`
	Script4Compile    *tengo.Compiled         `json:"-"`
	Script1ChangeOnce bool                    `json:"-"`
	Script2ChangeOnce bool                    `json:"-"`
	Script3ChangeOnce bool                    `json:"-"`
	Script4ChangeOnce bool                    `json:"-"`
	chobilaw          *ChobilawSite           `json:"-"`
	uhdata            *orderdatafile.DataFile `json:"-"`
}
type AddWebsite struct {
	settingpath                          string
	catalogfilter, Urlfilter, Namefilter string
	catalogsel, Urlsel, Namesel          int
	allcatalog                           []*websiteCatalog
	allbaseinfo                          []*websiteBaseInfo
	catals, Urlls, Namels                []string
	maxSiteId, maxCatalogId              int
	lock                                 sync.RWMutex
	NextSortTime                         time.Time
	taskCurI                             int
	allbaseinfofortask                   []*websiteBaseInfo
	RefreshTaskOnce                      bool
}

func NewAddWebsite() *AddWebsite {
	return &AddWebsite{catalogsel: -1, Urlsel: -1, Namesel: -1, maxSiteId: 0, maxCatalogId: 0}
}

func (aw *AddWebsite) Load() {
	catactt, catactte := ioutil.ReadFile(aw.settingpath + "addWebsiteCatalog.txt")
	if catactte == nil {
		json.Unmarshal(catactt, &aw.allcatalog)
	} else {
		basecata := []string{"综合", "其它", "软件硬件开发", "电子产品及周边", "自然科学", "数学", "化学", "英语语言", "艺术", "生物", "物理", "经济", "医学", "法律", "土木工程", "机械", "设计", "图片", "视频", "音乐", "新闻", "游戏", "理工科研", "下载", "产品商品", "军事"}
		for i, catan := range basecata {
			aw.allcatalog = append(aw.allcatalog, &websiteCatalog{catan, i + 1})
		}
		aw.maxCatalogId = len(aw.allcatalog)
	}
	allbaseinfoctt, allbaseinfoctte := ioutil.ReadFile(aw.settingpath + "addWebsiteBaseInfo.txt")
	if allbaseinfoctte == nil {
		json.Unmarshal(allbaseinfoctt, &aw.allbaseinfo)
	}

	sitmaxId, sitmaxIde := ioutil.ReadFile(aw.settingpath + "addWebsiteMaxId.txt")
	if sitmaxIde == nil {
		ls := strings.Split(string(sitmaxId), ",")
		aw.maxSiteId, _ = strconv.Atoi(ls[0])
		aw.maxCatalogId, _ = strconv.Atoi(ls[1])
	}
}

func (aw *AddWebsite) Save() {
	ctt, ctte := json.Marshal(&aw.allcatalog)
	if ctte != nil {
		panic(ctte)
	}
	ioutil.WriteFile(aw.settingpath+"addWebsiteCatalog.txt", ctt, 0666)
	ctt, ctte = json.Marshal(&aw.allbaseinfo)
	if ctte != nil {
		panic(ctte)
	}
	ioutil.WriteFile(aw.settingpath+"addWebsiteBaseInfo.txt", ctt, 0666)
	ioutil.WriteFile(aw.settingpath+"addWebsiteMaxId.txt", []byte(strconv.Itoa(aw.maxSiteId)+","+strconv.Itoa(aw.maxCatalogId)), 0666)
}

func (aw *AddWebsite) SetCatalogFilter(flt string) []string {
	aw.catalogfilter = flt
	aw.Urlsel = -1
	aw.Namesel = -1
	aw.catals = []string{}
	for _, val := range aw.allcatalog {
		if FilterSearch(val.Name, strings.Split(flt, " ")) {
			aw.catals = append(aw.catals, val.Name)
		}
		if val.Name == flt {
			aw.catalogsel = val.Id
		}
	}
	return aw.catals
}

func (aw *AddWebsite) SetUrlFilter(flt string) []string {
	aw.Urlfilter = flt
	aw.Urlsel = -1
	aw.Urlls = []string{}
	for _, val := range aw.allbaseinfo {
		if FilterSearch(val.Url, strings.Split(flt, " ")) && (aw.catalogsel == -1 || aw.catalogsel == val.CataId) {
			aw.Urlls = append(aw.Urlls, val.Url)
		}
		if val.Url == flt {
			aw.Urlsel = val.Id
		}
	}
	return aw.Urlls
}

func (aw *AddWebsite) SetNameFilter(flt string) []string {
	aw.Namefilter = flt
	aw.Namesel = -1
	aw.Namels = []string{}
	for _, val := range aw.allbaseinfo {
		if FilterSearch(val.Name, strings.Split(flt, " ")) && (aw.catalogsel == -1 || aw.catalogsel == val.CataId) {
			aw.Namels = append(aw.Namels, val.Name)
		}
		if val.Name == flt {
			aw.Namesel = val.Id
		}
	}
	return aw.Namels
}

func (aw *AddWebsite) NewSiteInfo(cataName string) *websiteBaseInfo {
	var CataId int = -1
	for _, v := range aw.allcatalog {
		if v.Name == cataName {
			CataId = v.Id
		}
	}
	if CataId == -1 {
		aw.maxCatalogId += 1
		aw.allcatalog = append([]*websiteCatalog{&websiteCatalog{cataName, aw.maxCatalogId}}, aw.allcatalog...)
		CataId = aw.maxCatalogId
	}
	aw.maxSiteId += 1
	ws := &websiteBaseInfo{Id: aw.maxSiteId, CataId: CataId}
	aw.allbaseinfo = append([]*websiteBaseInfo{ws}, aw.allbaseinfo...)
	return ws
}

func (aw *AddWebsite) NewOrGetCataId(cataName string) int {
	var CataId int = -1
	for _, v := range aw.allcatalog {
		if v.Name == cataName {
			return v.Id
		}
	}
	if CataId == -1 {
		aw.maxCatalogId += 1
		aw.allcatalog = append([]*websiteCatalog{&websiteCatalog{cataName, aw.maxCatalogId}}, aw.allcatalog...)
		return aw.maxCatalogId
	}
	return -1
}

func FilterSearch(val string, flt []string) bool {
	for _, findt := range flt {
		if len(findt) == 0 {
			continue
		}
		if strings.Index(val, findt) == -1 {
			return false
		}
	}
	return true
}

func (aw *AddWebsite) GetByUrl(url string) *websiteBaseInfo {
	for _, v := range aw.allbaseinfo {
		if v.Url == url {
			return v
		}
	}
	return nil
}

func (aw *AddWebsite) GetByName(name string) *websiteBaseInfo {
	for _, v := range aw.allbaseinfo {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (aw *AddWebsite) GetById(id int) *websiteBaseInfo {
	for _, v := range aw.allbaseinfo {
		if v.Id == id {
			return v
		}
	}
	return nil
}

func (aw *AddWebsite) GetCataName(cataid int) string {
	for _, v := range aw.allcatalog {
		if v.Id == cataid {
			return v.Name
		}
	}
	return ""
}

func (aw *AddWebsite) GetCount() int {
	cnt := 0
	aw.lock.Lock()
	for _, wb := range aw.allbaseinfo {
		if wb.StopRun == false {
			cnt += 1
		}
	}
	aw.lock.Unlock()
	return cnt
}

func (aw *AddWebsite) GetTaskSite() (owb *websiteBaseInfo) {
	//依据数据量，网站活跃度来，全部网站更新一次要多久：1000 1天1次    10000 1天1次  100000 2 天一次 1000000 13天下载一次
	if aw.NextSortTime.Before(time.Now()) || aw.RefreshTaskOnce {
		aw.allbaseinfofortask = make([]*websiteBaseInfo, 0)
		for _, wb := range aw.allbaseinfo {
			if wb.StopRun == false {
				aw.allbaseinfofortask = append(aw.allbaseinfofortask, wb)
			}
		}
		IQuickSort(aw.allbaseinfofortask, 0, len(aw.allbaseinfofortask)-1, func(left, right interface{}) int {
			l := left.(*websiteBaseInfo)
			r := right.(*websiteBaseInfo)
			lv := l.CataUrlCount*100000000 + l.ItemUrlCount
			rv := r.CataUrlCount*100000000 + r.ItemUrlCount
			if lv < rv {
				return 1
			} else if lv > rv {
				return -1
			} else {
				return 0
			}
		})
		if len(aw.allbaseinfofortask) < 10000 {
			aw.NextSortTime = time.Now().Add(24 * time.Hour)
		} else if len(aw.allbaseinfofortask) < 100000 {
			aw.NextSortTime = time.Now().Add(2 * 24 * time.Hour)
		} else if len(aw.allbaseinfofortask) < 200000 {
			aw.NextSortTime = time.Now().Add(3 * 24 * time.Hour)
		} else {
			aw.NextSortTime = time.Now().Add(time.Duration(float32(len(aw.allbaseinfofortask))/70000) * 24 * time.Hour)
		}
		aw.taskCurI = 0
	}
	aw.lock.Lock()

	for i := aw.taskCurI; i < len(aw.allbaseinfofortask); i++ {
		if aw.allbaseinfofortask[i] == nil {
			panic("error")
		}
		//log.Println(aw.allbaseinfofortask[i])
		if aw.allbaseinfofortask[i].NextStartTime.Before(time.Now()) && aw.allbaseinfofortask[i].Runing == false && aw.allbaseinfofortask[i].StopRun == false {
			aw.allbaseinfofortask[i].Runing = true
			owb = aw.allbaseinfofortask[i]
			owb.RunScript3Once = true
			owb.chobilaw = OpenChobilaw(owb.Id)
			aw.taskCurI = i
			log.Println("found task site", owb.Id)
			break
		}
	}
	if owb == nil {
		aw.taskCurI = 0
		for i := aw.taskCurI; i < len(aw.allbaseinfofortask); i++ {
			if aw.allbaseinfofortask[i].NextStartTime.Before(time.Now()) && aw.allbaseinfofortask[i].Runing == false && aw.allbaseinfofortask[i].StopRun == false {
				aw.allbaseinfofortask[i].Runing = true
				owb = aw.allbaseinfofortask[i]
				owb.RunScript3Once = true
				owb.chobilaw = OpenChobilaw(owb.Id)
				aw.taskCurI = i
				log.Println("2found task site", owb.Id)
				break
			}
		}
	}
	aw.lock.Unlock()
	if owb != nil {
		if owb.uhdata == nil {
			owb.uhdata = orderdatafile.OpenDataFile("allsite/"+strconv.Itoa(owb.Id)+"/uhdata", 3, 30)
		}
		if owb.chobilaw == nil {
			owb.chobilaw = OpenChobilaw(owb.Id)
		}
		if owb.chobilaw.no_down_catalog_url_db == nil {
			opt := orderfile.DefaultOption()
			opt.FixKeyLen = 8
			owb.chobilaw.no_down_catalog_url_db, _ = orderfile.OpenOrderFile("allsite/"+strconv.Itoa(owb.Id)+"/cataurl", opt)
		}
		if owb.chobilaw.no_down_item_url_db == nil {
			opt := orderfile.DefaultOption()
			opt.FixKeyLen = 8
			owb.chobilaw.no_down_item_url_db, _ = orderfile.OpenOrderFile("allsite/"+strconv.Itoa(owb.Id)+"/itemurl", opt)
		}
	}
	return owb
}

func (aw *AddWebsite) StopTaskSite(wb *websiteBaseInfo) {
	aw.lock.Lock()
	wb.Runing = false
	if wb.chobilaw.no_down_catalog_url_db.Count() == 0 && wb.chobilaw.no_down_item_url_db.Count() == 0 {
		wb.NextStartTime = aw.NextSortTime
	}
	wb.CataUrlCount = int(wb.chobilaw.CataUrlCount())
	wb.ItemUrlCount = int(wb.chobilaw.ItemUrlCount())
	wb.chobilaw.Close()

	if wb.uhdata != nil {
		wb.uhdata.Close()
		wb.uhdata = nil
	}
	if wb.chobilaw.no_down_catalog_url_db != nil {
		wb.chobilaw.no_down_catalog_url_db.Close()
	}
	if wb.chobilaw.no_down_item_url_db != nil {
		wb.chobilaw.no_down_item_url_db.Close()
	}

	wb.chobilaw = nil
	aw.lock.Unlock()
}

func ISwap(a []*websiteBaseInfo, i, j int) {
	if i == j {
		return
	}
	temp := a[j]
	a[j] = a[i]
	a[i] = temp
}
func IQuickSort(a []*websiteBaseInfo, left, right int, cmp func(left, right interface{}) int) {
	if left >= right {
		return
	}
	i := left
	j := right
	for i < j {
		for i < j && cmp(a[left], a[j]) <= 0 {
			j--
		}
		ISwap(a, i, j)
		for i < j && cmp(a[left], a[i]) >= 0 {
			i++
		}
		ISwap(a, i, j)
	}
	IQuickSort(a, left, i-1, cmp)
	IQuickSort(a, i+1, right, cmp)
}
