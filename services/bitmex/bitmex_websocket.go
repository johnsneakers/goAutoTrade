package bitmex

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"goAutoTrade/common"
	"github.com/shopspring/decimal"
)

const (
	RBUFFER = 512
	bitmexWebsocket                   = "wss://www.bitmex.com/realtime"
	EOS_INDEX_PRICE_URL = "https://www.bitmex.com/api/v1/trade?symbol=.EOSXBT&count=20&columns=price&reverse=true"
	ADA_INDEX_PRICE_URL = "https://www.bitmex.com/api/v1/trade?symbol=.ADAXBT&count=20&columns=price&reverse=true"
)

type DataSignal chan *Response
var DataBingo = NewDataSignal()

type Response struct {
	/*Success   bool        `json:"success,omitempty"`
	Subscribe string      `json:"subscribe,omitempty"`
	Request   interface{} `json:"request,omitempty"`*/
	Table     string      `json:"table,omitempty"`
	Action    string      `json:"action,omitempty"`
	Data      []Keys `json:"data,omitempty"`
}

type BitmexDepth struct {
	Id     int64   `json:"id,omitempty"`
	Symbol string  `json:"symbol,omitempty"`
	Side   string  `json:"side,omitempty"`
	Price  float64 `json:"price,omitempty"`
	Size   float64 `json:"size,omitempty"`
}


type Keys struct {
	Id     int64   `json:"id,omitempty"`
	Symbol string  `json:"symbol,omitempty"`
	Side   string  `json:"side,omitempty"`
	Price  float64 `json:"price,omitempty"`
	Size   float64 `json:"size,omitempty"`
	ImpactMidPrice float64 `json:"impactMidPrice"`
}

func (b *Bitmex) WebsocketPingHandler() error {
	request := make(map[string]string)
	request["op"] = "ping"
	return b.WebsocketSend(request)
}

func (b *Bitmex) WebsocketSend(data interface{}) error {
	json, err := common.JSONEncode(data)
	if err != nil {
		return err
	}

	return b.WebsocketConn.WriteMessage(websocket.TextMessage, json)
}



func (b *Bitmex) WebsocketSendAuth() error {
	request := make(map[string]interface{})
	payload := "AUTH" + strconv.FormatInt(time.Now().UnixNano(), 10)[:13]
	request["event"] = "auth"
	request["apiKey"] = b.Configuration.APIKey
	request["authSig"] = common.HexEncodeToString(common.GetHMAC(common.HashSHA512_384, []byte(payload), []byte(b.Configuration.APISecret)))
	request["authPayload"] = payload

	return b.WebsocketSend(request)
}


func (b *Bitmex) GetConnection() (err error) {
	if b.WebsocketConn != nil {
		return
	}

	var Dialer websocket.Dialer
	color.Yellow("connect to...%s ...",bitmexWebsocket)
	b.WebsocketConn, _, err = Dialer.Dial(bitmexWebsocket, http.Header{})
	if err != nil {
		color.Red(" connect fail, err %v", err)
		return
	}

	color.Green("connect succ!")
	return nil
}


func (b *Bitmex) InitConnection() (err error) {
	err = b.GetConnection()
	if err != nil {
		color.Red("Unable to connect to Websocket. Error: %s\n", err)
		return
	}

	b.heart = NewHeart()
	return
}


func (c *Bitmex) ListenHeart() {
	defer c.heart.timer.Stop()
	for {
		select {
		case <-c.doneSignal:
			c.Done("heart.")
			return
		case <-c.heart.timer.C:
			c.heart.timer.Reset(TFREQ)
			c.Write("ping")
			c.heart.cnt++
			log.Println("Send ping times:", c.heart.cnt)
			if c.heart.cnt > 3 {
				c.err = errors.New("Webscoket connection timeout.")
				c.Done("heart.")
				return
			}
		}
	}
}

func (c *Bitmex) Done(id string) {
	c.doneSignal <- true
	log.Println("Close Websocket ", id)
}

func (c *Bitmex) Write(in interface{}) {
	c.writeSignal <- in
}

func (b *Bitmex) WebsocketSubscribe(args []string) error {
	request := make(map[string]interface{})
	request["op"] = "subscribe"
	request["args"] = args

	return b.WebsocketSend(request)
}

func (b *Bitmex) Subscribe() {
	//args := []string{}
	for _, y := range b.Configuration.EnabledPairs {
		//args = append(args, y)
		b.WebsocketSubscribe([]string{y})
	}
}

func (b *Bitmex) WebsocketClient() {
	err := b.InitConnection()
	if err != nil {
		panic(err)
	}

	b.WebsocketPingHandler()
	b.Subscribe()
	for {
		msgType, resp, err := b.WebsocketConn.ReadMessage()
		if err != nil {
			color.Red("Unable to read from Websocket. Error: %s\n", err)
			continue
		}

		if msgType != websocket.TextMessage {
			continue
		}

		//fmt.Println(string(resp))
		depth := &Response{}
		err = common.JSONDecode(resp,depth)
		if err == io.EOF {
			b.err = errors.New("Webscoket connection break, error:" + err.Error())
			b.Done("read.")
		} else if err != nil {
			color.Red("read error--->%s", err.Error())
		}


		if depth.Table == "" {
			continue
		}


		DataBingo <- depth
		b.heart.Reset()

		//time.Sleep(time.Second * 1)
	}

	b.WebsocketConn.Close()
	log.Printf("Websocket client disconnected.\n")
}

func NewDataSignal() DataSignal {
	return make(DataSignal, RBUFFER)
}

func deal() {
	for {
		select {
		case in := <-DataBingo:
			if in.Data == nil {
				fmt.Println("indata is null")
			} else if in.Action == "partial" {
				color.Blue("action is %s , pass....", in.Action)

			} else if in.Table == "instrument" {
				instrument(in.Data)
			} else {
				color.Blue("table is %s, action is %s ↓↓↓↓↓↓", in.Table, in.Action)
				dataArr := in.Data
				for _, data := range dataArr  {
					if data.Side == "Sell" {
						color.Green("%s, [%s], 仓位:%v, 价格:%v",data.Symbol, getZhCn("side",data.Side),data.Size, data.Price)
					} else {
						color.Red("%s, [%s], 仓位:%v, 价格:%v",data.Symbol, getZhCn("side",data.Side),data.Size, data.Price)
					}

				}
			}
		}
	}
}


func instrument(dataArr []Keys) {
	for _, d := range dataArr {
		if d.ImpactMidPrice > 0 {
			index_price := -1.0
			if d.Symbol == "ADAU18" {
				index_price = ADA_INDEX_PRICE.Price
			} else if d.Symbol== "EOSU18" {
				index_price = EOS_INDEX_PRICE.Price
			} else {
				color.Red("fucking unknow symbol:%s", d.Symbol)
				continue
			}

			d1 := decimal.NewFromFloat(d.ImpactMidPrice)
			d2 := decimal.NewFromFloat(index_price)
			d3 := decimal.NewFromFloat(1)
			d4 := decimal.NewFromFloat(float64(30.0/365.0))
			ret := d1.Div(d2).Sub(d3).Div(d4)
			nowT := time.Now().Format("2006-01-02 15:04:05")
			color.Green("[%s],%s,合理基差年化率: %s", nowT, d.Symbol, ret.String())
		}
	}


}

func getZhCn(key,value string) string {
	if key == "side" {
		switch value {
		case "Sell":
			return "卖"
		case "Buy":
			return "买"
		}
	}

	return "🐩"
}