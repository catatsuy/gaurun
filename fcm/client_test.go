package fcm_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	. "github.com/mercari/gaurun/fcm"
)

func TestSend_Success(t *testing.T) {
	muxAPI := http.NewServeMux()
	testAPIServer := httptest.NewServer(muxAPI)
	defer testAPIServer.Close()

	payload := &Payload{
		Message: Message{
			Name: "aaaa",
		},
	}

	muxAPI.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		expectedType := "application/json"
		if contentType != expectedType {
			t.Errorf("Content-Type: expected %s, got %s", expectedType, contentType)
		}

		authorization := r.Header.Get("Authorization")
		expectedAuth := "Bearer token"
		if authorization != expectedAuth {
			t.Errorf("Authorization: expected %s, got %s", expectedAuth, authorization)
		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		defer r.Body.Close()

		actualBody := &Payload{}
		err = json.Unmarshal(bodyBytes, actualBody)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(actualBody, payload) {
			t.Errorf("expected %+v, got %+v", actualBody, payload)
		}

		http.ServeFile(w, r, "testdata/send_ok.json")
	})

	c := &Client{
		URL:        testAPIServer.URL,
		HTTPClient: http.DefaultClient,
		APIKey:     "token",
	}

	msg, err := c.Send(context.Background(), payload)
	if err != nil {
		t.Fatal(err)
	}

	expectedMsg := &Message{
		Name: "aaaa",
	}

	if !reflect.DeepEqual(msg, expectedMsg) {
		t.Errorf("expected %+v, got %+v", expectedMsg, msg)
	}
}
