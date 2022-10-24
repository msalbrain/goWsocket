package types

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
	Keyword        interface{} `json:"keyword"`
	Adj            interface{} `json:"adj"`
	Verb           interface{} `json:"verb"`
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

type SprocessReturn struct {
	Ws  WSocketReturn
	Pri PrivateApi
}
