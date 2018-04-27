package main

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"log"
	"net/http"
	"time"
)

const (
	username    = "user"
	password    = "password"
	influxdbUrl = "http://localhost:8086"
	MyDB        = "crypto"
)

type (
	CurrentPrice struct {
		PriceByCurrency map[string]Price `json:"bpi"`
	}
	Price struct {
		Code        string  `json:"code"`
		Symbol      string  `json:"symbol"`
		Rate        string  `json:"rate"`
		Description string  `json:"description"`
		RateFloat   float32 `json:"rate_float"`
	}

	CoinDeskClient interface {
		PullCurrentPrice() *CurrentPrice
	}
	CoinDeskClientImpl struct {
	}

	InfluxdbCryptoPoller interface {
		CreatePoints(bitcoinRate CurrentPrice)
	}
	InfluxdbCryptoPollerImpl struct {
	}
)

func (cdClient *CoinDeskClientImpl) PullCurrentPrice() *CurrentPrice {
	var uri = "https://api.coindesk.com/v1/bpi/currentprice.json"
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	var currentPrice CurrentPrice
	json.NewDecoder(response.Body).Decode(&currentPrice)

	return &currentPrice
}

func (poller *InfluxdbCryptoPollerImpl) CreatePoints(bitcoinPrice CurrentPrice) {

	// Create a new HTTPClient
	influxdbClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxdbUrl,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer influxdbClient.Close()

	// Create a new point batch
	batchPoints, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  MyDB,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	for currency, value := range bitcoinPrice.PriceByCurrency {

		// Create a point and add to batch
		tags := map[string]string{"currency": currency}
		fields := map[string]interface{}{
			"rate": value.RateFloat,
		}

		point, err := client.NewPoint("bitcoin_price", tags, fields, time.Now())
		if err != nil {
			log.Fatal(err)
		}
		batchPoints.AddPoint(point)
	}

	// Write the batch
	if err := influxdbClient.Write(batchPoints); err != nil {
		log.Fatal(err)
	}

	// Close client resources
	if err := influxdbClient.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var coinDeskClient CoinDeskClientImpl
	var price = coinDeskClient.PullCurrentPrice()
	fmt.Printf("%+v\n", price)

	var influxdbClient InfluxdbCryptoPollerImpl
	influxdbClient.CreatePoints(*price)
}
