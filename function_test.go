package function

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestSaveTemperature(t *testing.T) {
	tests := []struct {
		data string
		expected string
	}{
		{data: "{}", expected: "Hello, {}!\n"},
	}

	for _, test := range tests {
		r, w, _ := os.Pipe()
		log.SetOutput(w)
		originalFlags := log.Flags()
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

		m := PubSubMessage{
			Data: []byte(test.data),
		}

		SaveTemperature(context.Background(), m)


		w.Close()

		log.SetOutput(os.Stderr)
		log.SetFlags(originalFlags)

		out, err := ioutil.ReadAll(r)

		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}
		if got := string(out); got != test.expected {
			t.Fatalf("Error: want: %q, actual: %q", test.expected, got)
		}
	}
}
