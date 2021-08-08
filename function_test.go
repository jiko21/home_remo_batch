package function

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

//func TestSaveTemperature(t *testing.T) {
//	tests := []struct {
//		data     string
//		expected string
//	}{
//		{data: "{}", expected: "Hello, {}!\n"},
//	}
//
//	for _, test := range tests {
//		r, w, _ := os.Pipe()
//		log.SetOutput(w)
//		originalFlags := log.Flags()
//		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
//
//		m := PubSubMessage{
//			Data: []byte(test.data),
//		}
//
//		SaveTemperature(context.Background(), m)
//
//		w.Close()
//
//		log.SetOutput(os.Stderr)
//		log.SetFlags(originalFlags)
//
//		out, err := ioutil.ReadAll(r)
//
//		if err != nil {
//			t.Fatalf("ReadAll: %v", err)
//		}
//		if got := string(out); got != test.expected {
//			t.Fatalf("Error: want: %q, actual: %q", test.expected, got)
//		}
//	}
//}

func TestGetTemperature(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer not true" {
			w.WriteHeader(403)
		}
		io.WriteString(w, `[
  {
    "name": "Remo mini",
	"id": "1",
    "newest_events": {
      "te": {
        "val": 28.5,
        "created_at": "2021-08-07T15:34:27Z"
      }
    }
  }
]`)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()
	type Args struct {
		url   string
		token string
	}
	tests := []struct {
		name    string
		args    Args
		want    []Response
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Correctly call api",
			args: Args{
				url:   server.URL,
				token: "true token",
			},
			want: []Response{
				{
					Name: "Remo mini",
					Id: "1",
					NewestEvents: Events{
						Te: SensorValue{
							Val:       28.5,
							CreatedAt: time.Date(2021, 8, 7, 15, 34, 27, 0, time.UTC),
						},
					},
				},
			},
			wantErr: false,
		},
		// TODO: Add test cases.
		{
			name: "Fail call when token is not valid",
			args: Args{
				url:   server.URL,
				token: "not true",
			},
			want: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTemperature(tt.args.url, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemperature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTemperature() got = %v, want %v", got, tt.want)
			}
		})
	}
}
