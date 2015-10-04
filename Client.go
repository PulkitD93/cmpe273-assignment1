package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ResponseBuyingStocks struct {
	Result struct {
		Tradeid        int     `json:"TradeId"`
		Stocks         string  `json:"Stocks"`
		Unvestedamount float64 `json:"UnvestedAmount"`
	} `json:"result"`
	Error interface{} `json:"error"`
	ID    int         `json:"id"`
}

type ResponseCheckingPortfolio struct {
	Result struct {
		Stocks             string  `json:"Stocks"`
		Currentmarketvalue float64 `json:"CurrentMarketValue"`
		Unvestedamount     float64 `json:"UnvestedAmount"`
	} `json:"result"`
	Error interface{} `json:"error"`
	ID    int         `json:"id"`
}

func main() {

	var r1 ResponseBuyingStocks
	var r2 ResponseCheckingPortfolio

	url := "http://localhost:8080/rpc"

	if len(os.Args) == 3 {

		// check for percent validation
		s1 := strings.Replace(os.Args[1], "%", "", -1)

		s2 := strings.Split(s1, ",")

		var totalPercent int = 0

		for i := 0; i < len(s2); i++ {
			s3 := strings.Split(s2[i], ":")
			per, err := strconv.Atoi(s3[1])

			if err != nil {
				fmt.Println("Error in parsing percentage")
				return
			}
			totalPercent = totalPercent + per
		}

		if totalPercent > 100 {
			fmt.Println("Error : Total percentage is more than 100")
			return
		}

		var jsonStr string = "{\"method\":\"TradeStocks.BuyingStocks\",\"params\":[{\"stockSymbolAndPercentage\":\"" + os.Args[1] + "\",\"budget\":" + os.Args[2] + "}],\"id\":0}"

		req, err := http.NewRequest("POST", url, bytes.NewBufferString(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err.Error())
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println("response Body:", string(body))

		jsonRes := []byte(string(body))
		err2 := json.Unmarshal(jsonRes, &r1)

		if err2 != nil {
			fmt.Println("Errow while unmarshalling json response")
		}

		fmt.Printf("TradeId : " + strconv.Itoa(r1.Result.Tradeid) + "\n")
		fmt.Printf("Stocks : " + r1.Result.Stocks + "\n")
		fmt.Printf("Unvested Amount : " + strconv.FormatFloat(r1.Result.Unvestedamount, 'f', -1, 64) + "\n")

	} else if len(os.Args) == 2 {

		// tid, err := strconv.Atoi(os.Args[1])

		// if err != nil {
		// 	fmt.Println("Error in input Trade ID")
		// }

		var jsonStr string = "{\"method\":\"TradeStocks.CheckingPortfolio\",\"params\":[{\"TradeId\":" + os.Args[1] + "}],\"id\":0}"

		req, err := http.NewRequest("POST", url, bytes.NewBufferString(jsonStr))
		req.Header.Set("Content-Type", "application/json")


		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err.Error())
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		//fmt.Println("response Body:", string(body))

		jsonRes := []byte(string(body))
		err2 := json.Unmarshal(jsonRes, &r2)

		if err2 != nil {
			fmt.Println("Errow while unmarshalling json response")
		}

		fmt.Printf("Stocks : " + r2.Result.Stocks + "\n")
		fmt.Printf("CurrentMarketValue : " + strconv.FormatFloat(r2.Result.Currentmarketvalue, 'f', -1, 64) + "\n")
		fmt.Printf("Unvested Amount : " + strconv.FormatFloat(r2.Result.Unvestedamount, 'f', -1, 64) + "\n")

	}

}
