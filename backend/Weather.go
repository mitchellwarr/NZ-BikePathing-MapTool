package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type weatherJSON struct {
	List []weatherInfo `json:"list"`
}

type weatherInfo struct {
	Coord struct {
		Lon int `json:"lon"`
		Lat int `json:"lat"`
	} `json:"coord"`
	Main struct {
		Humidity   float64 `json:"humidity"`
		Pressure   float64 `json:"pressure"`
		Grnd_level float64 `json:"grnd_level"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Time string `json:"dt_txt"`
}

func getWeatherInfo() []weatherInfo {
	weatherAPIKey := "dae1c8895ce433822ad90e922ec6adf9"
	// napier := "2186313"
	// hastings := "2190224"
	lat := "-39.4928"
	lon := "176.9120"
	// QueryEscape escapes the string so
	// it can be safely placed inside a URL query
	safeLat := url.QueryEscape(lat)
	safeLon := url.QueryEscape(lon)
	safeWeatherAPIKey := url.QueryEscape(weatherAPIKey)

	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?lat=%s&lon=%s&APPID=%s", safeLat, safeLon, safeWeatherAPIKey)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	CheckErr(err)

	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	CheckErr(err)

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	// Fill the record with the data from the JSON
	var record weatherJSON
	if err := json.Unmarshal([]byte(string(body)), &record); err != nil {
		log.Fatal(err)
	}

	fmt.Println("List length: ", len(record.List))
	for _, r := range record.List {
		fmt.Println("--")
		fmt.Println("Text: ", r.Time)
		fmt.Println("Wind deg: ", r.Wind.Deg)
		fmt.Println("Wind speed: ", r.Wind.Speed)
	}

	return record.List
}

func getWeatherFromTime(t time.Time, weather []weatherInfo) weatherInfo {
	if USE_TEST_WIND_DATA {
		newWind := weather[0]
		newWind.Wind.Deg = 360
		newWind.Wind.Speed = 6
		return newWind
	}
	for i := 1; i < len(weather); i++ {
		if inTimeSpan(makeTime(weather[i-1].Time), makeTime(weather[i].Time), t) {
			return weather[i-1]
		}
	}
	CheckErr(errors.New("Need more waether forecasts"))
	return weatherInfo{}
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func makeTime(jtime string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05 (MST)", jtime+" (NZST)")
	CheckErr(err)
	return t
}
