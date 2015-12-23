// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sentinel

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the sentinel client being tested.
	client *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server and a client that is configured to talk to
// each other. Tests should register handlers on mux which provides the mock
// responses for the API method being tested.
func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func urlPath(t *testing.T, routeName string, routeVars map[string]string) string {
	url, err := client.url(routeName, routeVars, nil)
	if err != nil {
		t.Fatalf("Error constructing URL path for route %q with vars %+v: %s", routeName, routeVars, err)
	}
	return "/" + url.Path
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		panic("writeJSON: " + err.Error())
	}
}

func testMethod(t *testing.T, r *http.Request, expect string) {
	if expect != r.Method {
		t.Errorf("Request method should have been %s, but it was %s", expect, r.Method)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	expect := url.Values{}
	for k, v := range values {
		expect.Add(k, v)
	}

	r.ParseForm()
	if !reflect.DeepEqual(expect, r.Form) {
		t.Errorf("Form values should have been %v, but it was %v", expect, r.Form)
	}
}

func testBody(t *testing.T, r *http.Request, expected string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Error("Unable to read body")
	}
	s := string(b)
	if expected != s {
		t.Errorf("Request body should have been %s, but it was %s", s, expected)
	}
}

func normaLize(t *time.Time) {
	*t = t.In(time.UTC)
}
