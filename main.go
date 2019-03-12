package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type GameStreamInfo struct {
	SCdnType      string `json:"sCdnType"`
	IIsMaster     int    `json:"iIsMaster"`
	LChannelId    int    `json:"lChannelId"`
	LSubChannelId int    `json:"lSubChannelId"`
	LPresenterUid int    `json:"lPresenterUid"`
	SStreamName   string `json:"sStreamName"`
	SHlsUrl       string `json:"sHlsUrl"`
	SHlsUrlSuffix string `json:"sHlsUrlSuffix"`
	SHlsAntiCode  string `json:"sHlsAntiCode"`
}

type GameLiveInfo struct {
	Nick string `json:"nick"`
}

type StreamInfo struct {
	GameLiveInfo       *GameLiveInfo    `json:"gameLiveInfo"`
	GameStreamInfoList []GameStreamInfo `json:"gameStreamInfoList"`
}

type MultiStreamInfo struct {
	SDisplayName string `json:"sDisplayName"`
	IBitRate     int    `json:"iBitRate"`
}

type Stream struct {
	Status           int               `json:"status"`
	Msg              string            `json:"msg"`
	Data             []StreamInfo      `json:"data"`
	VMultiStreamInfo []MultiStreamInfo `json:"vMultiStreamInfo"`
}

type HyPlayerConfig struct {
	Html5     int     `json:"html5"`
	WEBYYHOST string  `json:"WEBYYHOST"`
	WEBYYSWF  string  `json:"WEBYYSWF"`
	WEBYYFROM string  `json:"WEBYYFROM"`
	Vappid    int     `json:"vappid"`
	Stream    *Stream `json:"stream"`
}

func main() {

	room := flag.String("room", "", "room")

	flag.Parse()

	if *room == "" {
		fmt.Println("房间ID错误，请查看主播直播页面https://huya.com/XXX，用./huya-live-recorder --room=XXX开启录制")
		return
	}

	api := fmt.Sprintf("https://www.huya.com/%s", *room)

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	resp, err := http.Get(api)

	if err != nil {
		fmt.Println("错误：", err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("错误：", err.Error())
		return
	}

	htmlStr := string(body)

	tmp := strings.Split(htmlStr, "hyPlayerConfig =")

	if len(tmp) < 2 {
		fmt.Println("错误：", "解析异常")
		return
	}

	tmp = strings.Split(tmp[1], "window.TT_LIVE_TIMING")

	if len(tmp) < 2 {
		fmt.Println("错误：", "解析异常")
		return
	}

	jsonStr := strings.Replace(tmp[0], "};", "}", 1)

	fmt.Println(jsonStr)

	var hyPlayerConfig HyPlayerConfig

	err = json.Unmarshal([]byte(jsonStr), &hyPlayerConfig)

	if err != nil {
		fmt.Println("错误：", err.Error())
		return
	}

	if hyPlayerConfig.Stream == nil {
		fmt.Println("错误：", "解析异常或主播还未开播")
		return
	}

	if len(hyPlayerConfig.Stream.Data) <= 0 {
		fmt.Println("错误：", "解析异常或主播还未开播")
		return
	}

	title := fmt.Sprintf("%s-%s", hyPlayerConfig.Stream.Data[0].GameLiveInfo.Nick, time.Now().Format("2006-01-02 15:04:05"))

	title = strings.Replace(title, " ", "-", -1)

	var m3u8 string

	for _, v := range hyPlayerConfig.Stream.Data[0].GameStreamInfoList {
		if v.SHlsUrlSuffix == "m3u8" {
			m3u8 = fmt.Sprintf("%s/%s.%s", v.SHlsUrl, v.SStreamName, v.SHlsUrlSuffix)
		}
	}

	download(m3u8, title)
}

func download(u string, title string) {

	fmt.Printf("开始下载：%s\n", title)

	c := fmt.Sprintf("ffmpeg -y -hide_banner -loglevel info -i %s -c:v copy -c:a copy %s.mp4", u, title)
	cmd := exec.Command("sh", "-c", c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		fmt.Println(err.Error(), title, "失败")
	}
}
