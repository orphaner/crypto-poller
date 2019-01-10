package main

import (
	"github.com/influxdata/influxdb/client/v2"
	"github.com/namsral/flag"
	"log"
	"fmt"
	"time"
	"net/http"
	"encoding/json"
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

	InfluxdbCryptoPoller interface {
		CreatePoints(bitcoinRate chan CurrentPrice)
	}
	InfluxdbCryptoPollerImpl struct {
		username string
		password string
		url      string
		database string
	}
)

var ch chan CurrentPrice
var influxdbClient InfluxdbCryptoPoller
const uri = "https://api.coindesk.com/v1/bpi/currentprice.json"

func (poller *InfluxdbCryptoPollerImpl) CreatePoints(ch chan CurrentPrice) {

	// Create a new HTTPClient
	influxdbClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     poller.url,
		Username: poller.username,
		Password: poller.password,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer influxdbClient.Close()

	// Create a new point batch
	batchPoints, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  poller.database,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}


	bitcoinPrice := <-ch
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

func handler(w http.ResponseWriter, r *http.Request) {
	ch := make(chan CurrentPrice)
	go influxdbClient.CreatePoints(ch)

	var currentPrice CurrentPrice
	client := NewHttpClient()
	if err:= client.JSON(uri, &currentPrice) ; err != nil {
		log.Fatal(err)
	}
	ch <- currentPrice

	js, err := json.Marshal(currentPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", "pong")
}

func main() {
	username := flag.String("influxdb-username", "username", "Username to connect to influxdb")
	password := flag.String("influxdb-password", "password", "Password to connect to influxdb")
	url := flag.String("influxdb-url", "http://localhost:8086", "URL to connect to influxdb")
	database := flag.String("influxdb-database", "crypto", "Database to use in influxdb")
	flag.Parse()

	influxdbClient = &InfluxdbCryptoPollerImpl{
		username: *username,
		password: *password,
		url:      *url,
		database: *database,
	}
	log.Printf("%+v\n", influxdbClient)

	http.HandleFunc("/", handler)
	http.HandleFunc("/ping", ping)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
