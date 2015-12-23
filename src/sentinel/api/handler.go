// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"log"
	"net/http"
	"net/url"

	"sentinel/datastore"
	"sentinel/router"

	"github.com/gorilla/mux"
)

var (
	store     = datastore.NewDatastore(nil)
	baseURL   *url.URL
	apiRouter = router.API(nil)
)

// SetbaseURL sets the given URL as the baseURL for the router.
func SetbaseURL(u *url.URL) {
	baseURL = u
	apiRouter = router.API(u)
}

// Handler returns a router with predefined handlers.
func Handler() *mux.Router {
	m := router.API(baseURL)
	m.Get(router.Signup).Handler(handler(serveSignup))
	m.Get(router.GetUserDetails).Handler(handler(serveGetUserDetails))
	m.Get(router.UpdateUserDetails).Handler(handler(serveUpdateUserDetails))
	m.Get(router.CreateToken).Handler(handler(serveCreateToken))
	m.Get(router.AckEmail).Handler(handler(serveAckEmail))
	m.Get(router.AddEmail).Handler(handler(serveAddEmail))
	m.Get(router.ListEmail).Handler(handler(serveListEmail))
	m.Get(router.GetEmail).Handler(handler(serveGetEmail))
	m.Get(router.DelEmail).Handler(handler(serveDelEmail))
	m.Get(router.PublicKey).Handler(handler(servePublicKey))
	m.Get(router.Service).Handler(handler(serveGetService))
	m.Get(router.AuthService).Handler(handler(serveAuthService))
	m.Get(router.OneTimeLogin).Handler(handler(serveOneTimeLogin))
	m.Get(router.APIDocs).Handler(handler(serveAPIDocs))
	return m
}

type handler func(http.ResponseWriter, *http.Request) error

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		switch err.(type) {
		case Error:
			WriteError(w, err.(Error))
			return
		default:
			log.Println("Error: unknow error occured:", err)
			WriteError(w, ErrServerError)
			return
		}
	}
}
