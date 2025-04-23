package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gocolly/colly"
)

// url parameters breakdown:
// https://www.skyscanner.es/transporte/vuelos/{origin_code}/{destination_code}/?adultsv2={num_adults}&cabinclass={class}&childrenv2={num_children}&ref={reference}&rtn=0&preferdirects=false&outboundaltsenabled=false&inboundaltsenabled=false&oym=2603&selectedoday=01

// url full example:
// "https://www.skyscanner.es/transporte/vuelos/mad/hnd/?adultsv2=1&cabinclass=economy&childrenv2=&ref=home&rtn=0&preferdirects=false&outboundaltsenabled=false&inboundaltsenabled=false&oym={yymm}&selectedoday=01"

const DOMAIN string = "www.skyscanner.es"
const URL_PATTERN string = "https://www.skyscanner.es/transporte/vuelos/mad/hnd/?adultsv2=1&cabinclass=economy&childrenv2=&ref=home&rtn=0&preferdirects=false&outboundaltsenabled=false&inboundaltsenabled=false&oym=%s&selectedoday=01"

type FlightData struct {
	origin      string
	destination string
	date        string
	price       string
}

func storeData(file *os.File, data []FlightData) {
	var writer *csv.Writer = csv.NewWriter(file)

	// Write csv headers
	headers := []string{
		"origin",
		"destination",
		"date",
		"price",
	}
	writer.Write(headers)

	// Write data
	for _, flight := range data {
		row := []string{
			flight.origin,
			flight.destination,
			flight.date,
			flight.price,
		}
		writer.Write(row)
	}
	defer writer.Flush()
}

func createFile(file_path string) *os.File {
	file, error := os.Create(file_path)
	if error != nil {
		panic(error)
	}
	return file
}

func main() {

	// Define CLI parameters
	var datePtr *string = flag.String("date", "0000", "Date in format yymm. E.g: 2503 = March 2025")

	// Parse CLI parameters
	flag.Parse()

	//Extract CLI parameters
	var dateValue string = *datePtr
	if dateValue == "0000" {
		panic("A proper date was not provided via CLI. Usage: go run sky-scanner.go  --date {date}")
	}

	// Format URL
	var urlFormatted string = fmt.Sprintf(URL_PATTERN, dateValue)

	// Instantiate colly's collector.
	// It handles the context of the webscraping.
	var collector *colly.Collector = colly.NewCollector(colly.AllowedDomains(DOMAIN))
	collector.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0"

	// Instanciate FlightData
	var data []FlightData

	// Detect when target HTML elements appear
	collector.OnHTML("button.month-view-calendar__cell", func(e *colly.HTMLElement) {
		var flight FlightData = FlightData{}
		flight.date = e.Attr("isodate")
		flight.origin = "Madrid"
		flight.destination = "Haneda"
		flight.price = e.ChildText(".price")

		data = append(data, flight)
	},
	)

	collector.OnResponse(func(r *colly.Response) {
		fmt.Println("Status Code:", r.StatusCode)
		fmt.Println("Response Body (HTML):")
		fmt.Println(string(r.Body)) // Print the raw HTML
	})

	collector.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Request URL: %s failed with error: %v", r.Request.URL, err)
	})

	// When finished crawling, write to csv
	collector.OnScraped(func(r *colly.Response) {
		file := createFile(fmt.Sprintf("data/%s.csv", time.Now().Format(time.DateOnly)))
		defer file.Close()

		storeData(file, data)
	})

	// Finally, visit page and let everything be handled by events :)
	err := collector.Visit(urlFormatted)
	if err != nil {
		panic(err)
	}
}
