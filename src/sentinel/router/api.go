// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package router

import (
	"net/url"

	"github.com/gorilla/mux"
)

func API(baseURL *url.URL) *mux.Router {
	m := mux.NewRouter()
	if baseURL != nil {
		m = m.Schemes(baseURL.Scheme).Host(baseURL.Host).PathPrefix(baseURL.Path).Subrouter()
	}
	// m.Path("/user/authorize").Methods("POST").Name(Signin)
	m.Path("/onetimelogin").Methods("POST").Name(OneTimeLogin)
	m.Path("/signup").Methods("POST").Name(Signup)
	// m.Path("/user/activity").Methods("GET").Name(GetActivity)
	// m.Path("/user/history").Methods("GET").Name(GetHistory)
	m.Path("/user/self").Methods("GET").Name(GetUserDetails)
	m.Path("/user/self").Methods("PUT").Name(UpdateUserDetails)
	m.Path("/email/{uid:.+}").Methods("GET").Name(GetEmail)
	m.Path("/email").Methods("GET").Name(ListEmail)
	m.Path("/email").Methods("POST").Name(AddEmail)
	m.Path("/email/{uid:.+}").Methods("DELETE").Name(DelEmail)
	m.Path("/verify").Methods("POST").Name(AckEmail)
	// m.Path("/user/service").Methods("POST").Name(AuthService)

	m.Path("/service/{uid:.+}").Methods("GET").Name(Service)
	m.Path("/service/{uid:.+}/auth").Methods("POST").Name(AuthService)

	m.Path("/token").Methods("POST").Name(CreateToken)
	m.Path("/pubkey").Methods("GET").Name(PublicKey)
	m.Path("/docs").Methods("GET").Name(APIDocs)
	return m
}
