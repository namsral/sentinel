// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"sentinel"
	"sentinel/datastore"
)

var (
	serveMux   = http.NewServeMux()
	httpClient = http.Client{Transport: (*muxTransport)(serveMux)}
	apiClient  = sentinel.NewClient(&httpClient)
)

func init() {
	serveMux.Handle("/", Handler())
}

func setup() {
	store = datastore.NewMockDatastore()
}

type muxTransport http.ServeMux

// Roundtrip is a custom http.RounTripper for test API requests/responses. It
// intercepts all HTTP traffic and serves a local reponse instead of dialing
// out.
func (t *muxTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	w.Body = new(bytes.Buffer)
	(*http.ServeMux)(t).ServeHTTP(w, r)
	return &http.Response{
		StatusCode:    w.Code,
		Status:        http.StatusText(w.Code),
		Header:        w.HeaderMap,
		Body:          ioutil.NopCloser(w.Body),
		ContentLength: int64(w.Body.Len()),
		Request:       r,
	}, nil
}
