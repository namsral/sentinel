// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sentinel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ErrorResponse struct {
	Response *http.Response `json:",omitempty"`
	Name     string         `json:"error"`
	Desc     string         `json:"error_description"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %s, %s",
		r.Response.Request.Method,
		r.Response.Request.URL,
		r.Response.StatusCode,
		http.StatusText(r.Response.StatusCode),
		r.Name,
	)
}

func (r *ErrorResponse) HTTPStatusCode() int { return r.Response.StatusCode }

func CheckResponse(r *http.Response) error {
	if r.StatusCode/100 == 2 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}
