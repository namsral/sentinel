// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"net/http"
)

var (
	ErrInvalidAuthenticationCredentials = New("invalid_credentials", "missing or invalid authentication credentials", 401)
	ErrInvalidClient                    = New("invalid_client", "client authentication failed", 401)
	ErrNoAuthentionMethodIncluded       = ErrInvalidClient.Append("no authentication method was included")
	ErrUnsupportedAuthenticationMethod  = ErrInvalidClient.Append("included authentication method is not supported")
	ErrInvalidAuthenticationToken       = ErrInvalidClient.Append("authentication token was invalid")
	ErrUnknownClient                    = ErrInvalidClient.Append("unknown client")

	ErrUnauthorizedClient   = New("unauthorized_client", "client is not authorized", 403)
	ErrUnsupportedMediatype = New("unsupported_mediatype", "provided mediatype is not supported", 415)

	ErrInvalidToken    = New("invalid_token", "invalid JSON Web Token", 422)
	ErrInvalidRequest  = New("invalid_request", "", 422)
	ErrInvalidEmail    = ErrInvalidRequest.Append(`email parameter must match regex '[^@\s]+@[^@\s]+'`)
	ErrInvalidPassword = ErrInvalidRequest.Append(`password parameter must be at least 8 characters`)
	ErrEmailRegistered = ErrInvalidRequest.Append(`email already registered`)

	ErrConfilt       = New("conflict", "", 409)
	ErrNotAcceptable = New("not_acceptable", "resource not available in the requested mediatype.", 406)
	ErrNotFound      = New("not_found", "resource not found", 404)
	ErrServerError   = New("server_error", "unknown server error", 500)
)

// Error type implemented by the HTTP handlers
type Error struct {
	Name       string `json:"error"`
	Desc       string `json:"error_description,omitempty"`
	StatusCode int    `json:"-"`
}

func (e Error) Error() string {
	return e.Desc
}

// New return a new Error with the given desc and code.
func New(name, desc string, code int) Error {
	return Error{
		Name:       name,
		Desc:       desc,
		StatusCode: code,
	}
}

// Append appends the given desc to the error message.
func (e Error) Append(desc string) Error {
	switch e.Desc {
	case "":
		e.Desc = desc
	default:
		e.Desc = e.Desc + "; " + desc
	}
	return e
}

// WriteError marshals the given error and writes it to the given
// ResponseWriter.
func WriteError(w http.ResponseWriter, e Error) {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if e.StatusCode == 401 {
		SetUnauthenticateHeader(w, e)
	}
	w.WriteHeader(e.StatusCode)
	w.Write(data)
}

// SetUnauthenticateHeader formats the given error in a authentication error
// and writes it to the ResponseWriter.
func SetUnauthenticateHeader(w http.ResponseWriter, e Error) {
	h := AuthenticationScheme + ` realm="` + AuthenticationRealm + `", error="` + e.Name + `", error_description="` + e.Desc + `"`
	w.Header().Set("WWW-Authenticate", h)
}

// ParseError unmarshals the given data and return an Error.
func ParseError(data []byte) (*Error, error) {
	var v *Error
	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}
	return v, nil
}
