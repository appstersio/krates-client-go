package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kontena/terraform-provider-kontena/api"
	"github.com/stretchr/testify/assert"
)

const testAccessToken = "0123456789abcdef"

type test struct {
	mux    *http.ServeMux
	server *httptest.Server
	config Config
	client *Client
}

func makeTest() *test {
	var test = test{}

	test.mux = http.NewServeMux()
	test.server = httptest.NewServer(test.mux)
	test.config = Config{
		URL:         test.server.URL,
		AccessToken: testAccessToken,
	}

	if client, err := test.config.MakeClient(); err != nil {
		panic(err)
	} else {
		test.client = client
	}

	return &test
}

func TestPing(t *testing.T) {
	var test = makeTest()

	test.mux.HandleFunc("/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		var ping = api.Ping{Message: "pong"}
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ping)
	})

	if err := test.client.Ping(); err != nil {
		t.Errorf("ping error: %v", err)
	} else {
		t.Logf("ping ok")
	}
}

func TestPing404(t *testing.T) {
	var test = makeTest()

	test.mux.HandleFunc("/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	if err := test.client.Ping(); err == nil {
		t.Errorf("get error: %v", err)
	} else if err, ok := err.(NotFoundError); !ok {
		t.Errorf("get error: %v", err)
	} else {
		assert.EqualError(t, err, fmt.Sprintf("GET %v/v1/ping => HTTP 404 Not Found: <nil>", test.server.URL))
	}
}

func TestPing500(t *testing.T) {
	var test = makeTest()

	test.mux.HandleFunc("/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		var err = api.Error{Error: map[string]string{"test": "foo"}}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(err)
	})

	if err := test.client.Ping(); err == nil {
		t.Errorf("get error: %v", err)
	} else if err, ok := err.(Error); !ok {
		t.Errorf("get error type %T: %v", err, err)
	} else {
		assert.EqualError(t, err, fmt.Sprintf("GET %v/v1/ping => HTTP 500 Internal Server Error: map[test:foo]", test.server.URL))
	}
}