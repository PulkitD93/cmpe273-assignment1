package main

import (
	"encoding/csv"
	"fmt"
	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/justinas/alice"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var count int32 = 0
var remainAmt float32 = 0
var dataMap = make(map[int32]DataStruct)

type DataStruct struct {
	stocks         string
	unvestedAmount float32
}

type RequestBuyingStocks struct {
	StockSymbolAndPercentage string
	Budget                   float32
}

type ResponseBuyingStocks struct {
	TradeId        int32
	Stocks         string
	UnvestedAmount float32
}

type RequestCheckingPortfolio struct {
	TradeId int32
}

type ResponseCheckingPortfolio struct {
	Stocks             string
	CurrentMarketValue float32
	UnvestedAmount     float32
}

type TradeStocks struct{}

func (t *TradeStocks) BuyingStocks(r *http.Request, args *RequestBuyingStocks, reply *ResponseBuyingStocks) error {
	tradeId := genTradeId()
	reply.TradeId = tradeId
	remainAmt = 0

	
	inputReplace := strings.Replace(args.StockSymbolAndPercentage, "%", "", -1) //to remove % symbol
	stockBlock := strings.Split(inputReplace, ",")

	var stockDisplay string = ""
	var stockDisplayPerShare string = ""

	for i := 0; i < len(stockBlock); i++ {

		stockNamePercent := strings.Split(stockBlock[i], ":")
		stockPercent, err := strconv.ParseFloat(stockNamePercent[1], 32)

		if err != nil {
			fmt.Println("Parse error during conversion of stock percentage to float : " + err.Error())
		}

		stockPercent32 := float32(stockPercent)

		//get the number of stocks bought and share price
		nsb, persharePrice := buyStock(stockNamePercent[0], stockPercent32, args.Budget)
		outputStockValue := float32(nsb) * persharePrice
		stockDisplay = stockDisplay + stockNamePercent[0] + ":" + strconv.FormatInt(int64(nsb), 10) + ":$" + strconv.FormatFloat(float64(outputStockValue), 'f', 2, 32) + ","
		stockDisplayPerShare = stockDisplayPerShare + stockNamePercent[0] + ":" + strconv.FormatInt(int64(nsb), 10) + ":$" + strconv.FormatFloat(float64(persharePrice), 'f', 2, 32) + ","

	}

	stockDisplay = stockDisplay[0 : len(stockDisplay)-1]
	stockDisplayPerShare = stockDisplayPerShare[0 : len(stockDisplayPerShare)-1]
	data := DataStruct{stockDisplayPerShare, remainAmt}
	dataMap[tradeId] = data
	reply.Stocks = stockDisplay
	reply.UnvestedAmount = remainAmt
	return nil
}

func (t *TradeStocks) CheckingPortfolio(r *http.Request, args *RequestCheckingPortfolio, reply *ResponseCheckingPortfolio) error {

	var displaycurrentMarketVal float32 = 0
	var displayStockString string = ""
	var plSymbol string = ""

	data := dataMap[args.TradeId]

	//create a stock string to dislay
	storedStocksString := data.stocks
	stockBlock := strings.Split(storedStocksString, ",")

	for i := 0; i < len(stockBlock); i++ {

		stockNamePercent := strings.Split(stockBlock[i], ":")
		buyPrice, err := strconv.ParseFloat(strings.Trim(stockNamePercent[2], "$"), 32)
		if err != nil {
			fmt.Println("Parsing error to float : " + err.Error())
		}

		buyPrice32 := float32(buyPrice)   //converts to float32

		//parses the number of shares bought
		nShare, err := strconv.ParseInt(stockNamePercent[1], 10, 32)
		if err != nil {
			fmt.Println("Parse error to get number of shares : " + err.Error())
		}

		nShare32 := int32(nShare)

		currentPrice := getPrice(stockNamePercent[0])
		currentTotalPrice := currentPrice * float32(nShare32)
		displaycurrentMarketVal = displaycurrentMarketVal + currentTotalPrice

		if buyPrice32 < currentPrice {
			plSymbol = "+"
		} else if buyPrice32==currentPrice {
			plSymbol = ""
		} else{
			plSymbol="-"
		}

		displayStockString = displayStockString + stockNamePercent[0] + ":" + stockNamePercent[1] + ":" + plSymbol + "$" + strconv.FormatFloat(float64(currentTotalPrice), 'f', 2, 32) + ","
	}

	displayStockString = displayStockString[0 : len(displayStockString)-1]
	reply.UnvestedAmount = data.unvestedAmount
	reply.Stocks = displayStockString
	reply.CurrentMarketValue = displaycurrentMarketVal

	return nil
}

func main() {

	r := mux.NewRouter()

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")

	ts := new(TradeStocks)
	s.RegisterService(ts, "")

	chain := alice.New(
		func(h http.Handler) http.Handler {
			return handlers.CombinedLoggingHandler(os.Stdout, h)
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		})

	r.Handle("/rpc", chain.Then(s))
	log.Fatal(http.ListenAndServe(":8080", r))

}

func genTradeId() int32 {
	count = count + 1
	return count
}

func buyStock(companyCode string, companyPercent float32, totalBudget float32) (int32, float32) {
	companySharePrice := getPrice(companyCode)
	companyBudget := companyPercent / 100 * totalBudget

	var numberOfShares int32 = int32(companyBudget / companySharePrice)
	remainAmt = remainAmt + companyBudget - (float32(numberOfShares) * companySharePrice)
	return numberOfShares, companySharePrice

}

func getPrice(companyCode string) float32 {
	downloadFromUrl("http://finance.yahoo.com/d/quotes.csv?s=" + companyCode + "&f=sa")
	price := csvRead()
	return price
}

func downloadFromUrl(url string) {
	fileName := "quotes.csv"
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	io.Copy(output, response.Body)

}

func csvRead() float32 {

	csvfile, err := os.Open("quotes.csv")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer csvfile.Close()

	reader := csv.NewReader(csvfile)

	reader.FieldsPerRecord = -1

	rawCSVdata, err := reader.ReadAll()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var price32 float32

	for _, each := range rawCSVdata {
		price, err := strconv.ParseFloat(each[1], 32)
		if err != nil {
			fmt.Println("error during conversion of price")
			os.Exit(1)
		}

		price32 = float32(price)
	}

	return price32

}
