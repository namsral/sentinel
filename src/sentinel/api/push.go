// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"io/ioutil"
	"log"
	"mime"
	"net/http"

	"sentinel/push/apn"
	"sentinel/validate"
)

func serveSendPush(w http.ResponseWriter, r *http.Request) error {
	expectMediatype := "application/x-www-form-urlencoded"
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != expectMediatype {
		return ErrUnsupportedMediatype.Append("expected " + expectMediatype)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	// Get and validate form input
	// tokenStr := r.PostForm.Get("token") // used to validate the request
	serviceUIDStr := r.PostForm.Get("service_id")
	email := r.PostForm.Get("email")
	hash1 := r.PostForm.Get("hash1")

	// Validate form input
	if err := validate.Email(email); err != nil {
		return ErrInvalidEmail
	}
	if err := validate.NotEmpty(hash1); err != nil {
		e := ErrInvalidRequest.Append(`; hash1 parameter should not be empty`)
		return e
	}
	if err := validate.UUIDv4(serviceUIDStr); err != nil {
		e := ErrInvalidRequest.Append(`; service_id parameter should be a UUID verion 4`)
		return e
	}

	//TODO: finish implementation

	// claims, err := tokens.Verify(tokenStr, publicKey, &tokens.PushNotificationOptions)
	// if err != nil {
	// 	return ErrInvalidToken
	// }

	b, _ := ioutil.ReadAll(r.Body)
	log.Println("received sendPush request", r.RequestURI, string(b))

	n := &apn.PushNotification{
		AlertText: "Hello World!",
		Token:     "d2a84f4b8b650937ec8f73cd8be2c74add5a911ba64df27458ed8229da804a26",
	}
	_ = n

	return nil
}
