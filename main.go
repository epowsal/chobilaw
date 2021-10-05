package main

import (
	"encoding/binary"
	"encoding/json"
	"filelock"
	"io/ioutil"
	"log"
	"math/rand"
	"orderdatafile"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/epowsal/toolfunc"

	"github.com/d5/tengo/v2"

	"fyne.io/fyne/v2/dialog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	widgetx "fyne.io/x/fyne/widget"
	"github.com/epowsal/orderfile"
)

var allsitedb *orderfile.OrderFile
var runsitemax, runsitecurcnt int32
var runsitecntmu sync.Mutex
var startallbtn *widget.Button
var stopallsite bool

func websitedownload() {
	// 网站获取下载权利后，判断获取所有链接是否完毕
	// 	没有完毕：
	// 		获取一个没有调用过目录链接，用目录链接调用获取新目录链接函数;
	// 	完毕：
	// 		判断更新时间是否到了
	// 		没到：
	// 			获取一个数据链接，用数据链接调用获取数据函数
	// 		到了：
	// 			获取一次更新
	// 	退出，等待下一次权利
	var tasksites []*websiteBaseInfo
	for true {
		var tasksite *websiteBaseInfo
		var tasksitepos int = -1
		if stopallsite == false && len(tasksites) < 1 && runsitecurcnt < runsitemax {
			tasksite := aw.GetTaskSite()
			if tasksite != nil {
				tasksite.StopTaskTime = time.Now().Add(30 * time.Minute)
				tasksites = append(tasksites, tasksite)
				atomic.AddInt32(&runsitecurcnt, 1)
				log.Println("new task site", tasksite.Id, "runsitecurcnt", runsitecurcnt, "runsitemax", runsitemax)
				if runsitecurcnt > 0 {
					startallbtn.SetText("暂停全部")
				}
			}
		}

		if len(tasksites) > 0 {
			tasksitepos = rand.Intn(len(tasksites))
			tasksite = tasksites[tasksitepos]
		}

		if len(tasksites) == 0 || tasksite == nil {
			log.Println("no task", "tasksites count", len(tasksites), "runsitecurcnt", runsitecurcnt, "runsitemax", runsitemax)
			time.Sleep(5 * time.Second)
			continue
		}

		if !time.Now().After(tasksite.NextDownloadTime) {
			if len(tasksites) > 1 {
				continue
			} else {
				time.Sleep(5 * time.Second)
			}
			continue
		}
		if time.Now().After(tasksite.StopTaskTime) || runsitecurcnt > runsitemax || stopallsite {
			var bstop bool
			runsitecntmu.Lock()
			if runsitecurcnt > runsitemax || stopallsite {
				log.Println("remove task site", tasksite.Id)
				atomic.AddInt32(&runsitecurcnt, -1)
				if runsitecurcnt <= 0 {
					startallbtn.SetText("开始全部")
				}
				aw.StopTaskSite(tasksite)
				tasksites = append(tasksites[:tasksitepos], tasksites[tasksitepos+1:]...)
				bstop = true
			}
			runsitecntmu.Unlock()
			if bstop {
				continue
			}
		}
		if tasksite.RunScript3Once && len(tasksite.Script3) > 0 {
			tasksite.NextDownloadTime = time.Now().Add(time.Duration(tasksite.DownloadOneStopMillisecond) * time.Millisecond)
			if tasksite.Script3Compile == nil || tasksite.Script1ChangeOnce {
				src := tasksite.Script3
				script := tengo.NewScript([]byte(src))
				script.SetImports(tasksite.chobilaw.GetModuleMap())
				compiled, err := script.Compile()
				if err == nil {
					tasksite.Script3Compile = compiled
				}
				tasksite.Script1ChangeOnce = false
			}
			if tasksite.Script3Compile != nil {
				tasksite.Script3Compile.Run()

				foundcataurls := tasksite.Script3Compile.Get("foundcataurls").Map()
				founditemurls := tasksite.Script3Compile.Get("founditemurls").Map()

				keybt := make([]byte, 8)
				for url, _ := range foundcataurls {
					binary.BigEndian.PutUint64(keybt, orderdatafile.BKDRHash([]byte(url)))
					if !(tasksite.uhdata.ExistsKey(keybt) || tasksite.chobilaw.no_down_catalog_url_db.KeyExists(keybt)) {
						tasksite.chobilaw.no_down_catalog_url_db.RealRmPush(keybt, []byte(url))
					}
				}
				for url, _ := range founditemurls {
					binary.BigEndian.PutUint64(keybt, orderdatafile.BKDRHash([]byte(url)))
					if !(tasksite.uhdata.ExistsKey(keybt) || tasksite.chobilaw.no_down_item_url_db.KeyExists(keybt)) {
						tasksite.chobilaw.no_down_item_url_db.RealRmPush(keybt, []byte(url))
					}
				}
			}
			tasksite.RunScript3Once = false
		} else if tasksite.chobilaw.CataUrlCount() > 0 || len(tasksite.Script1) > 0 && tasksite.uhdata.Count() == 0 {
			//log.Println("script 1 task")
			tasksite.NextDownloadTime = time.Now().Add(time.Duration(tasksite.DownloadOneStopMillisecond) * time.Millisecond)
			if tasksite.Script1Compile == nil || tasksite.Script2ChangeOnce {
				src := tasksite.Script1
				script := tengo.NewScript([]byte(src))
				script.SetImports(tasksite.chobilaw.GetModuleMap())
				compiled, err := script.Compile()
				if err == nil {
					tasksite.Script1Compile = compiled
				}
				tasksite.Script1ChangeOnce = false
			}
			if tasksite.Script1Compile != nil {
				tasksite.Script1Compile.Run()

				if tasksite.chobilaw.lasttaskhttpcode == "200" {

					foundcataurls := tasksite.Script1Compile.Get("foundcataurls").Map()
					founditemurls := tasksite.Script1Compile.Get("founditemurls").Map()
					tasksite.chobilaw.pagehtml = tasksite.Script1Compile.Get("pagehtml").String()
					headurl := tasksite.chobilaw.catataskurl
					if tasksite.chobilaw.catataskurl != tasksite.chobilaw.pageurl {
						tasksite.uhdata.Put(orderdatafile.BKDRHash([]byte(tasksite.chobilaw.catataskurl)), orderdatafile.ZipEncode(nil, []byte(tasksite.chobilaw.pageurl), 6))
						headurl = tasksite.chobilaw.pageurl
					}
					tasksite.uhdata.Put(orderdatafile.BKDRHash([]byte(headurl)), orderdatafile.ZipEncode(nil, []byte(headurl+"\n"+tasksite.chobilaw.pagehtml), 6))

					keybt := make([]byte, 8)
					for url, _ := range foundcataurls {
						binary.BigEndian.PutUint64(keybt, orderdatafile.BKDRHash([]byte(url)))
						if !(tasksite.uhdata.ExistsKey(keybt) || tasksite.chobilaw.no_down_catalog_url_db.KeyExists(keybt)) {
							tasksite.chobilaw.no_down_catalog_url_db.RealRmPush(keybt, []byte(url))
						}
					}
					for url, _ := range founditemurls {
						binary.BigEndian.PutUint64(keybt, orderdatafile.BKDRHash([]byte(url)))
						if !(tasksite.uhdata.ExistsKey(keybt) || tasksite.chobilaw.no_down_item_url_db.KeyExists(keybt)) {
							tasksite.chobilaw.no_down_item_url_db.RealRmPush(keybt, []byte(url))
						}
					}

					if len(tasksite.Script4) > 0 && (tasksite.Script4Compile == nil || tasksite.Script4ChangeOnce) {
						src := tasksite.Script4
						script := tengo.NewScript([]byte(src))
						script.SetImports(tasksite.chobilaw.GetModuleMap())
						compiled, err := script.Compile()
						if err == nil {
							tasksite.Script4Compile = compiled
						}
						tasksite.Script4ChangeOnce = false
					}
					if tasksite.Script4Compile != nil {
						tasksite.Script4Compile.Run()

						foundnamegraph := tasksite.Script4Compile.Get("foundnamegraph").Map()
						for key, val := range foundnamegraph {
							if name_attributes.KeyExists([]byte(key)) {
								fullkey, fke := name_attributes.FillKey([]byte(key))
								if fke {
									omp := make(map[string]interface{}, 0)
									json.Unmarshal(fullkey[len(key)+1:], &omp)
									omp["site"+strconv.Itoa(tasksite.Id)] = val
									njnctt, njnctte := json.Marshal(omp)
									if njnctte == nil {
										name_attributes.RealRmPush([]byte(key), njnctt)
									}
								}
							} else {
								u32id := name_attributes.NewId()
								u32idbt := make([]byte, 4+len(key))
								binary.BigEndian.PutUint32(u32idbt[:4], u32id)
								copy(u32idbt[4:], []byte(key))
								id_name.RealPush(u32idbt)

								m := val.(tengo.Map)
								m.Value["id"] = &tengo.Int{Value: int64(u32id)}
								jctt, jctte := json.Marshal(m.Value)
								if jctte == nil {
									name_attributes.RealRmPush([]byte(key), jctt)
								}
							}
						}
					}

					tasksite.chobilaw.no_down_catalog_url_db.RealRm(orderdatafile.BKDRHashB([]byte(tasksite.chobilaw.catataskurl)))
				} else {

				}
			}
		} else if tasksite.chobilaw.ItemUrlCount() > 0 || len(tasksite.Script2) > 0 && tasksite.uhdata.Count() == 0 {
			//log.Println("script 2 task")
			tasksite.NextDownloadTime = time.Now()
			if tasksite.Script2Compile == nil || tasksite.Script3ChangeOnce {
				src := tasksite.Script2
				script := tengo.NewScript([]byte(src))
				script.SetImports(tasksite.chobilaw.GetModuleMap())
				compiled, err := script.Compile()
				if err == nil {
					tasksite.Script2Compile = compiled
				}
				tasksite.Script1ChangeOnce = false
			}
			if tasksite.Script2Compile != nil {
				tasksite.Script2Compile.Run()

				if tasksite.chobilaw.lasttaskhttpcode == "200" {
					foundcataurls := tasksite.Script2Compile.Get("foundcataurls").Map()
					founditemurls := tasksite.Script2Compile.Get("founditemurls").Map()
					tasksite.chobilaw.pagehtml = tasksite.Script2Compile.Get("pagehtml").String()
					headurl := tasksite.chobilaw.itemtaskurl
					//log.Println("taskurl pageurl", tasksite.chobilaw.itemtaskurl, tasksite.chobilaw.pageurl)
					if tasksite.chobilaw.itemtaskurl != tasksite.chobilaw.pageurl {
						tasksite.uhdata.Put(orderdatafile.BKDRHash([]byte(tasksite.chobilaw.itemtaskurl)), orderdatafile.ZipEncode(nil, []byte(tasksite.chobilaw.pageurl), 6))
						headurl = tasksite.chobilaw.pageurl
					}
					tasksite.uhdata.Put(orderdatafile.BKDRHash([]byte(headurl)), orderdatafile.ZipEncode(nil, []byte(headurl+"\n"+tasksite.chobilaw.pagehtml), 6))

					keybt := make([]byte, 8)
					for url, _ := range foundcataurls {
						binary.BigEndian.PutUint64(keybt, orderdatafile.BKDRHash([]byte(url)))
						if !(tasksite.uhdata.ExistsKey(keybt) || tasksite.chobilaw.no_down_catalog_url_db.KeyExists(keybt)) {
							//log.Println("new cata url", url)
							tasksite.chobilaw.no_down_catalog_url_db.RealRmPush(keybt, []byte(url))
						}
					}
					for url, _ := range founditemurls {
						binary.BigEndian.PutUint64(keybt, orderdatafile.BKDRHash([]byte(url)))
						if !(tasksite.uhdata.ExistsKey(keybt) || tasksite.chobilaw.no_down_item_url_db.KeyExists(keybt)) {
							//log.Println("new item url", url, tasksite.uhdata.ExistsKey(keybt), tasksite.chobilaw.no_down_item_url_db.KeyExists(keybt), keybt)
							tasksite.chobilaw.no_down_item_url_db.RealRmPush(keybt, []byte(url))
						}
					}

					if len(tasksite.Script4) > 0 && (tasksite.Script4Compile == nil || tasksite.Script4ChangeOnce) {
						src := tasksite.Script4
						script := tengo.NewScript([]byte(src))
						script.SetImports(tasksite.chobilaw.GetModuleMap())
						compiled, err := script.Compile()
						if err == nil {
							tasksite.Script4Compile = compiled
						}
						tasksite.Script4ChangeOnce = false
					}
					if tasksite.Script4Compile != nil {
						tasksite.Script4Compile.Run()

						foundnamegraph := tasksite.Script4Compile.Get("foundnamegraph").Map()
						for key, val := range foundnamegraph {
							if name_attributes.KeyExists([]byte(key)) {
								fullkey, fke := name_attributes.FillKey([]byte(key))
								if fke {
									omp := make(map[string]interface{}, 0)
									json.Unmarshal(fullkey[len(key)+1:], &omp)
									omp["site"+strconv.Itoa(tasksite.Id)] = val
									njnctt, njnctte := json.Marshal(omp)
									if njnctte == nil {
										name_attributes.RealRmPush([]byte(key), njnctt)
									}
								}
							} else {
								u32id := name_attributes.NewId()
								u32idbt := make([]byte, 4+len(key))
								binary.BigEndian.PutUint32(u32idbt[:4], u32id)
								copy(u32idbt[4:], []byte(key))
								id_name.RealPush(u32idbt)

								m := val.(tengo.Map)
								m.Value["id"] = &tengo.Int{Value: int64(u32id)}
								jctt, jctte := json.Marshal(m.Value)
								if jctte == nil {
									name_attributes.RealRmPush([]byte(key), jctt)
								}
							}
						}
					}

					delrl := tasksite.chobilaw.no_down_item_url_db.RealRm(orderdatafile.BKDRHashB([]byte(tasksite.chobilaw.itemtaskurl)))
					if delrl == false {
						log.Println("item url db remove result", delrl, orderdatafile.BKDRHashB([]byte(tasksite.chobilaw.itemtaskurl)))
					}
				} else {

				}
			}
		} else {
			runsitecntmu.Lock()
			log.Println("new task site", tasksite.Id)
			atomic.AddInt32(&runsitecurcnt, -1)
			if runsitecurcnt <= 0 {
				startallbtn.SetText("开始全部")
			}
			aw.StopTaskSite(tasksite)
			tasksites = append(tasksites[:tasksitepos], tasksites[tasksitepos+1:]...)
			runsitecntmu.Unlock()
			continue
		}
	}
}

