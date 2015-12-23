// Copyright 2015 Lars Wiegman. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datastore

import (
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	d := NewDatastore(DB)

	expect := services[0]
	result, err := d.Services.Get(expect.UID)
	if err != nil {
		t.Error(err)
	}

	if reflect.DeepEqual(expect, result) {
		t.Fatalf("Result should have been %v, but it was %v", expect, result)
	}
}

func TestSetStatus(t *testing.T) {
	d := NewDatastore(DB)

	service := services[0]
	user := users[0]
	err := d.Services.Auth(service.UID, user.AuthEmailList[0].Email, "Accepted")
	if err != nil {
		t.Error(err)
	}
}
