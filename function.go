// Package p contains an HTTP Cloud Function.
package function

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

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

type Secret struct {
	RemoApiToken           string `json:"REMO_API_TOKEN"`
	DbUser                 string `json:"DB_USER"`
	DbPass                 string `json:"DB_PASS"`
	InstanceConnectionName string `json:"INSTANCE_CONNECTION_NAME"`
	DbName                 string `json:"DB_NAME"`
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
	// secretの取得
	secretName := os.Getenv("SECRET_NAME")
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Panicf("Error: %v", err)
	}
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		log.Panicf("Error: %v", err)
	}

	secrets := new(Secret)
	err = json.Unmarshal(result.Payload.Data, secrets)
	if err != nil {
		log.Panicf("Error: %v", err)
	}

	body, err := GetTemperature(apiUrl, secrets.RemoApiToken)
	if err != nil {
		log.Panicf("Error: %v", err)
	}
	log.Printf("correctly get, %v", body[0])
	socketDir, isSet := os.LookupEnv("DB_SOCKET_DIR")
	if !isSet {
		socketDir = "/cloudsql"
	}
	uri := fmt.Sprintf("%s:%s@unix(/%s/%s)/%s?parseTime=true", secrets.DbUser, secrets.DbPass, socketDir, secrets.InstanceConnectionName, secrets.DbName)
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
