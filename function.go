// Package p contains an HTTP Cloud Function.
package function

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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

var (
	dbUser                 = os.Getenv("DB_USER")
	dbPwd                  = os.Getenv("DB_PASS")
	instanceConnectionName = os.Getenv("INSTANCE_CONNECTION_NAME")
	dbName                 = os.Getenv("DB_NAME")
)

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

func SaveData(db *sql.DB, response Response) error {
	defer db.Close()
	ins, err := db.Prepare("INSERT INTO temperature(remo_id, measured_at, value) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = ins.Exec(response.Id, response.NewestEvents.Te.CreatedAt, response.NewestEvents.Te.Val)
	return err
}

/**
 * entry point of cloud functions
 */
func SaveTemperature(ctx context.Context, m PubSubMessage) error {
	body, err := GetTemperature(apiUrl, accessToken)
	if err != nil {
		log.Panicf("Error: %v", err)
	}
	log.Printf("correctly get, %v", body[0])
	socketDir, isSet := os.LookupEnv("DB_SOCKET_DIR")
	if !isSet {
		socketDir = "/cloudsql"
	}
	uri := fmt.Sprintf("%s:%s@unix(/%s/%s)/%s?parseTime=true", dbUser, dbPwd, socketDir, instanceConnectionName, dbName)
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return err
	}
	err = SaveData(db, body[0])
	if err != nil {
		log.Panicf("Error: %v", err)
	}
	log.Printf("correctly saved, %v", body[0])
	return nil
}
