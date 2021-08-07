// Package p contains an HTTP Cloud Function.
package function

import (
	"context"
	"log"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

// HelloWorld prints the JSON encoded "message" field in the body
// of the request or "Hello, World!" if there isn't one.
func SaveTemperature(ctx context.Context, m PubSubMessage) error {
	name := string(m.Data) // Automatically decoded from base64.
	if name == "" {
		name = "World"
	}
	log.Printf("Hello, %s!", name)
	return nil
}
