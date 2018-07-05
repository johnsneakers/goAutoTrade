package bitmex

import (
	"github.com/gorilla/websocket"
	"goAutoTrade/conf"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"github.com/fatih/color"
)

type Bitmex struct {
	WebsocketConn         *websocket.Conn
	WebsocketSubdChannels map[int]WebsocketChanInfo
	Configuration *conf.BitmexConf
	heart       *Heart
	writeSignal chan interface{}
	readSignal  DataSignal
	doneSignal  chan bool
	err         error
}

type IndexPrice struct {
	Symbol string `json:"symbol"`
	TimeStamp string `json:"timestamp"`
	Price float64 `json:"price"`
}


func PricesTimer() {
	GetAllIndexPrice()
	var ticker *time.Ticker = time.NewTicker(35 * time.Second)
	go func() {
		for _ = range ticker.C {
			GetAllIndexPrice()
		}
	}()
}

func GetAllIndexPrice()  {
	p1,err := GetIndexPrice(EOS_INDEX_PRICE_URL)
	if err != nil {
		color.Red("get eos price err:",err)
		EOS_INDEX_PRICE = nil
	} else {
		EOS_INDEX_PRICE = p1
	}

	p2,err := GetIndexPrice(ADA_INDEX_PRICE_URL)
	if err != nil {
		color.Red("get ada price err:",err)
		ADA_INDEX_PRICE = nil
	} else {
		ADA_INDEX_PRICE = p2
	}
}

var EOS_INDEX_PRICE *IndexPrice
var ADA_INDEX_PRICE *IndexPrice
func GetIndexPrice(url string) (price *IndexPrice, err error) {
	resp,err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	b,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	x := []*IndexPrice{}
	json.Unmarshal(b,&x)
	return x[0],nil
}