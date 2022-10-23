package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	// "os"
	"fmt"
	"io/ioutil"

	"github.com/gorilla/websocket"
	"github.com/msalbrain/goWsocket.git/Db"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

type Msg struct {
	msg string
}

type CintrData struct {
	AuthorIDFromFromMerchant string
	AuthorName               string
	Product                  string
	Rating                   int
	CreatedAt                string
	Text                     string
	Title                    string
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
	Product           string
	Bert              string
	Pegasus           string
	Textrank          string
	// Overall_sentiment map[string]interface{}
	NLP_rating        float64
	Average_rating    float64
	Count             int
	Limit             int
	Skip              int
	Next              interface{}
	Prev              interface{}
	Keyword           []Word
	Adj               []Word
	Verb              []Word
	// Data              []map[string]interface{}
}

type WSocketReturn struct {
	WSocketBase
	Data              []map[string]interface{}
	Overall_sentiment map[string]interface{}
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

func startProcess(prodId string, limit int, skip int) SprocessReturn {

	res := Db.CheckDb(prodId)
	is := false
	var anlysis map[string]interface{}
	for _, doc := range res {
		ma := doc.Map()
		fmt.Printf("product: %s, limit: %d, skip: %d", prodId, ma["limit"], ma["skip"])
		fmt.Print("\n")
		if int32(ma["limit"].(int32)) == int32(limit) && int32(ma["skip"].(int32)) == int32(skip) {
			fmt.Printf("product: %s, limit: %d, skip: %d", prodId, ma["limit"], ma["skip"])
			fmt.Print("\n")
			is = true
			anlysis = ma
			break
		}

	}

	if is == true {
		jsonStr, err := json.Marshal(anlysis)
		if err != nil{
			log.Fatal(err)
		}
		// fmt.Println(jsonStr)

		var cintrcompose WSocketBase
		
		if err := json.Unmarshal(jsonStr, &cintrcompose); err != nil {
			panic(err)
		}
		// fmt.Println(cintrcompose)



		// cintrData := getCintrData(prodId, limit, skip)
		// cintrcompose.Data = cintrData.Data
		// cintrcompose.Count = cintrData.Count
		// cintrcompose.Product = anlysis["product"].(string)
		// cintrcompose.Bert = anlysis["bert"].(string)
		// cintrcompose.Pegasus = anlysis["pegasus"].(string)
		// cintrcompose.Textrank = anlysis["textrank"].(string)
		// cintrcompose.NLP_rating = float64(anlysis["NLP_rating"].(float64))
		// cintrcompose.Average_rating = float64(anlysis["Average_rating"].(float64))
		// if limit == 0 {
		// 	cintrcompose.Limit = cintrData.Count
		// } else {
		// 	cintrcompose.Limit = limit
		// }
		// cintrcompose.Skip = anlysis["skip"].(int)

		return SprocessReturn{ws: cintrcompose}

	} else {
		// link := fmt.Sprintf("http://3.235.109.178/private/api/reviews/?productId=%s&limit=%v&skip=%v&access_token=85e53c1f044bb27455557fd3cdf",
		// 	prodId, limit, skip)
		// response, err := http.Get(link)
		// if err != nil {
		// 	log.Println(err.Error())
		// }
		// responseData, err := ioutil.ReadAll(response.Body)
		// fmt.Println(string(responseData))
		// var d PrivateApi
		// err = json.Unmarshal(responseData, &d)
		// d.Status = response.StatusCode
		// return SprocessReturn{pri: d}
	}
	return SprocessReturn{}

}

func WSocketReply(c *websocket.Conn, Val DataInfo) {

}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()
	for {
		var V map[string]interface{}
		err := c.ReadJSON(&V)

		if err != nil {
			log.Println("read:", err)
			break
		}

		m := Msg{msg: V["msg"].(string)}

		log.Printf("recv: %v", m.msg)

		err = c.WriteJSON(map[string]interface{}{"said": "fuck you", "age": 4})
		// err = c.WriteMessage(mt, jsonMap)
		// err = c.WriteMessage(websocket.TextMessage, stringJSON)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	// flag.Parse()
	// log.SetFlags(0)
	// http.HandleFunc("/echo", echo)
	// log.Fatal(http.ListenAndServe(*addr, nil))
	// c := startProcess("63179a3bc987724acf03ddc8", 4, 4)
	// fmt.Println(c)
	// getCintrData("63179a3bc987724acf03ddc8", 4, 4)

	fmt.Println(startProcess("63179a3bc987724acf03ddc8", 2, 0))
}
