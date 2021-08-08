// Package p contains an HTTP Cloud Function.
package function

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

var accessToken = os.Getenv("REMO_API_TOKEN")
var apiUrl = "https://api.nature.global/1/devices"

type SensorValue struct {
	Val       float64   `json:"val"`
	CreatedAt time.Time `json:"created_at"`
}

type Events struct {
	Te SensorValue `json:"te"`
}

type Response struct {
	Name         string `json:"name"`
	Id           string `json:"id"`
	NewestEvents Events `json:"newest_events"`
}

/**
 * get temperature from nature remo api
 */
func GetTemperature(url, token string) ([]Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := new(http.Client)
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("response error(%d)", res.StatusCode))
	}

	byteArray, _ := ioutil.ReadAll(res.Body)
	body := new([]Response)
	err = json.Unmarshal(byteArray, body)
	return *body, err
}

/**
 * entry point of cloud functions
 */
func SaveTemperature(ctx context.Context, m PubSubMessage) error {
	body, err := GetTemperature(apiUrl, accessToken)
	if err != nil {
		log.Panicf("Error: %v", err)
	}
	log.Printf("Hello, %v", body[0])
	return nil
}