var sitemaxspeed float32 = 17 * 1024 * 1024
var memfree, memminfree int64 = 32 * 1024 * 1024 * 1024, 256 * 1024 * 1024

func startdownloadallsite() {
	// 网站获取下载权利机制：
	// 	检测下载速度没有达，内存没满，线程个数没满：
	// 		开一个下载线程：
	// 			获取10个下载网站：
	// 				获得一个下载任务：
	// 					下载
	// 				检测到要求退出一些，没完成，尝试退出
	// 				继续，或时间到了退出。
	// 			继续获取一个下载网站
	// 	爆满：
	// 		停百分之个下载线程
	// 		等待几秒
	// 		继续
	// 	满了：
	// 		等待几秒
	// 		继续

	time.Sleep(5 * time.Second)
	for true {
		if speeddown < sitemaxspeed && runsitemax < int32(aw.GetCount()) && runsitecurcnt >= runsitemax && runsitemax < 360 { // && memfree < memminfree {
			atomic.AddInt32(&runsitemax, 1)
			go websitedownload()
			time.Sleep(200 * time.Millisecond)
		} else if speeddown > sitemaxspeed || runsitemax > int32(aw.GetCount()) {
			log.Println("runsitemax-1", runsitemax, speeddown, sitemaxspeed, aw.GetCount())
			atomic.AddInt32(&runsitemax, -1)
			time.Sleep(500 * time.Millisecond)
		} else {
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

func initdb(w fyne.Window) {
	// var err error
	// opt := orderfile.DefaultOption()
	// allsitedb, err = orderfile.OpenOrderFile("data/allsitedb", opt)
	// if err != nil {
	// 	dialog.ShowError(err, w)
	// 	panic(err)
	// }

}

func main() {
	a := app.NewWithID("epowsal.chobilaw")
	a.SetIcon(appicon)
	a.Settings().SetTheme(&myTheme{})
	mwin := a.NewWindow("Chobilaw Epowsal")

	chobilawopt = &ChobilawOption{StopAllSite: false}
	_, opte := os.Stat("chobilaw.opt")
	if opte == nil {
		ctt, _ := ioutil.ReadFile("chobilaw.opt")
		json.Unmarshal(ctt, chobilawopt)
	}

	applock := filelock.Lock("Epowsal Chobilaw" + toolfunc.AppDir())
	if applock <= 0 {
		dialog.ShowConfirm("提示", "已经在运行了，确定后自动关闭。", func(b bool) {
			mwin.Close()
		}, mwin)
		mwin.Resize(fyne.NewSize(400, 400))
		mwin.ShowAndRun()
		return
	}
	logf, logfe := os.OpenFile("Chobilaw.log", os.O_CREATE|os.O_WRONLY, 0666)
	if logfe != nil {
		dialog.ShowConfirm("提示", "日志文件创建失败，确定后自动关闭。", func(b bool) {
			mwin.Close()
		}, mwin)
		mwin.Resize(fyne.NewSize(400, 400))
		mwin.ShowAndRun()
		return
	}
	endpos, _ := logf.Seek(0, os.SEEK_END)
	if endpos > 100*1024*1024 {
		logf.Seek(0, os.SEEK_SET)
		logf.Truncate(0)
	}
	log.SetOutput(logf)

	startallbtn = widget.NewButton("暂停全部", func() {
		if startallbtn.Text == "暂停全部" {
			stopallsite = true
			chobilawopt.StopAllSite = true
			optctt, _ := json.Marshal(chobilawopt)
			ioutil.WriteFile("chobilaw.opt", optctt, 0666)
		} else {
			stopallsite = false
			chobilawopt.StopAllSite = false
			optctt, _ := json.Marshal(chobilawopt)
			ioutil.WriteFile("chobilaw.opt", optctt, 0666)
		}
	})
	if chobilawopt.StopAllSite {
		stopallsite = true
		startallbtn.SetText("开始全部")
	} else {
		stopallsite = false
		startallbtn.SetText("暂停全部")
	}

	aw = NewAddWebsite()
	aw.Load()

	initdb(mwin)
	go startdownloadallsite()

	//search tab
	searchedit := widget.NewEntry()
	searchbtn := widget.NewButton("Search", func() {})
	searchhbox := container.NewBorder(nil, nil, nil, searchbtn, searchedit)
	homepage := container.NewBorder(searchhbox, nil, nil, nil, container.NewMax())
	hometab := container.NewTabItem("搜索", homepage)

	favoritecr := container.NewMax()
	favtab := container.NewTabItem("收藏", favoritecr)

	addwebsitebtn := widget.NewButton("添加编辑网站", func() {
		child := a.NewWindow("添加编辑网站")

		if chobilawopt.StopAllSite {
			stopallsite = true
			startallbtn.SetText("开始全部")
		} else {
			stopallsite = false
			startallbtn.SetText("暂停全部")
		}

		testresultw := widget.NewMultiLineEntry()
		testresultw.SetText(
			`script 1-4:
	chobilaw.download(string) string
script1:
	chobilaw.getcatatask() string
script2:
	chobilaw.getitemtask() string
script4:
	chobilaw.getpageurl() string
	chobilaw.getpagehtml() string
script 1-3 output:
	foundcataurls  map[string]
	founditemurls  map[string]
script4 output:
	foundnamegraph map[string]interface
script 1-2 output:
	[pageurl],pagehtml	string
`)
		testresultw.Wrapping = fyne.TextTruncate

		var gpageurl, gpagehtml string
		script1 := widget.NewMultiLineEntry()
		script1url := widget.NewEntry()
		testbtn1 := widget.NewButton("测试", func() {
			src := script1.Text
			script := tengo.NewScript([]byte(src))
			chobilaw := NewTestChobilaw(script1url.Text, "")
			script.SetImports(chobilaw.GetModuleMap())
			compiled, err := script.Compile()
			if err == nil {
				err = compiled.Run()
				if err != nil {
					testresultw.SetText(err.Error())
					return
				}
				fcu := compiled.Get("foundcataurls").Map()
				fiu := compiled.Get("founditemurls").Map()
				fng := compiled.Get("foundnamegraph").Map()
				pageurl := compiled.Get("pageurl").String()
				pagehtml := compiled.Get("pagehtml").String()

				gpageurl = pageurl
				gpagehtml = pagehtml

				j1, _ := json.MarshalIndent(fcu, "", "\t")
				j2, _ := json.MarshalIndent(fiu, "", "\t")
				j3, _ := json.MarshalIndent(fng, "", "\t")
				str := "pageurl:" + pageurl + "\npagehtml:" + pagehtml + "\nfoundcataurls:\n" + string(j1) + "\nfounditemurls:\n" + string(j2) + "\nfoundnamegraph:\n" + string(j3) + "\n"
				testresultw.SetText(str)
			} else {
				testresultw.SetText(err.Error())
			}
		})
		test1h := container.NewBorder(nil, nil, testbtn1, nil, script1url)
		test1v := container.NewVBox(script1, test1h)

		script2 := widget.NewMultiLineEntry()
		script2url := widget.NewEntry()
		testbtn2 := widget.NewButton("测试", func() {
			src := script2.Text
			script := tengo.NewScript([]byte(src))
			chobilaw := NewTestChobilaw("", script2url.Text)
			script.SetImports(chobilaw.GetModuleMap())
			compiled, err := script.Compile()
			if err == nil {
				err = compiled.Run()
				if err != nil {
					testresultw.SetText(err.Error())
					return
				}
				fcu := compiled.Get("foundcataurls").Map()
				fiu := compiled.Get("founditemurls").Map()
				fng := compiled.Get("foundnamegraph").Map()
				pageurl := compiled.Get("pageurl").String()
				pagehtml := compiled.Get("pagehtml").String()

				gpageurl = pageurl
				gpagehtml = pagehtml

				j1, _ := json.MarshalIndent(fcu, "", "\t")
				j2, _ := json.MarshalIndent(fiu, "", "\t")
				j3, _ := json.MarshalIndent(fng, "", "\t")
				str := "pageurl:" + pageurl + "\npagehtml:" + pagehtml + "\nfoundcataurls:\n" + string(j1) + "\nfounditemurls:\n" + string(j2) + "\nfoundnamegraph:\n" + string(j3) + "\n"
				testresultw.SetText(str)
			} else {
				testresultw.SetText(err.Error())
			}
		})
		test2h := container.NewBorder(nil, nil, testbtn2, nil, script2url)
		test2v := container.NewVBox(script2, test2h)

		script3 := widget.NewMultiLineEntry()
		testbtn3 := widget.NewButton("测试", func() {
			src := script3.Text
			script := tengo.NewScript([]byte(src))
			chobilaw := NewTestChobilaw("", "")
			script.SetImports(chobilaw.GetModuleMap())
			compiled, err := script.Compile()
			if err == nil {
				err = compiled.Run()
				if err != nil {
					testresultw.SetText(err.Error())
					return
				}
				fcu := compiled.Get("foundcataurls").Map()
				fiu := compiled.Get("founditemurls").Map()
				fng := compiled.Get("foundnamegraph").Map()
				pageurl := compiled.Get("pageurl").String()
				pagehtml := compiled.Get("pagehtml").String()

				j1, _ := json.MarshalIndent(fcu, "", "\t")
				j2, _ := json.MarshalIndent(fiu, "", "\t")
				j3, _ := json.MarshalIndent(fng, "", "\t")
				str := "pageurl:" + pageurl + "\npagehtml:" + pagehtml + "\nfoundcataurls:\n" + string(j1) + "\nfounditemurls:\n" + string(j2) + "\nfoundnamegraph:\n" + string(j3) + "\n"
				testresultw.SetText(str)
			} else {
				testresultw.SetText(err.Error())
			}
		})
		test3h := container.NewBorder(nil, nil, testbtn3, nil)
		test3v := container.NewVBox(script3, test3h)

		script4 := widget.NewMultiLineEntry()
		testbtn4 := widget.NewButton("测试", func() {
			src := script4.Text
			script := tengo.NewScript([]byte(src))
			chobilaw := NewTestChobilaw("", "")
			chobilaw.pageurl = gpageurl
			chobilaw.pagehtml = gpagehtml
			script.SetImports(chobilaw.GetModuleMap())
			compiled, err := script.Compile()
			if err == nil {
				err = compiled.Run()
				if err != nil {
					testresultw.SetText(err.Error())
					return
				}
				fcu := compiled.Get("foundcataurls").Map()
				fiu := compiled.Get("founditemurls").Map()
				fng := compiled.Get("foundnamegraph").Map()
				pageurl := compiled.Get("pageurl").String()
				pagehtml := compiled.Get("pagehtml").String()

				j1, _ := json.MarshalIndent(fcu, "", "\t")
				j2, _ := json.MarshalIndent(fiu, "", "\t")
				j3, _ := json.MarshalIndent(fng, "", "\t")
				str := "pageurl:" + pageurl + "\npagehtml:" + pagehtml + "\nfoundcataurls:\n" + string(j1) + "\nfounditemurls:\n" + string(j2) + "\nfoundnamegraph:\n" + string(j3) + "\n"
				testresultw.SetText(str)
			} else {
				testresultw.SetText(err.Error())
			}
		})
		test4h := container.NewBorder(nil, nil, testbtn4, nil)
		test4v := container.NewVBox(script4, test4h)

		var name3 *widgetx.CompletionEntry
		stopck := widget.NewCheck("是", nil)
		downloadonestopmsec := widget.NewEntry()

		name1 := widgetx.NewCompletionEntry(aw.SetCatalogFilter(""))
		name1.OnConfirm = func(t string) {
			aw.SetCatalogFilter(t)
		}
		name1.OnChanged = func(v string) {
			name1.SetOptions(aw.SetCatalogFilter(v))
			name1.ShowCompletion()
		}

		name2 := widgetx.NewCompletionEntry([]string{})
		name2.OnConfirm = func(t string) {
			si := aw.GetByUrl(t)
			if si != nil {
				name1.SetTextSilient(aw.GetCataName(si.CataId))
				name3.SetTextSilient(si.Name)
				script1.SetText(si.Script1)
				script1url.SetText(si.Script1Url)
				script2.SetText(si.Script2)
				script2url.SetText(si.Script2Url)
				script3.SetText(si.Script3)
				testresultw.SetText(si.TestResult)
				stopck.SetChecked(si.StopRun)
				downloadonestopmsec.SetText(strconv.Itoa(si.DownloadOneStopMillisecond))
			}
		}
		name2.OnChanged = func(v string) {
			name2.SetOptions(aw.SetUrlFilter(v))
			name2.ShowCompletion()
		}

		name3 = widgetx.NewCompletionEntry([]string{})
		name3.OnConfirm = func(t string) {
			si := aw.GetByName(t)
			if si != nil {
				name1.SetTextSilient(aw.GetCataName(si.CataId))
				name2.SetTextSilient(si.Url)
				script1.SetText(si.Script1)
				script1url.SetText(si.Script1Url)
				script2.SetText(si.Script2)
				script2url.SetText(si.Script2Url)
				script3.SetText(si.Script3)
				testresultw.SetText(si.TestResult)
				stopck.SetChecked(si.StopRun)
				downloadonestopmsec.SetText(strconv.Itoa(si.DownloadOneStopMillisecond))
			}
		}
		name3.OnChanged = func(v string) {
			name3.SetOptions(aw.SetNameFilter(v))
			name3.ShowCompletion()
		}

		newsiteform := widget.NewForm(
			widget.NewFormItem("网站类别", name1),
			widget.NewFormItem("抓取网站地址", name2),
			widget.NewFormItem("网站名称", name3),
			widget.NewFormItem("停止运行", stopck),
			widget.NewFormItem("下载间隔毫秒", downloadonestopmsec),
			widget.NewFormItem("用目录链接获取下层目录与条目链接脚本", test1v),
			widget.NewFormItem("用条目链接获取条目内容脚本", test2v),
			widget.NewFormItem("定时更新条目链接脚本", test3v),
			widget.NewFormItem("获取名字属性脚本", test4v),
			widget.NewFormItem("测试结果", testresultw),
		)
		var savebtn *widget.Button
		savebtn = widget.NewButton("保存", func() {
			if name2.Text == "" || name3.Text == "" {
				return
			}
			var basesite *websiteBaseInfo = nil
			if aw.Urlsel == -1 && aw.Namesel == -1 {
				for _, bi := range aw.allbaseinfo {
					if bi.Url == name2.Text || bi.Name == name3.Text {
						var bt bool
						dialog.ShowConfirm("提示", "已经存在，要替换吗？", func(b bool) {
							bt = b
							if b == true {
								basesite = bi
							} else {
								name1.SetTextSilient(aw.GetCataName(bi.CataId))
								return
							}
						}, child)
						if bt == false {
							return
						}
					}
				}
				if basesite == nil {
					basesite = aw.NewSiteInfo(name1.Text)
				}
			} else {
				if aw.Urlsel != -1 {
					basesite = aw.GetById(aw.Urlsel)
				} else if aw.Namesel != -1 {
					basesite = aw.GetById(aw.Namesel)
				}
			}

			basesite.CataId = aw.NewOrGetCataId(name1.Text)
			basesite.Url = name2.Text
			basesite.Name = name3.Text
			basesite.Script1 = script1.Text
			basesite.Script1Url = script1url.Text
			basesite.Script2 = script2.Text
			basesite.Script2Url = script2url.Text
			basesite.Script3 = script3.Text
			basesite.TestResult = testresultw.Text
			basesite.StopRun = stopck.Checked
			basesite.Script1ChangeOnce = true
			basesite.Script2ChangeOnce = true
			basesite.Script3ChangeOnce = true
			aw.Urlsel = basesite.Id
			aw.Namesel = basesite.Id
			aw.RefreshTaskOnce = true
			aw.Save()
			aw.catalogsel = basesite.CataId
			savebtn.SetText("已经保存")
			go func(savebtn *widget.Button) {
				time.Sleep(2 * time.Second)
				savebtn.SetText("保存")
			}(savebtn)
		})

		vbox := container.NewVBox(newsiteform, savebtn, startallbtn)
		script1.Resize(fyne.NewSize(944, 688))
		child.SetContent(vbox)
		child.Resize(fyne.NewSize(944, 688))
		child.CenterOnScreen()
		child.Show()
	})
	addtablebtn := widget.NewButton("添加编辑表", func() {

	})
	addarticalbtn := widget.NewButton("添加编辑文章", func() {

	})
	addfilebtn := widget.NewButton("添加编辑文件", func() {

	})
	addfolderbtn := widget.NewButton("添加目录", func() {

	})
	addcr := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(90, 90)), addwebsitebtn, addtablebtn, addarticalbtn, addfilebtn, addfolderbtn)
	addtab := container.NewTabItem("添加", addcr)

	optioncr := container.NewMax()
	opttab := container.NewTabItem("设置", optioncr)

	tabs := container.NewAppTabs(hometab, favtab, addtab, opttab)

	mwin.SetContent(container.NewMax(tabs))
	mwin.Resize(fyne.NewSize(944, 688))
	mwin.CenterOnScreen()
	mwin.SetMaster()
	mwin.SetOnClosed(func() {
		stopallsite = true
		runsitemax = 0
		if runsitecurcnt > runsitemax {
			time.Sleep(1 * time.Second)
		}
		time.Sleep(3 * time.Second)
	})
	mwin.ShowAndRun()
}
