// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"log"
	"mime"
	"net/http"

	"sentinel/validate"

	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/mux"
)

func serveGetService(w http.ResponseWriter, r *http.Request) error {
	_, err := Authorized(r)
	if err != nil {
		return err
	}

	serviceUID := uuid.Parse(mux.Vars(r)["uid"])
	service, err := store.Services.Get(serviceUID)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, service)

	return nil
}

func serveAuthService(w http.ResponseWriter, r *http.Request) error {
	_, err := Authorized(r)
	if err != nil {
		return err
	}

	expectMediatype := "application/x-www-form-urlencoded"
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != expectMediatype {
		return ErrUnsupportedMediatype.Append("expected " + expectMediatype)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	// Get and validate form input
	email := r.PostForm.Get("email")
	status := r.PostForm.Get("status")
	enc1 := r.PostForm.Get("enc1")
	if err := validate.Email(email); err != nil {
		return ErrInvalidEmail
	}
	if err := validate.NotEmpty(status); err != nil {
		e := ErrInvalidRequest.Append(`status parameter should not be empty`)
		return e
	}
	if err := validate.NotEmpty(enc1); err != nil {
		e := ErrInvalidRequest.Append(`enc1 parameter should not be empty`)
		return e
	}

	//TODO: finish implementation
	// check if email is associated by authenticated user
	serviceUID := uuid.Parse(mux.Vars(r)["uid"])
	_, err = store.Services.Get(serviceUID)
	if err != nil {
		return err
	}

	log.Printf("serveAuthService called with params: id=%s, email:%s, status:%s, enc1: %s",
		serviceUID, email, status, enc1,
	)
	return nil
}
