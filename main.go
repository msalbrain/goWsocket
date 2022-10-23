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
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type Msg struct {
	Msg string
}

type CintrData struct {
	AuthorIDFromFromMerchant string `json:"authorIDFromMerchant"`
	AuthorName               string `json:"authorName"`
	Product                  string `json:"product"`
	Rating                   int    `json:"rating"`
	CreatedAt                string `json:"createdAt"`
	Text                     string `json:"text"`
	Title                    string `json:"title"`
}

type Sentiment struct {
	Neg      float32
	Neu      float32
	Pos      float32
	Compound float32
}

type Word struct {
	Word       string
	Total      int
	Review_ids []string
}

type WSocketBase struct {
	Product  string `json:"product"`
	Bert     string `json:"bert"`
	Pegasus  string `json:"pegasus"`
	Textrank string `json:"textrank"`
	// Overall_sentiment map[string]interface{}
	NLP_rating     float64     `json:"NLP_rating"`
	Average_rating float64     `json:"Average_rating"`
	Count          int         `json:"count"`
	Limit          int         `json:"limit"`
	Skip           int         `json:"skip"`
	Next           interface{} `json:"next"`
	Prev           interface{} `json:"prev"`
	Keyword        []Word      `json:"keyword"`
	Adj            []Word      `json:"adj"`
	Verb           []Word      `json:"verb"`
	// Data              []map[string]interface{}
}

type WSocketReturn struct {
	WSocketBase
	Data              []CintrData            `json:"data"`
	Overall_sentiment map[string]interface{} `json:"overall_sentiment"`
	Status            int                    `json:"status"`
}

type CintrResult struct {
	Count  int
	Data   []CintrData
	Status int
}

type PartialError struct {
	Error string
}

type PrivateApi struct {
	Msg    string
	Detail PartialError
	Status int
}

type DataInfo struct {
	ProdId string
	Skip   int
	Limit  int
}
type DataInput struct {
	ProductId string `json:"productId"`
	Skip      int    `json:"skip"`
	Limit     int    `json:"limit"`
}

func getCintrData(prodId string, limit int, skip int) CintrResult {
	url := fmt.Sprintf("https://cintr.herokuapp.com/api/reviews?productId=%s&limit=%d&skip=%d", prodId, limit, skip)
	response, err := http.Get(url)

	if err != nil {
		fmt.Print(err.Error())
		// os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	var d CintrResult
	err = json.Unmarshal(responseData, &d)

	return d
}

type SprocessReturn struct {
	ws  WSocketReturn
	pri PrivateApi
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

func startProcess(prodId string, limit int, skip int) SprocessReturn {

	//NOTE: The spelling below isn't analysis
	is, anlysis := confirmInDb(prodId, limit, skip)

	if is == true {
		jsonStr, err := json.Marshal(anlysis)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(jsonStr)

		var cintrcompose WSocketBase

		if err := json.Unmarshal(jsonStr, &cintrcompose); err != nil {
			panic(err)
		}
		// fmt.Println(cintrcompose)
		if limit == 0 {
			limit = cintrcompose.Count
		}
		g := getCintrData(prodId, limit, skip)

		var cintret WSocketReturn = WSocketReturn{cintrcompose, g.Data, nil, 200}

		return SprocessReturn{ws: cintret, pri: PrivateApi{}}

	} else {
		link := fmt.Sprintf("http://3.235.109.178/private/api/reviews/?productId=%s&limit=%v&skip=%v&access_token=85e53c1f044bb27455557fd3cdf",
			prodId, limit, skip)
		response, err := http.Get(link)
		if err != nil {
			log.Println(err.Error())
		}
		responseData, err := ioutil.ReadAll(response.Body)
		fmt.Println(string(responseData))
		var d PrivateApi
		err = json.Unmarshal(responseData, &d)
		d.Status = response.StatusCode
		return SprocessReturn{ws: WSocketReturn{}, pri: d}
	}

}

func WSocketReply(c *websocket.Conn, Val DataInfo) interface{} {
	s := startProcess(Val.ProdId, Val.Limit, Val.Skip)
	if (s.pri == PrivateApi{}) && (s.ws.Count == 0) {
		err := c.WriteJSON(map[string]interface{}{"error": "internal error", "status": 500})
		if err != nil {
			log.Println("write:", err)

		}
		return nil
	}
	if s.ws.Count != 0 {
		err := c.WriteJSON(s.ws)
		if err != nil {
			log.Println("write:", err)

		}
		return nil
	} else {
		if (s.pri != PrivateApi{}) {
			if (s.pri.Status != 200) || (s.pri.Status != 403) {
				err := c.WriteJSON(map[string]interface{}{"error": s.pri.Detail.Error, "status": s.pri.Status})
				if err != nil {
					log.Println("write:", err)
				}
				return nil
			} else if s.pri.Status == 403 {
				err := c.WriteJSON(map[string]interface{}{"detail": "internal error", "status": s.pri.Status})
				if err != nil {
					log.Println("write:", err)
				}
				return nil
			} else if s.pri.Status == 200 {
				err := c.WriteJSON(map[string]interface{}{"msg": s.pri.Msg, "status": s.pri.Status})
				if err != nil {
					log.Println("write:", err)
				}
			}
		}
		for j := 0; j < 300; j++ {
			time.Sleep(6000 * time.Millisecond)
			err := c.WriteJSON(map[string]interface{}{"msg": "processing in progress",
				"status": s.pri.Status})
			if err != nil {
				log.Println("write:", err)
			}
			if is, analysis := confirmInDb(Val.ProdId, Val.Limit, Val.Skip); is == true {
				st := startProcess(Val.ProdId, Val.Limit, Val.Skip)
				err := c.WriteJSON(st.ws)
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
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()
	for {
		// var V map[string]interface{}
		var V DataInput
		err := c.ReadJSON(&V)

		fmt.Println(V)
		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("recv: \nproduct: %v, limit: %v, skip: %v", V.ProductId,
			V.Limit, V.Skip)



		WSocketReply(c, DataInfo{ProdId: V.ProductId,
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
	flag.Parse()
	log.Println(*addr)
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
	// c := startProcess("63179a3bc987724acf03ddc8", 4, 4)
	// fmt.Println(c)
	// getCintrData("63179a3bc987724acf03ddc8", 4, 4)

	// fmt.Println(startProcess("63179a3bc987724acf03ddc8", 2, 0))
}
