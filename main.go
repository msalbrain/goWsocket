package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
	// "os"
	"fmt"
	"io/ioutil"

	"github.com/gorilla/websocket"
	"github.com/msalbrain/goWsocket.git/Db"
	"github.com/msalbrain/goWsocket.git/conf"
	Ws "github.com/msalbrain/goWsocket.git/types"
)




var addr = flag.String("addr", "0.0.0.0:9001", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func getCintrData(prodId string, limit int, skip int) Ws.CintrResult {
	configure, err := conf.NewConfig("/home/ubuntu/test/goWsocket/config.yaml")
	if err != nil{
		log.Fatal(err)	
	}
	url := fmt.Sprintf(configure.Server.DataLink, prodId, limit, skip)
	response, err := http.Get(url)

	if err != nil {
		fmt.Print(err.Error())
		// os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	var d Ws.CintrResult
	err = json.Unmarshal(responseData, &d)

	return d
}

func confirmInDb(prodId string, limit int, skip int) (bool, map[string]interface{}) {
	res := Db.CheckDb(prodId)
	is := false
	var anlysis map[string]interface{}
	for _, doc := range res {
		ma := doc.Map()
		if int32(ma["limit"].(int32)) == int32(limit) && int32(ma["skip"].(int32)) == int32(skip) {
			fmt.Printf("product: %s, limit: %d, skip: %d", prodId, ma["limit"], ma["skip"])
			fmt.Print("\n")
			is = true
			anlysis = ma
			break
		}

	}
	return is, anlysis
}

func startProcess(prodId string, limit int, skip int) Ws.SprocessReturn {
	configure, err := conf.NewConfig("/home/ubuntu/test/goWsocket/config.yaml")
	if err != nil{
		// log.Fatal("this is start process")
		log.Fatal(err)	
	}
	//NOTE: The spelling below isn't analysis
	is, anlysis := confirmInDb(prodId, limit, skip)

	if is == true {
		jsonStr, err := json.Marshal(anlysis)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(jsonStr)

		var cintrcompose Ws.WSocketBase

		if err := json.Unmarshal(jsonStr, &cintrcompose); err != nil {
			panic(err)
		}
		// fmt.Println(cintrcompose)
		if limit == 0 {
			limit = cintrcompose.Count
		}
		g := getCintrData(prodId, limit, skip)

		var cintret Ws.WSocketReturn = Ws.WSocketReturn{WSocketBase: cintrcompose, Data: g.Data, Overall_sentiment: nil, Status: 200}

		return Ws.SprocessReturn{Ws: cintret, Pri: Ws.PrivateApi{}}

	} else {
		link := fmt.Sprintf(configure.Server.PrivateLink,
			prodId, limit, skip)
		response, err := http.Get(link)
		if err != nil {
			log.Println(err.Error())
		}
		responseData, err := ioutil.ReadAll(response.Body)
		fmt.Printf("pass by private with productId '%s', limit '%v', skip '%v'\n", prodId, limit, skip)
		// fmt.Println(string(responseData))

		var d Ws.PrivateApi
		err = json.Unmarshal(responseData, &d)
		d.Status = response.StatusCode
		return Ws.SprocessReturn{Ws: Ws.WSocketReturn{}, Pri: d}
	}

}

func WSocketReply(c *websocket.Conn, Val Ws.DataInfo) interface{} {
	s := startProcess(Val.ProdId, Val.Limit, Val.Skip)
	if (s.Pri == Ws.PrivateApi{}) && (s.Ws.Count == 0) {
		err := c.WriteJSON(map[string]interface{}{"error": "internal error", "status": 500})
		if err != nil {
			log.Println("write:", err)

		}
		return nil
	}
	if s.Ws.Count != 0 {
		err := c.WriteJSON(s.Ws)
		if err != nil {
			log.Println("write:", err)

		}
		return nil
	} else {
		if (s.Pri != Ws.PrivateApi{}) {
			if (s.Pri.Status != 403) && (s.Pri.Status != 200) {
				err := c.WriteJSON(map[string]interface{}{"error": s.Pri.Detail.Error, "status": s.Pri.Status})
				if err != nil {
					log.Println("write:", err)
				}
				return nil

			} else if s.Pri.Status == 403 {
				err := c.WriteJSON(map[string]interface{}{"detail": "internal error", "status": s.Pri.Status})
				if err != nil {
					log.Println("write:", err)
				}
				return nil
			} else if s.Pri.Status == 200 {
				err := c.WriteJSON(map[string]interface{}{"msg": s.Pri.Msg, "status": s.Pri.Status})
				if err != nil {
					log.Println("write:", err)
				}
			}
		}
		for j := 0; j < 300; j++ {
			time.Sleep(6 * time.Second)
			err := c.WriteJSON(map[string]interface{}{"msg": "processing in progress",
				"status": s.Pri.Status})
			if err != nil {
				log.Println("write:", err)
			}
			if is, analysis := confirmInDb(Val.ProdId, Val.Limit, Val.Skip); is == true {
				st := startProcess(Val.ProdId, Val.Limit, Val.Skip)
				err := c.WriteJSON(st.Ws)
				if err != nil {
					if analysis == nil {
					}
					log.Println("write:", err)
				}
				return nil
			}
		}
	}
	return nil
}

func echo(w http.ResponseWriter, r *http.Request) {
	configure, err := conf.NewConfig("/home/ubuntu/test/goWsocket/config.yaml")
	if err != nil{
		// log.Fatal("this is main")
		log.Fatal(err)	
	}
	auth := r.Header.Get("access_token")
	// var b bool
	// if auth != configure.Server.AccessToken{
	// 	b = true

	// 	// c.Close()
	// }
	c, err := upgrader.Upgrade(w, r, nil)
	if true{
		c.WriteJSON(map[string]interface{}{"error": fmt.Sprintf("couldn't validate token, auth: %s, access_t: %s, %v", auth, configure.Server.AccessToken, r.Header)})
	}
	if err != nil {
		log.Print("upgrade:", err)
		return
	} else {
		log.Print("upgrade successfull, new connection made")
	}

	defer c.Close()
	for {
		// var V map[string]interface{}
		var V Ws.DataInput
		err := c.ReadJSON(&V)

		
		fmt.Println(V)
		if err != nil {
			log.Println("read:", err)
			break
		}

		if V.Limit > 50 {
			V.Limit = 50
		}

		log.Printf("recv: \nproduct: %v, limit: %v, skip: %v", V.ProductId,
			V.Limit, V.Skip)

		WSocketReply(c, Ws.DataInfo{ProdId: V.ProductId,
			Skip: V.Skip, Limit: V.Limit})

		// log.Printf("recv: \nproduct: %v, limit: %v, skip: %v", V["productId"].(string),
		// 				V["skip"].(int), V["productId"].(int))

		// WSocketReply(c, DataInfo{ProdId: V["productId"].(string),
		// 	Skip: V["skip"].(int), Limit: V["limit"].(int)})
		// err = c.WriteJSON(WSocketReturn{})

		// err = c.WriteMessage(websocket.TextMessage, jsonStr)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	configure, err := conf.NewConfig("/home/ubuntu/test/goWsocket/config.yaml")
	if err != nil{
		// log.Fatal("this is main")
		log.Fatal(err)	
	}
	flag.Parse()
	log.Println(*addr)
	log.SetFlags(0)
	http.HandleFunc(configure.Server.Route, echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
	// c := startProcess("63179a3bc987724acf03ddc8", 4, 4)
	// fmt.Println(c)
	// getCintrData("63179a3bc987724acf03ddc8", 4, 4)

	// fmt.Println(startProcess("63179a3bc987724acf03ddc8", 2, 0))
}
